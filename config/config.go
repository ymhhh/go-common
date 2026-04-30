package config

import (
	"fmt"
	"path/filepath"
)

// Tree is a loaded, mutable configuration tree.
// It is safe for single-goroutine usage; synchronize externally if needed.
type Tree struct {
	root    map[string]any
	baseDir string
}

// Load reads a JSON/JSONC/YAML config file, processes #include, resolves ${...},
// and returns a Config.
func Load(path string) (Config, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("config: abs path: %w", err)
	}

	root, err := loadFile(abs, map[string]struct{}{})
	if err != nil {
		return nil, err
	}

	c := &Tree{
		root:    root,
		baseDir: filepath.Dir(abs),
	}

	if err := c.Resolve(); err != nil {
		return nil, err
	}
	return c, nil
}

// Get returns a typed Value for a dot path like "a.b.c".
func (c *Tree) Get(path string) Value {
	v, _ := getPath(c.root, path)
	return Value{v: v}
}

// GetOK returns the Value and whether the path exists.
func (c *Tree) GetOK(path string) (Value, bool) {
	v, ok := getPath(c.root, path)
	return Value{v: v}, ok
}

// Set updates a dot path like "a.b.c". Intermediate objects are created as maps.
func (c *Tree) Set(path string, value any) error {
	return setPath(c.root, path, value)
}

// Resolve resolves ${ENV} and ${a.b.c} references in-place.
func (c *Tree) Resolve() error {
	return resolveAll(c.root, c.lookupRef)
}

func (c *Tree) decodeSubtree(path string, out any) error {
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

func (c *Tree) lookupRef(ref string) (any, bool) {
	// env first
	if v, ok := lookupEnv(ref); ok {
		return v, true
	}
	return getPath(c.root, ref)
}
