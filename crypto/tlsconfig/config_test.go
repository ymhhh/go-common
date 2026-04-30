package tlsconfig

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestGetTLSConfig_ClientCertAndCA(t *testing.T) {
	dir := t.TempDir()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ca key: %v", err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-ca"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:         true,
		BasicConstraintsValid: true,
	}
	caDer, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create ca: %v", err)
	}
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDer})
	caPath := filepath.Join(dir, "ca.pem")
	writeFile(t, caPath, string(caPEM))

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("leaf key: %v", err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "example.com"},
		DNSNames:     []string{"example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	caCert, err := x509.ParseCertificate(caDer)
	if err != nil {
		t.Fatalf("parse ca: %v", err)
	}
	leafDer, err := x509.CreateCertificate(rand.Reader, leafTmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create leaf: %v", err)
	}
	leafPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafDer})
	certPath := filepath.Join(dir, "cert.pem")
	writeFile(t, certPath, string(leafPEM))

	keyDer, err := x509.MarshalECPrivateKey(leafKey)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDer})
	keyPath := filepath.Join(dir, "key.pem")
	writeFile(t, keyPath, string(keyPEM))

	cfg := Config{
		CertPath:   certPath,
		KeyPath:    keyPath,
		CAPath:     caPath,
		ServerName: "example.com",
	}
	tlsCfg, err := cfg.GetTLSConfig()
	if err != nil {
		t.Fatalf("GetTLSConfig: %v", err)
	}
	if tlsCfg == nil {
		t.Fatalf("tls cfg nil")
	}
	if len(tlsCfg.Certificates) != 1 {
		t.Fatalf("expected 1 client cert, got %d", len(tlsCfg.Certificates))
	}
	if tlsCfg.RootCAs == nil {
		t.Fatalf("expected RootCAs")
	}
}

func TestGetTLSConfig_PartialClientCertPaths(t *testing.T) {
	_, err := (Config{CertPath: "a.pem"}).GetTLSConfig()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestConfig_UnmarshalYAML(t *testing.T) {
	const in = `
cert_path: /tmp/cert.pem
key_path: /tmp/key.pem
ca_path: /tmp/ca.pem
server_name: example.com
insecure_skip_verify: true
`
	var c Config
	if err := yaml.Unmarshal([]byte(in), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if c.CertPath != "/tmp/cert.pem" || c.KeyPath != "/tmp/key.pem" || c.CAPath != "/tmp/ca.pem" {
		t.Fatalf("paths: %+v", c)
	}
	if c.ServerName != "example.com" || !c.InsecureSkipVerify {
		t.Fatalf("tls opts: %+v", c)
	}
}

func TestConfig_MarshalYAML_RoundTrip(t *testing.T) {
	src := Config{
		CertPath:           "/a/cert.pem",
		KeyPath:            "/a/key.pem",
		CAPath:             "/a/ca.pem",
		ServerName:         "svc.local",
		InsecureSkipVerify: true,
	}
	out, err := yaml.Marshal(&src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var dst Config
	if err := yaml.Unmarshal(out, &dst); err != nil {
		t.Fatalf("unmarshal roundtrip: %v", err)
	}
	if dst != src {
		t.Fatalf("roundtrip: got %+v want %+v", dst, src)
	}
}
