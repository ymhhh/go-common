# errcode

带 **namespace / id / 数字 code** 与 **上下文字段** 的结构化错误，支持多 cause（`errors.Join` / `errors.Is`）与便捷构造：

- `NewCode(...Option)`：主构造
- `New(message)` / `Newf(...)`：将普通错误包装为 `ErrorCode` 的便捷方法
- `ErrorCodeTmpl` + `NewTmpl`：按 `(namespace, code)` 注册模板，重复注册会 `panic`

**导入**：`github.com/ymhhh/go-common/errcode`
