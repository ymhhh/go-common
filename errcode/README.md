# errcode

带 **namespace / id / 数字 code** 与 **上下文字段** 的结构化错误，支持多 cause（`errors.Join` / `errors.Is`）与便捷构造：

- `NewCode(...Option)`：主构造
- `New(message)` / `Newf(...)`：将普通错误包装为 `ErrorCode` 的便捷方法
- `ErrorCodeTmpl` + `NewTmpl`：按 `(namespace, code)` 注册模板，重复注册会 `panic`

**导入**：`github.com/ymhhh/go-common/errcode`

### Error() 格式

`Error()` 字符串表示遵循以下规则，确保数字 code 始终可见：

| 条件 | 格式 |
|------|------|
| namespace + id + code(>0) + msg | `namespace.id[code]: msg` |
| namespace + id + msg（code=0） | `namespace.id: msg` |
| msg + code(>0)（无 namespace/id） | `err[code]: msg` |
| 仅 msg | `msg` |
| 仅 code | `errcode:code` |

示例：`user.create_failed[10001]: failed to create user`
