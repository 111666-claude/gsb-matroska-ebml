// Command ebml 是 EBML 解析的场景入口。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"example.com/matroska-ebml/ebml"
)

func emit(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func main() {
	sample := flag.String("sample", "", "vint|id|unknown|tree|crc|work")
	flag.Parse()

	for _, s := range ebml.Samples() {
		if s.Name == *sample {
			emit(s.Run())
			return
		}
	}
	if *sample == "work" {
		emit(ebml.Work(800))
		return
	}
	fmt.Fprintln(os.Stderr, "未知场景："+*sample)
	os.Exit(2)
}
