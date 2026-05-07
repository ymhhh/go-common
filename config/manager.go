package config

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ymhhh/go-common/types"
)

// GetInterface returns the raw value at key, or the first default if missing.
func (c *Tree) GetInterface(key string, defValue ...any) (res any) {
	v, ok := getPath(c.root, key)
	if !ok {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return nil
	}
	return v
}

func (c *Tree) GetString(key string, defValue ...string) (res string) {
	val, ok := c.GetOK(key)
	if !ok {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return ""
	}
	s, err := val.String()
	if err != nil {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return ""
	}
	return s
}

func (c *Tree) GetBoolean(key string, defValue ...bool) (b bool) {
	val, ok := c.GetOK(key)
	if !ok {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return false
	}
	x, err := val.Bool()
	if err != nil {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return false
	}
	return x
}

func (c *Tree) GetInt(key string, defValue ...int) (res int) {
	val, ok := c.GetOK(key)
	if !ok {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return 0
	}
	x, err := val.Int()
	if err != nil {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return 0
	}
	return x
}

func (c *Tree) GetFloat(key string, defValue ...float64) (res float64) {
	val, ok := c.GetOK(key)
	if !ok {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return 0
	}
	x, err := val.Float64()
	if err != nil {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return 0
	}
	return x
}

func (c *Tree) GetList(key string) (res []any) {
	val, ok := c.GetOK(key)
	if !ok {
		return nil
	}
	sl, err := val.Slice()
	if err != nil {
		return nil
	}
	return sl
}

func (c *Tree) GetStringList(key string) []string {
	sl := c.GetList(key)
	if len(sl) == 0 {
		return nil
	}
	out := make([]string, 0, len(sl))
	for _, it := range sl {
		s, err := Value{v: it}.String()
		if err != nil {
			out = append(out, fmt.Sprint(it))
			continue
		}
		out = append(out, s)
	}
	return out
}

func (c *Tree) GetBooleanList(key string) []bool {
	sl := c.GetList(key)
	if len(sl) == 0 {
		return nil
	}
	out := make([]bool, 0, len(sl))
	for _, it := range sl {
		b, err := Value{v: it}.Bool()
		if err != nil {
			continue
		}
		out = append(out, b)
	}
	return out
}

func (c *Tree) GetIntList(key string) []int {
	sl := c.GetList(key)
	if len(sl) == 0 {
		return nil
	}
	out := make([]int, 0, len(sl))
	for _, it := range sl {
		n, err := Value{v: it}.Int()
		if err != nil {
			continue
		}
		out = append(out, n)
	}
	return out
}

func (c *Tree) GetFloatList(key string) []float64 {
	sl := c.GetList(key)
	if len(sl) == 0 {
		return nil
	}
	out := make([]float64, 0, len(sl))
	for _, it := range sl {
		f, err := Value{v: it}.Float64()
		if err != nil {
			continue
		}
		out = append(out, f)
	}
	return out
}

func (c *Tree) GetTimeDuration(key string, defValue ...time.Duration) time.Duration {
	val, ok := c.GetOK(key)
	if !ok {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return 0
	}

	// numeric -> treat as nanoseconds
	if i, err := val.Int(); err == nil {
		return time.Duration(i)
	}
	if f, err := val.Float64(); err == nil {
		return time.Duration(f)
	}

	s, err := val.String()
	if err != nil {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return 0
	}
	d := types.ParseStringTime(s)
	if d == 0 && s != "" && s != "0" {
		// allow stdlib duration strings like "300ms"
		if d2, err := time.ParseDuration(s); err == nil {
			return d2
		}
	}
	if d == 0 {
		if len(defValue) > 0 {
			return defValue[0]
		}
	}
	return d
}

func (c *Tree) GetByteSize(key string, defValue ...*big.Int) *big.Int {
	val, ok := c.GetOK(key)
	if !ok {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return nil
	}

	switch x := val.Any().(type) {
	case *big.Int:
		return x
	case int:
		return big.NewInt(int64(x))
	case int64:
		return big.NewInt(x)
	case uint64:
		return new(big.Int).SetUint64(x)
	case float64:
		return big.NewInt(int64(x))
	}

	s, err := val.String()
	if err != nil {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return nil
	}
	out := types.ParseStringByteSize(s, defValue...)
	if out == nil && len(defValue) > 0 {
		return defValue[0]
	}
	return out
}

func (c *Tree) GetMap(key string) Options {
	val, ok := c.GetOK(key)
	if !ok {
		return nil
	}
	m, err := val.Map()
	if err != nil {
		return nil
	}
	return Options(m)
}

func (c *Tree) GetConfig(key string) Config {
	val, ok := c.GetOK(key)
	if !ok {
		return &Tree{root: map[string]any{}, baseDir: c.baseDir}
	}
	m, err := Value{v: val.Any()}.Map()
	if err != nil {
		return &Tree{root: map[string]any{}, baseDir: c.baseDir}
	}
	return &Tree{root: m, baseDir: c.baseDir}
}

func (c *Tree) GetValuesConfig(key string) Config {
	v, ok := getPath(c.root, key)
	if !ok {
		panic(fmt.Sprintf("config: path not found: %s", key))
	}
	m, ok := v.(map[string]any)
	if !ok {
		panic(fmt.Sprintf("config: %q is not a map[string]any (got %T)", key, v))
	}
	return &Tree{root: m, baseDir: c.baseDir}
}

func (c *Tree) SetKeyValue(key string, value any) error {
	return c.Set(key, value)
}

func (c *Tree) Dump() ([]byte, error) {
	return json.MarshalIndent(c.root, "", "  ")
}

func (c *Tree) GetRootKeys() []string {
	if len(c.root) == 0 {
		return nil
	}
	keys := make([]string, 0, len(c.root))
	for k := range c.root {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (c *Tree) Copy() Config {
	return &Tree{
		root:    DeepCopy(c.root).(map[string]any),
		baseDir: c.baseDir,
	}
}

func (c *Tree) IsEmpty() bool {
	return len(c.root) == 0
}

// ToObject deserializes config subtree at key into model.
//
// Deprecated: use Object(model, WithObjectPath(key)).
func (c *Tree) ToObject(key string, model any) error {
	return c.decodeSubtree(key, model)
}

func (c *Tree) Object(model any, opts ...ObjOption) error {
	var oo objectOpts
	for _, opt := range opts {
		if opt != nil {
			opt(&oo)
		}
	}
	return c.decodeSubtree(oo.path, model)
}
