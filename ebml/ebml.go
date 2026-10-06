// Package ebml 解析 EBML 元素树。
package ebml

import (
	"encoding/binary"
	"errors"
	"fmt"
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
}

func readVint(data []byte, pos int, keepMarker bool, c *Counter) (uint64, int, error) {
	if pos >= len(data) {
		return 0, pos, errors.New("eof")
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
		return 0, pos, errors.New("eof")
	}
	value := uint64(b)
	for i := 1; i < length; i++ {
		c.Scanned++
		value = value<<8 | uint64(data[pos+i])
	}
	if !keepMarker {
		value &= (uint64(1) << uint(7*length)) - 1
	}
	return value, pos + length, nil
}

// Parse 解析元素流。
func Parse(data []byte, c *Counter) (*Doc, error) {
	doc := &Doc{}
	pos := 0
	for pos < len(data) {
		id, np, err := readVint(data, pos, true, c)
		if err != nil {
			break
		}
		size, np2, err := readVint(data, np, false, c)
		if err != nil || np2 > len(data) {
			break
		}
		declared := int64(size)
		el := &Element{
			ID:   id,
			Size: declared,
			Data: data[np2:],
			Path: fmt.Sprintf("%X", id),
		}
		doc.Roots = append(doc.Roots, el)
		pos = np2 + int(declared)
	}
	return doc, nil
}

// Find 按路径找元素。
func (d *Doc) Find(path string, c *Counter) *Element {
	for _, el := range d.Roots {
		c.Scanned++
		if el.Path == path {
			return el
		}
	}
	return nil
}

var _ = binary.BigEndian
