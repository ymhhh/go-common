package config

import (
	"math/big"
	"time"
)

// Config is a high-level, getter-oriented view of a configuration tree.
//
// The concrete type in this package is *Tree.
type Config interface {
	// Dot-path accessors used throughout this package.
	Get(path string) Value
	GetOK(path string) (Value, bool)
	Set(path string, value any) error
	Resolve() error

	// GetInterface get a object
	GetInterface(key string, defValue ...any) (res any)
	// GetString get a string
	GetString(key string, defValue ...string) (res string)
	// GetBoolean get a bool
	GetBoolean(key string, defValue ...bool) (b bool)
	// GetInt get a int
	GetInt(key string, defValue ...int) (res int)
	// GetFloat get a float
	GetFloat(key string, defValue ...float64) (res float64)
	// GetList get list of objects
	GetList(key string) (res []any)
	// GetStringList get list of strings
	GetStringList(key string) []string
	// GetBooleanList get list of bools
	GetBooleanList(key string) []bool
	// GetIntList get list of ints
	GetIntList(key string) []int
	// GetFloatList get list of float64s
	GetFloatList(key string) []float64
	// GetTimeDuration get time duration by (int)(uint), exp: 1s, 1day
	GetTimeDuration(key string, defValue ...time.Duration) time.Duration
	// GetByteSize get byte size by (int)(uint), exp: 1k, 1m
	GetByteSize(key string, defValue ...*big.Int) *big.Int
	// GetMap get map value
	GetMap(key string) Options
	// GetConfig get key's config
	GetConfig(key string) Config
	// ToObject unmarshal values to object
	// Deprecated: see function: Object
	ToObject(key string, model any) error
	// Object unmarshal values to object
	Object(model any, opts ...ObjOption) error
	// GetValuesConfig get key's values if values can be Config, or panic
	GetValuesConfig(key string) Config
	// SetKeyValue set key's value into config
	SetKeyValue(key string, value any) (err error)
	// Dump get all config
	Dump() (bs []byte, err error)
	// GetKeys get root keys
	GetRootKeys() []string
	// Copy deep copy configs
	Copy() Config
	IsEmpty() bool
}

// Options is a string-keyed map of configuration values.
type Options map[string]any

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
