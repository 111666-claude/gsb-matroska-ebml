package ebml

import "testing"

func TestSingleRootCount(t *testing.T) {
	data := join(EncID(0x4286, 2), EncSize(0))
	doc, _ := Parse(data, &Counter{})
	if len(doc.Roots) != 1 {
		t.Fatalf("应有一个元素：%d", len(doc.Roots))
	}
}

func TestTwoRoots(t *testing.T) {
	one := join(EncID(0x4286, 2), EncSize(0))
	doc, _ := Parse(join(one, one), &Counter{})
	if len(doc.Roots) != 2 {
		t.Fatalf("应有两个元素：%d", len(doc.Roots))
	}
}

func TestDataPresent(t *testing.T) {
	data := join(EncID(0x4286, 2), EncSize(3), []byte{1, 2, 3})
	doc, _ := Parse(data, &Counter{})
	if len(doc.Roots[0].Data) == 0 {
		t.Fatal("元素应有数据")
	}
}

func TestFindMissing(t *testing.T) {
	data := join(EncID(0x4286, 2), EncSize(0))
	doc, _ := Parse(data, &Counter{})
	if doc.Find("FFFF", &Counter{}) != nil {
		t.Fatal("缺失路径应返回 nil")
	}
}

func TestParseEmpty(t *testing.T) {
	doc, _ := Parse(nil, &Counter{})
	if len(doc.Roots) != 0 {
		t.Fatalf("空输入应无元素：%d", len(doc.Roots))
	}
}

func TestCounterCounts(t *testing.T) {
	c := &Counter{}
	Parse(join(EncID(0x4286, 2), EncSize(0)), c)
	if c.Scanned == 0 {
		t.Fatal("应统计比较次数")
	}
}
