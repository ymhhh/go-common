# go-common

## config

`config` 提供统一的配置读取与访问能力：

- 支持 JSON / YAML（`.yaml`/`.yml`）
- JSON 支持 `//` 与 `/* ... */` 注释（JSONC）
- 支持 `#include` 关键词加载其它配置并合并（后者覆盖前者）
- 支持 `${a.b.c}` 引用其它配置值、`${ENV}` 引用环境变量
- 支持通过 `Get/Set` 使用点路径访问与更新配置
- 支持 `GetOK(path)` 返回 `(Value, bool)` 判断配置是否存在
- 支持 `ToObject(path, &obj)` 将某个子树反序列化到自定义结构体（`path==""` 表示整棵配置）

### 示例

#### 1) 读取配置（JSONC/YAML）

```go
cfg, err := config.Load("conf/app.yaml")
if err != nil {
  panic(err)
}
```

#### 2) Get / GetOK / Set

```go
v := cfg.Get("a.b.c")
n, _ := v.Int()

vv, ok := cfg.GetOK("a.b.c") // ok=false 表示路径不存在
_ = vv

_ = cfg.Set("a.b.d", 123)
```

#### 3) 引用与环境变量

```yaml
a:
  b:
    c: 9
    d: ${a.b.c}
    e: ${ENV}
    mix: "x-${a.b.c}-${ENV}"
```

#### 4) include 合并（后者覆盖前者）

```yaml
#include base.yaml
server:
  port: 8080
```

也支持以 key 形式声明（JSON/YAML 均可）：

```yaml
"#include":
  - base.yaml
  - prod.yaml
```

#### 5) ToObject 反序列化到结构体

```go
type Server struct {
  Port int `json:"port"`
}
var s Server
if err := cfg.ToObject("server", &s); err != nil {
  panic(err)
}
```

