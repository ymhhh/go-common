# tlsconfig

客户端 TLS 配置：根据 PEM 文件路径组合出 `crypto/tls.Config`（可选客户端证书、自定义 CA、`ServerName`、跳过校验等），并支持序列化到配置（`yaml`/`json` 标签 + `yaml.Marshaler` / `yaml.Unmarshaler`）。

**导入**：`github.com/ymhhh/go-common/crypto/tlsconfig`
