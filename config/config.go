package config

import (
	"fmt"
	"path/filepath"
)

// Config is a loaded, mutable configuration tree.
// It is safe for single-goroutine usage; synchronize externally if needed.
type Config struct {
	root    map[string]any
	baseDir string
}

// Load reads a JSON/JSONC/YAML config file, processes #include, resolves ${...},
// and returns a Config.
func Load(path string) (*Config, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("config: abs path: %w", err)
	}

	root, err := loadFile(abs, map[string]struct{}{})
	if err != nil {
		return nil, err
	}

	c := &Config{
		root:    root,
		baseDir: filepath.Dir(abs),
	}

	if err := c.Resolve(); err != nil {
		return nil, err
	}
	return c, nil
}

// Get returns a typed Value for a dot path like "a.b.c".
func (c *Config) Get(path string) Value {
	v, _ := getPath(c.root, path)
	return Value{v: v}
}

// GetOK returns the Value and whether the path exists.
func (c *Config) GetOK(path string) (Value, bool) {
	v, ok := getPath(c.root, path)
	return Value{v: v}, ok
}

// Set updates a dot path like "a.b.c". Intermediate objects are created as maps.
func (c *Config) Set(path string, value any) error {
	return setPath(c.root, path, value)
}

// Resolve resolves ${ENV} and ${a.b.c} references in-place.
func (c *Config) Resolve() error {
	return resolveAll(c.root, c.lookupRef)
}

// ToObject deserializes config subtree at path into out (pointer).
// Path may be empty to deserialize the whole config.
func (c *Config) ToObject(path string, out any) error {
	var v any
	if path == "" {
		v = c.root
	} else {
		var ok bool
		v, ok = getPath(c.root, path)
		if !ok {
			return fmt.Errorf("config: path not found: %s", path)
		}
	}
	return decodeToObject(v, out)
}

func (c *Config) lookupRef(ref string) (any, bool) {
	// env first
	if v, ok := lookupEnv(ref); ok {
		return v, true
	}
	return getPath(c.root, ref)
}
