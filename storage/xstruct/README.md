# 结构体（`storage/xstruct`）

基于 `sync/atomic` 的 `atomic.Pointer` 实现的无锁并发结构：

- `Stack[T]`：Treiber stack（LIFO）
- `Queue[T]`：Michael–Scott queue（FIFO），使用 `NewQueue`

**导入**：`github.com/ymhhh/go-common/storage/xstruct`
