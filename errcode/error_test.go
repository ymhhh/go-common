package errcode

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	base := errors.New("db timeout")
	extra := errors.New("retry exhausted")
	msg := errors.New("failed to create user")

	e := NewCode(
		WithNamespace("user"),
		WithID("create_failed"),
		WithCode(10001),
		WithErrs(msg, base, nil, extra),
		WithContext(map[string]any{
			"user_id": 123,
		}),
	)

	if e.Namespace() != "user" {
		t.Fatalf("namespace: got %q", e.Namespace())
	}
	if e.ID() != "create_failed" {
		t.Fatalf("id: got %q", e.ID())
	}
	if e.Code() != 10001 {
		t.Fatalf("code: got %d", e.Code())
	}
	if e.Message() != "failed to create user" {
		t.Fatalf("message: got %q", e.Message())
	}
	if e.Error() != "user.create_failed: failed to create user" {
		t.Fatalf("error string: got %q", e.Error())
	}
	if !errors.Is(e, msg) || !errors.Is(e, base) || !errors.Is(e, extra) {
		t.Fatalf("errors.Is should match wrapped errors")
	}

	ctx := e.Context()
	ctx["user_id"] = 999
	if got := e.Context()["user_id"]; got != 123 {
		t.Fatalf("context should be copied, got %v", got)
	}
}

func TestFallbackErrorString(t *testing.T) {
	e := NewCode(WithCode(42))
	if e.Error() != "errcode:42" {
		t.Fatalf("fallback error string: got %q", e.Error())
	}
}

func TestSetGetContext(t *testing.T) {
	e := NewCode()

	if _, ok := e.GetContext("a"); ok {
		t.Fatalf("expected missing key")
	}

	e.SetContext("", 1)
	if _, ok := e.GetContext(""); ok {
		t.Fatalf("empty key should not be set")
	}

	e.SetContext("a", 1)
	if got, ok := e.GetContext("a"); !ok || got != 1 {
		t.Fatalf("get a: got=%v ok=%v", got, ok)
	}
}

func TestNew_Empty(t *testing.T) {
	e := New("")
	if e.Error() != "errcode:0" {
		t.Fatalf("empty New: got %q", e.Error())
	}
}

func TestNew_Message(t *testing.T) {
	e := New("hello")
	if e.Message() != "hello" {
		t.Fatalf("message: got %q", e.Message())
	}
	if e.Error() != "hello" {
		t.Fatalf("error string: got %q", e.Error())
	}
}

func TestNewf_WrapsFormattedCause(t *testing.T) {
	e := Newf("oops %d", 9)
	if e.Message() != "oops 9" {
		t.Fatalf("message: got %q", e.Message())
	}
	if e.Error() != "oops 9" {
		t.Fatalf("error string: got %q", e.Error())
	}
}

func TestErrorCodeTmpl_New(t *testing.T) {
	tmpl := NewTmpl("order", 20001, "create order failed")

	e := tmpl.New(WithID("create_failed"))
	if e.Namespace() != "order" {
		t.Fatalf("namespace: got %q", e.Namespace())
	}
	if e.Code() != 20001 {
		t.Fatalf("code: got %d", e.Code())
	}
	if e.Message() != "create order failed" {
		t.Fatalf("message: got %q", e.Message())
	}
	if e.ID() != "create_failed" {
		t.Fatalf("id: got %q", e.ID())
	}

	e2 := tmpl.New(WithCode(9), WithErrs(errors.New("override")))
	if e2.Code() != 9 || e2.Message() != "override" {
		t.Fatalf("override: code=%d message=%q", e2.Code(), e2.Message())
	}
}

func TestNewErrorCodeTmpl_DuplicatePanics(t *testing.T) {
	ns := "dup_" + t.Name()
	code := uint64(7788)

	_ = NewTmpl(ns, code, "a")
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on duplicate tmpl")
		}
	}()
	_ = NewTmpl(ns, code, "b")
}
