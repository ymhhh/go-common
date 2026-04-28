# go-common

## config

`config` 提供统一的配置读取与访问能力：

- 支持 JSON / YAML（`.yaml`/`.yml`）
- JSON 支持 `//` 与 `/* ... */` 注释（JSONC）
- 支持 `#include` 关键词加载其它配置并合并（后者覆盖前者）
- 支持 `${a.b.c}` 引用其它配置值、`${ENV}` 引用环境变量
- 支持通过 `Get/Set` 使用点路径访问与更新配置
- 支持 `ToObject(path, &obj)` 将某个子树反序列化到自定义结构体

