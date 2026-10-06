// Package ebml 解析 EBML 元素树。
package ebml

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

// Counter 记录比较次数。
type Counter struct {
	Scanned int
}

// Element 是一个元素。
type Element struct {
	ID       uint64
	Size     int64
	Data     []byte
	Children []*Element
	Path     string
}

// Doc 是解析结果。
type Doc struct {
	Roots []*Element
	index map[string]*Element
}

// masterIDs 是内含子元素的主元素 ID。
var masterIDs = map[uint64]bool{
	0x1A45DFA3: true, // EBML
	0x18538067: true, // Segment
	0x1654AE6B: true, // Tracks
	0x1549A966: true, // Info
	0x1F43B675: true, // Cluster
}

// crcElementID 是校验元素的 ID。
const crcElementID = 0xBF

var errCRC = errors.New("crc")

func readVint(data []byte, pos int, keepMarker bool, c *Counter) (uint64, int, int, error) {
	if pos >= len(data) {
		return 0, 0, pos, errors.New("eof")
	}
	c.Scanned++
	b := data[pos]
	length := 1
	mask := byte(0x80)
	for mask != 0 && b&mask == 0 {
		length++
		mask >>= 1
	}
	if mask == 0 || pos+length > len(data) {
		return 0, 0, pos, errors.New("eof")
	}
	value := uint64(b)
	for i := 1; i < length; i++ {
		c.Scanned++
		value = value<<8 | uint64(data[pos+i])
	}
	if !keepMarker {
		value &= (uint64(1) << uint(7*length)) - 1
	}
	return value, length, pos + length, nil
}

// Parse 解析元素流。
func Parse(data []byte, c *Counter) (*Doc, error) {
	doc := &Doc{index: make(map[string]*Element)}
	roots, err := parseRange(data, 0, len(data), "", doc, c)
	if err != nil {
		return nil, err
	}
	doc.Roots = roots
	return doc, nil
}

// parseRange 解析 [start, end) 内的元素，父路径为 parentPath。
func parseRange(data []byte, start, end int, parentPath string, doc *Doc, c *Counter) ([]*Element, error) {
	var out []*Element
	var ends []int
	pos := start
	for pos < end {
		id, _, np, err := readVint(data, pos, true, c)
		if err != nil {
			break
		}
		size, sizeLen, np2, err := readVint(data, np, false, c)
		if err != nil {
			break
		}
		declared := int64(size)
		var dataEnd int
		if size == (uint64(1)<<uint(7*sizeLen))-1 {
			// 未知长度：延伸到父元素末尾。
			declared = -1
			dataEnd = end
		} else {
			dataEnd = np2 + int(size)
			if dataEnd > end {
				break
			}
		}
		path := fmt.Sprintf("%X", id)
		if parentPath != "" {
			path = parentPath + "." + path
		}
		el := &Element{
			ID:   id,
			Size: declared,
			Data: data[np2:dataEnd],
			Path: path,
		}
		if masterIDs[id] && np2 < dataEnd {
			children, err := parseRange(data, np2, dataEnd, path, doc, c)
			if err != nil {
				return nil, err
			}
			el.Children = children
		}
		out = append(out, el)
		ends = append(ends, dataEnd)
		doc.index[path] = el
		pos = dataEnd
	}
	for i, el := range out {
		if el.ID != crcElementID {
			continue
		}
		if el.Size != 4 || binary.BigEndian.Uint32(el.Data) != crc32.ChecksumIEEE(data[ends[i]:end]) {
			return nil, errCRC
		}
	}
	return out, nil
}

// Find 按路径找元素。
func (d *Doc) Find(path string, c *Counter) *Element {
	c.Scanned++
	return d.index[path]
}
