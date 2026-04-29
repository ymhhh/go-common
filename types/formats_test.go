package types

import (
	"math/big"
	"testing"
	"time"
)

func TestFindStringSubmatchMap(t *testing.T) {
	m, ok := FindStringSubmatchMap("10ms", timeReg)
	if !ok {
		t.Fatalf("expected match")
	}
	if m["value"] != "10" || m["unit"] != "ms" {
		t.Fatalf("unexpected captures: %+v", m)
	}

	_, ok = FindStringSubmatchMap("not-a-duration", timeReg)
	if ok {
		t.Fatalf("expected not match")
	}
}

func TestParseStringByteSize(t *testing.T) {
	// decimal
	got := ParseStringByteSize("2kb")
	want := new(big.Int).Mul(big.NewInt(2), _KByte)
	if got == nil || got.Cmp(want) != 0 {
		t.Fatalf("2kb: got=%v want=%v", got, want)
	}

	// binary
	got = ParseStringByteSize("3mib")
	want = new(big.Int).Mul(big.NewInt(3), _MiByte)
	if got == nil || got.Cmp(want) != 0 {
		t.Fatalf("3mib: got=%v want=%v", got, want)
	}

	// default when not matched
	def := big.NewInt(99)
	got = ParseStringByteSize("bad", def)
	if got == nil || got.Cmp(def) != 0 {
		t.Fatalf("default: got=%v want=%v", got, def)
	}

	// nil when not matched and no default
	if got := ParseStringByteSize("bad"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestParseStringTime(t *testing.T) {
	if got := ParseStringTime("2s"); got != 2*time.Second {
		t.Fatalf("2s: got %v", got)
	}
	if got := ParseStringTime("3m"); got != 3*time.Minute {
		t.Fatalf("3m: got %v", got)
	}
	if got := ParseStringTime("4h"); got != 4*time.Hour {
		t.Fatalf("4h: got %v", got)
	}

	// default when not matched
	if got := ParseStringTime("bad", 7*time.Second); got != 7*time.Second {
		t.Fatalf("default: got %v", got)
	}

	// zero when not matched and no default
	if got := ParseStringTime("bad"); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}
