# tlsconfig

客户端 TLS 配置：根据 PEM 文件路径组合出 `crypto/tls.Config`（可选客户端证书、自定义 CA、`ServerName`、跳过校验等），支持通过 `yaml` / `json` 标签直接序列化到配置文件中。

若 `ca_path` 指定了 CA 证书且系统证书池不可用（例如在某些容器环境中），将回退到空证书池，仅信任显式提供的 CA 证书。

**导入**：`github.com/ymhhh/go-common/crypto/tlsconfig`
