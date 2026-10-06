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

// 主元素 ID：EBML / Segment / Tracks / Info / Cluster。
var masterIDs = map[uint64]bool{
	0x1A45DFA3: true,
	0x18538067: true,
	0x1654AE6B: true,
	0x1549A966: true,
	0x1F43B675: true,
}

// crc32ID 是 CRC-32 校验元素的 ID。
const crc32ID = 0xBF

// readVint 读一个变长整数，返回去掉/保留标记位的值与占用字节数。
func readVint(data []byte, pos, end int, keepMarker bool, c *Counter) (uint64, int, error) {
	if pos >= end {
		return 0, 0, errors.New("eof")
	}
	c.Scanned++
	b := data[pos]
	length := 1
	mask := byte(0x80)
	for mask != 0 && b&mask == 0 {
		length++
		mask >>= 1
	}
	if mask == 0 || pos+length > end {
		return 0, 0, errors.New("eof")
	}
	value := uint64(b)
	for i := 1; i < length; i++ {
		c.Scanned++
		value = value<<8 | uint64(data[pos+i])
	}
	if !keepMarker {
		value &= (uint64(1) << uint(7*length)) - 1
	}
	return value, length, nil
}

// isUnknownSize 判断长度值是否为未知长度（有效位全为 1）。
func isUnknownSize(value uint64, length int) bool {
	return value == (uint64(1)<<uint(7*length))-1
}

func joinPath(parent string, id uint64) string {
	h := fmt.Sprintf("%X", id)
	if parent == "" {
		return h
	}
	return parent + "." + h
}

type parsedEl struct {
	el        *Element
	bodyStart int
	bodyEnd   int
}

// parseElements 在 [start, end) 范围内解析兄弟元素。
func parseElements(data []byte, start, end int, parentPath string, doc *Doc, c *Counter) ([]*Element, error) {
	var parsed []parsedEl
	pos := start
	for pos < end {
		id, idLen, err := readVint(data, pos, end, true, c)
		if err != nil {
			break
		}
		size, sizeLen, err := readVint(data, pos+idLen, end, false, c)
		if err != nil {
			break
		}
		bodyStart := pos + idLen + sizeLen
		unknown := isUnknownSize(size, sizeLen)
		var bodyEnd int
		var declared int64
		if unknown {
			bodyEnd = end
			declared = -1
		} else {
			bodyEnd = bodyStart + int(size)
			declared = int64(size)
			if bodyEnd > end {
				break
			}
		}

		el := &Element{
			ID:   id,
			Size: declared,
			Data: data[bodyStart:bodyEnd],
			Path: joinPath(parentPath, id),
		}
		if masterIDs[id] {
			children, err := parseElements(data, bodyStart, bodyEnd, el.Path, doc, c)
			if err != nil {
				return nil, err
			}
			el.Children = children
		}

		parsed = append(parsed, parsedEl{el: el, bodyStart: bodyStart, bodyEnd: bodyEnd})
		if _, ok := doc.index[el.Path]; !ok {
			doc.index[el.Path] = el
		}

		if unknown {
			break
		}
		pos = bodyEnd
	}

	els := make([]*Element, len(parsed))
	for i, p := range parsed {
		els[i] = p.el
	}
	if err := verifyCRC(data, end, parsed); err != nil {
		return els, err
	}
	return els, nil
}

// verifyCRC 校验本层内每个 CRC-32 元素：其值是该元素之后到父元素末尾字节的 CRC-32。
func verifyCRC(data []byte, end int, parsed []parsedEl) error {
	for _, p := range parsed {
		if p.el.ID != crc32ID {
			continue
		}
		if len(p.el.Data) != 4 {
			return errors.New("crc")
		}
		expected := binary.BigEndian.Uint32(p.el.Data)
		actual := crc32.ChecksumIEEE(data[p.bodyEnd:end])
		if expected != actual {
			return errors.New("crc")
		}
	}
	return nil
}

// Parse 解析元素流。
func Parse(data []byte, c *Counter) (*Doc, error) {
	doc := &Doc{index: map[string]*Element{}}
	roots, err := parseElements(data, 0, len(data), "", doc, c)
	doc.Roots = roots
	return doc, err
}

// Find 按路径走索引找元素。
func (d *Doc) Find(path string, c *Counter) *Element {
	c.Scanned++
	return d.index[path]
}
