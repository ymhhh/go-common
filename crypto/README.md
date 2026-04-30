# crypto

与传输安全相关的子包集合（目前提供客户端 TLS 配置封装）。

| 子包 | 说明 |
| --- | --- |
| [tlsconfig](tlsconfig) | 从文件路径加载证书/私钥/CA，生成 `*tls.Config`；`Config` 支持 YAML/JSON 标签与 `yaml.Marshaler` / `yaml.Unmarshaler` |

**根导入**：本目录下请使用子包路径，例如 `github.com/ymhhh/go-common/crypto/tlsconfig`。
