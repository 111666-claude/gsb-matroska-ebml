# matroska-ebml

解析 EBML 元素树：变长整数、未知长度、主元素递归、CRC-32 校验，并按路径建索引。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/ebml --sample=vint
go run ./cmd/ebml --sample=id
go run ./cmd/ebml --sample=unknown
go run ./cmd/ebml --sample=tree
go run ./cmd/ebml --sample=crc
go run ./cmd/ebml --sample=work
```

## 口径

- **变长整数**：首字节最高位起连续 0 的个数加一就是字节数；ID 保留标记位，长度去掉标记位，值由全部字节拼成。
- **未知长度**：长度位的有效位全为 1 表示未知，数据延伸到父元素末尾（记为 -1）。
- **主元素**：`EBML` / `Segment` / `Tracks` / `Info` / `Cluster` 这几种 ID 的元素内含子元素，子元素只能在父元素声明的范围内解析。
- **CRC-32**：ID 为 `BF` 的元素是校验元素，其值是父元素内该元素之后全部字节的 CRC-32（大端 4 字节）；对不上报 `crc`。
- **路径**：元素路径是父路径加自身 ID 的十六进制串，点号分隔；`Find(path)` 走索引。

## 不变量

- 元素数据长度等于声明长度（未知长度延伸到父元素末尾）。
- 子元素的字节范围不超过父元素声明的范围。
- `scanned` 不随元素数乘查询次数放大：八百个元素查八百次的 `scanned` 不超过 6000。

## 输出契约

`Element` 含 `ID` / `Size` / `Data` / `Children` / `Path`；`Find(path)` 按路径取元素。
场景打印一行 JSON，校验失败打印 `{"error":"crc"}`。`--sample=work` 打印 `{"elements": N, "scanned": N}`。
