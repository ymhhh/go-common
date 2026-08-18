package config

import (
	"math/big"
	"time"
)

// Config is a high-level, getter-oriented view of a configuration tree.
//
// The concrete type in this package is *Tree.
//
// Typed getters (GetString, GetInt, …) return the provided default when the
// path is missing or the value cannot be converted. Use GetOK plus Value
// methods when conversion errors must be distinguished from defaults.
type Config interface {
	Get(path string) Value
	GetOK(path string) (Value, bool)
	Set(path string, value any) error
	Resolve() error

	GetInterface(key string, defValue ...any) (res any)
	GetString(key string, defValue ...string) (res string)
	GetBoolean(key string, defValue ...bool) (b bool)
	GetInt(key string, defValue ...int) (res int)
	GetFloat(key string, defValue ...float64) (res float64)
	GetList(key string) (res []any)
	GetStringList(key string) []string
	GetBooleanList(key string) []bool
	GetIntList(key string) []int
	GetFloatList(key string) []float64
	// GetTimeDuration parses a duration. Numeric values are nanoseconds;
	// strings accept units such as "1s" or "1d".
	GetTimeDuration(key string, defValue ...time.Duration) time.Duration
	// GetByteSize parses a byte size. Numeric values are raw bytes;
	// strings accept units such as "1k" or "1m".
	GetByteSize(key string, defValue ...*big.Int) *big.Int
	GetMap(key string) Options
	GetConfig(key string) Config
	// ToObject unmarshals a subtree into model.
	//
	// Deprecated: use Object(model, WithObjectPath(key)).
	ToObject(key string, model any) error
	Object(model any, opts ...ObjOption) error
	// GetValuesConfig returns the subtree at key as a Config.
	//
	// Deprecated: use GetConfig. Missing keys and non-map values return an
	// empty Config instead of panicking.
	GetValuesConfig(key string) Config
	SetKeyValue(key string, value any) (err error)
	Dump() (bs []byte, err error)
	GetRootKeys() []string
	Copy() Config
	IsEmpty() bool
}

// Options is a string-keyed map of configuration values.
type Options map[string]any

// ToConfig Options to config
func (p *Options) ToConfig() Config {
	if p == nil {
		return &Tree{
			root:    map[string]any{},
			baseDir: "",
		}
	}
	return &Tree{
		root:    DeepCopy(map[string]any(*p)).(map[string]any),
		baseDir: "",
	}
}

// ObjOption configures Object unmarshalling.
type ObjOption func(*objectOpts)

type objectOpts struct {
	path string
}

// WithObjectPath selects the subtree path for Object. Empty path means the whole config.
func WithObjectPath(path string) ObjOption {
	return func(o *objectOpts) {
		o.path = path
	}
}

var _ Config = (*Tree)(nil)
