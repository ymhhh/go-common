package errcode

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
)

// ErrorCode represents an error with a unique code and context.
type ErrorCode interface {
	error

	Namespace() string
	ID() string
	Code() uint64
	Message() string

	Context() map[string]any
	SetContext(key string, val any)
	GetContext(key string) (any, bool)

	Causes() []error
	Unwrap() error
}

// ErrorCodeTmpl struct for error code template.
type ErrorCodeTmpl struct {
	namespace string
	code      uint64
	message   string
}

var tmplRegistry = struct {
	mu sync.Mutex
	m  map[string]struct{}
}{
	m: make(map[string]struct{}),
}

func tmplKey(namespace string, code uint64) string {
	return namespace + "\x00" + strconv.FormatUint(code, 10)
}

// NewTmpl creates an ErrorCodeTmpl and registers (namespace, code).
// If the same (namespace, code) is registered before, it panics.
func NewTmpl(namespace string, code uint64, message string) *ErrorCodeTmpl {
	k := tmplKey(namespace, code)

	tmplRegistry.mu.Lock()
	defer tmplRegistry.mu.Unlock()

	if _, ok := tmplRegistry.m[k]; ok {
		panic(fmt.Sprintf("errcode: duplicate ErrorCodeTmpl for namespace=%q code=%d", namespace, code))
	}
	tmplRegistry.m[k] = struct{}{}

	return &ErrorCodeTmpl{
		namespace: namespace,
		code:      code,
		message:   message,
	}
}

func (t ErrorCodeTmpl) New(opts ...Option) ErrorCode {
	var probe ErrorOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&probe)
		}
	}

	base := []Option{
		WithNamespace(t.namespace),
		WithCode(t.code),
	}
	if len(probe.errs) == 0 && t.message != "" {
		base = append(base, WithErrs(errors.New(t.message)))
	}
	opts = append(base, opts...)
	return New(opts...)
}

// ErrorOptions describes attributes used to build an Error.
type ErrorOptions struct {
	namespace string
	id        string
	code      uint64

	errs []error

	ctx map[string]any
}

// Option mutates ErrorOptions.
type Option func(*ErrorOptions)

func WithNamespace(namespace string) Option {
	return func(o *ErrorOptions) { o.namespace = namespace }
}

func WithID(id string) Option {
	return func(o *ErrorOptions) { o.id = id }
}

func WithCode(code uint64) Option {
	return func(o *ErrorOptions) { o.code = code }
}

func WithErrs(errs ...error) Option {
	return func(o *ErrorOptions) {
		for _, err := range errs {
			if err != nil {
				o.errs = append(o.errs, err)
			}
		}
	}
}

func WithContext(ctx map[string]any) Option {
	return func(o *ErrorOptions) {
		if len(ctx) == 0 {
			return
		}
		if o.ctx == nil {
			o.ctx = make(map[string]any, len(ctx))
		}
		for k, v := range ctx {
			o.ctx[k] = v
		}
	}
}

// Error is a custom structured error with code and context.
type Error struct {
	namespace string
	id        string
	code      uint64
	errs      []error
	ctx       map[string]any
}

func New(opts ...Option) ErrorCode {
	var o ErrorOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	e := &Error{
		namespace: o.namespace,
		id:        o.id,
		code:      o.code,
		errs:      append([]error(nil), o.errs...),
	}
	if len(o.ctx) > 0 {
		e.ctx = make(map[string]any, len(o.ctx))
		for k, v := range o.ctx {
			e.ctx[k] = v
		}
	}
	return e
}

func (e *Error) Namespace() string { return e.namespace }

func (e *Error) ID() string { return e.id }

func (e *Error) Code() uint64 { return e.code }

func (e *Error) Message() string {
	if len(e.errs) > 0 && e.errs[0] != nil {
		return e.errs[0].Error()
	}
	return ""
}

func (e *Error) Context() map[string]any {
	if len(e.ctx) == 0 {
		return nil
	}
	out := make(map[string]any, len(e.ctx))
	for k, v := range e.ctx {
		out[k] = v
	}
	return out
}

func (e *Error) SetContext(key string, val any) {
	if key == "" {
		return
	}
	if e.ctx == nil {
		e.ctx = make(map[string]any, 1)
	}
	e.ctx[key] = val
}

func (e *Error) GetContext(key string) (any, bool) {
	if key == "" || len(e.ctx) == 0 {
		return nil, false
	}
	v, ok := e.ctx[key]
	return v, ok
}

func (e *Error) Causes() []error {
	if len(e.errs) == 0 {
		return nil
	}
	return append([]error(nil), e.errs...)
}

func (e *Error) Error() string {
	if msg := e.Message(); msg != "" {
		if e.namespace != "" && e.id != "" {
			return fmt.Sprintf("%s.%s: %s", e.namespace, e.id, msg)
		}
		return msg
	}

	switch {
	case e.namespace != "" && e.id != "":
		return fmt.Sprintf("%s.%s", e.namespace, e.id)
	default:
		return fmt.Sprintf("errcode:%d", e.code)
	}
}

func (e *Error) Unwrap() error {
	switch len(e.errs) {
	case 0:
		return nil
	case 1:
		return e.errs[0]
	default:
		return errors.Join(e.errs...)
	}
}
