# types

通用类型与小型工具函数集合，便于在配置、命令行 flag、YAML 之间复用同一套表示：

- **时间/大小/字符串解析**：从人类可读字符串解析时间间隔（`ParseStringTime`）、字节大小（`ParseStringByteSize`），正则预编译避免重复开销
- **标量包装类型**：
  - `Duration` — flag.Value + YAML 序列化
  - `Found` — 金额/浮点展示（两位小数）
  - `Secret` — 日志/序列化脱敏（`String()` 返回 `<hidden>`）
  - `Strings` — 多值 flag 与 YAML 列表
- **泛型切片工具**：`Contains`、`ContainsFunc`、`Filter`、`Map`、`Reduce`、`Unique`、`Chunk`、`Shuffle`（并发安全）、`Reverse`、`Intersect`、`Union`、`Difference`、`Any`、`All`、`First`、`Count`、`Distinct`、`Flatten`、`Partition`、`Zip`、`Unzip`、`Sort`、`Sum`、`Max`、`Min`、`Average`、`Take`、`Drop`、`Convert`
- **类型转换**：`ToInt`、`ToInt64`、`ToFloat64`、`RoundFund`（使用 `math.Round` 正确处理正负数）

**导入**：`github.com/ymhhh/go-common/types`
