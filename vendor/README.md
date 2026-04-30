# vendor

本目录为 Go modules **vendor** 目录（由 `go mod vendor` 生成），包含第三方依赖源码副本，用于可重复构建与离线编译。

一般情况下 **不要手工修改** `vendor` 中的文件；更新依赖请通过模块版本管理与重新执行 `go mod vendor`。
