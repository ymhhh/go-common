# types

通用类型与小型工具函数集合，便于在配置、命令行 flag、YAML 之间复用同一套表示：

- **时间/大小/字符串解析**（`formats` 等）：从人类可读字符串解析时间间隔、字节大小等
- **标量包装类型**：如 `Duration`（`flag.Value` + YAML）、`Found`（金额/浮点展示）、`Secret`（日志/序列化脱敏）、`Strings`（多值 flag 与 YAML 列表）
- **泛型切片工具**（`types_slice`）：`Contains`、`Filter`、`Unique`、`Chunk` 等

**导入**：`github.com/ymhhh/go-common/types`
