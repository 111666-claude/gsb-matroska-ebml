package ebml

import "fmt"

// Scenario 是一个固定场景。
type Scenario struct {
	Name string
	Run  func() map[string]any
}

// EncID 按给定字节数编码一个元素 ID。
func EncID(id uint64, length int) []byte {
	out := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		out[i] = byte(id)
		id >>= 8
	}
	return out
}

// EncSize 按最短字节数编码一个长度值。
func EncSize(value uint64) []byte {
	for length := 1; length <= 8; length++ {
		if value < uint64(1)<<uint(7*length) {
			out := make([]byte, length)
			v := value
			for i := length - 1; i >= 0; i-- {
				out[i] = byte(v)
				v >>= 8
			}
			out[0] |= 0x80 >> uint(length-1)
			return out
		}
	}
	return nil
}

// EncUnknownSize 编码未知长度。
func EncUnknownSize() []byte {
	out := make([]byte, 8)
	out[0] = 0x01
	for i := 1; i < 8; i++ {
		out[i] = 0xff
	}
	return out
}

func join(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// Samples 返回全部场景。
func Samples() []Scenario {
	return []Scenario{
		{Name: "unknown", Run: func() map[string]any {
			data := join(EncID(0x18538067, 4), EncUnknownSize(), []byte{1, 2, 3})
			doc, _ := Parse(data, &Counter{})
			return map[string]any{"size": doc.Roots[0].Size}
		}},
		{Name: "tree", Run: func() map[string]any {
			child1 := join(EncID(0x4286, 2), EncSize(0))
			child2 := join(EncID(0x42f7, 2), EncSize(0))
			kids := join(child1, child2)
			data := join(EncID(0x18538067, 4), EncSize(uint64(len(kids))), kids)
			doc, _ := Parse(data, &Counter{})
			el := doc.Find("18538067", &Counter{})
			if el == nil {
				return map[string]any{"children": 0}
			}
			return map[string]any{"children": len(el.Children)}
		}},
		{Name: "crc", Run: func() map[string]any {
			crc := join([]byte{0xbf}, EncSize(4), make([]byte, 4))
			leaf := join(EncID(0x4286, 2), EncSize(0))
			body := join(crc, leaf)
			data := join(EncID(0x18538067, 4), EncSize(uint64(len(body))), body)
			_, err := Parse(data, &Counter{})
			if err != nil {
				return map[string]any{"error": err.Error()}
			}
			return map[string]any{"ok": true}
		}},
		{Name: "bound", Run: func() map[string]any {
			data := join(EncID(0x4286, 2), EncSize(8), make([]byte, 8), make([]byte, 20))
			doc, _ := Parse(data, &Counter{})
			return map[string]any{"datalen": len(doc.Roots[0].Data)}
		}},
	}
}

// Work 跑规模线场景。
func Work(n int) map[string]any {
	var data []byte
	for i := 0; i < n; i++ {
		data = join(data, EncID(uint64(0x4000+i), 2), EncSize(0))
	}
	doc, _ := Parse(data, &Counter{})
	c := &Counter{}
	for i := 0; i < n; i++ {
		doc.Find(fmt.Sprintf("%X", 0x4000+i), c)
	}
	return map[string]any{"elements": n, "scanned": c.Scanned}
}
