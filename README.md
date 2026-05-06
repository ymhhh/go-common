# go-common

Go 业务无关的公共库与工具集，按子目录拆分为独立包。各包说明见对应目录下的 `README.md`。

| 目录 | 说明 |
| --- | --- |
| [builder](builder) | 启动时打印构建/运行环境信息 |
| [config](config) | JSONC/YAML 配置加载、合并、引用与点路径访问 |
| [crypto](crypto) | TLS 等加密封装（如 `tlsconfig`） |
| [errcode](errcode) | 带 code 与上下文的结构化错误 |
| [logger](logger) | 基于 logrus 的配置化日志封装 |
| [storage](storage) | 并发数据结构（如无锁 Stack/Queue，见 [`storage/xstruct`](storage/xstruct)） |
| [types](types) | 通用标量类型、flag/YAML 适配、泛型切片工具等 |

模块路径：`github.com/ymhhh/go-common`
