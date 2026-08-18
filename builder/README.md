# builder

启动时展示构建与运行环境信息的工具包（基于 `dimiro1/banner`），用于在 CLI/服务进程启动阶段打印程序名、版本、分支、提交、编译器、构建时间、作者，以及当前 Go 运行时与操作系统等信息。

`Show()` 默认打印；横幅标题使用 `ProgramName`（为空时回退为 `GO COMMON BUILDER`）。不需要输出时传入 `OffShow()`。

**导入**：`github.com/ymhhh/go-common/builder`
