# logger

基于 [`sirupsen/logrus`](https://github.com/sirupsen/logrus) 的轻量封装，配合 `go-common/config` 实现配置化管理。

**导入**：`github.com/ymhhh/go-common/logger`

### 配置示例

```yaml
logger:
  level: info
  format: json         # text|json
  output: stdout       # stdout|stderr|discard|/path/to/app.log|file:/path/to/app.log
  reportCaller: false
  file:
    path: ./app.log
    rotate:
      enabled: true
      maxSizeMB: 100
      maxBackups: 7
      maxAgeDays: 7
      compress: false
      localTime: false
  text:
    disableColors: true
    fullTimestamp: true
  json:
    prettyPrint: false
```

> `output` 为空时自动选用 `file.path`；若两者都未设置则默认输出到 `stderr`。相对路径相对于配置文件所在目录解析，无法解析时返回错误。

### 使用示例

```go
cfg, _ := config.Load("conf/app.yaml")

// 全局 logger
_ = logger.InitGlobal(cfg) // default subtree: "logger"
logger.L().WithField("module", "main").Info("started")

// 独立实例
l, _ := logger.FromConfig(cfg) // default subtree: "logger"
defer l.Close()
l.WithField("module", "worker").Warn("something happened")

// 带 trace_id/span_id 的 context-aware 日志
logger.LFromContext(ctx).Info("request processed")
```
