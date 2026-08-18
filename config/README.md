# config

统一的配置文件加载与访问：

- 支持 JSON / YAML（`.yaml`/`.yml`）
- JSON 支持 `//` 与 `/* ... */` 注释（JSONC）
- 支持 `#include` 合并多文件配置（后者覆盖前者）
- 支持 `${a.b.c}` 引用其它配置项、`${ENV}` 引用环境变量
- 支持 `Get` / `GetOK` / `Set` 点路径读写
- 类型化 `GetString` / `GetInt` 等在路径缺失或无法转换时返回 default；需要区分错误时用 `GetOK` + `Value`
- 支持 `Object(&obj, WithObjectPath(path))` 将子树反序列化到结构体（`path==""` 表示整棵配置）
- `GetMap` / `GetConfig` 返回深拷贝，外部修改不影响原始配置树
- `GetValuesConfig` 已废弃，请改用 `GetConfig`（缺 key / 非 map 返回空配置，不再 panic）
- 未定义的 `${REF}` 引用在 `Load` 时会返回错误

**导入**：`github.com/ymhhh/go-common/config`

### 示例

#### 读取配置（JSONC/YAML）

```go
cfg, err := config.Load("conf/app.yaml")
if err != nil {
  panic(err)
}
```

#### Get / GetOK / Set

```go
v := cfg.Get("a.b.c")
n, _ := v.Int()

vv, ok := cfg.GetOK("a.b.c") // ok=false 表示路径不存在
_ = vv

_ = cfg.Set("a.b.d", 123)
```

#### 引用与环境变量

```yaml
a:
  b:
    c: 9
    d: ${a.b.c}
    e: ${ENV}
    mix: "x-${a.b.c}-${ENV}"
```

> `${REF}` 在环境变量和配置树中都无法解析时，`Load` 会返回错误。

#### include 合并（后者覆盖前者）

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

#### Object 反序列化到结构体

```go
type Server struct {
  Port int `json:"port"`
}
var s Server
if err := cfg.Object(&s, WithObjectPath("server")); err != nil {
  panic(err)
}
```
