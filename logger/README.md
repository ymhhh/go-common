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
      localTime: false   # 整数后缀模式下无效，保留仅为兼容旧配置
  text:
    disableColors: true
    fullTimestamp: true
  json:
    prettyPrint: false
```

> `output` 为空时自动选用 `file.path`；若两者都未设置则默认输出到 `stderr`。相对路径相对于配置文件所在目录解析，无法解析时返回错误。

### 日志滚动

滚动由 [`golift.io/rotatorr`](https://github.com/golift/rotatorr) 配合自定义后缀 rotator 实现：

- **主日志路径**（`output` / `file.path` 解析后的路径）始终是**真实文件**，不创建软链接
- **备份命名**：`app.log.1`、`app.log.2`（启用 `compress` 后为 `app.log.1.gz`）
- **logbook 采集**：请直接采集主日志真实路径（如 `/var/log/app.log`），勿依赖 symlink

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
