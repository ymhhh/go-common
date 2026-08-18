package types

import (
	"math/big"
	"testing"
	"time"
)

func TestFindStringSubmatchMap(t *testing.T) {
	m, ok := FindStringSubmatchMap("10ms", timeRe)
	if !ok {
		t.Fatalf("expected match")
	}
	if m["value"] != "10" || m["unit"] != "ms" {
		t.Fatalf("unexpected captures: %+v", m)
	}

	_, ok = FindStringSubmatchMap("not-a-duration", timeRe)
	if ok {
		t.Fatalf("expected not match")
	}
}

func TestParseStringByteSize(t *testing.T) {
	def := big.NewInt(99)
	tests := []struct {
		name string
		in   string
		def  []*big.Int
		want *big.Int
	}{
		{name: "2kb", in: "2kb", want: new(big.Int).Mul(big.NewInt(2), _KByte)},
		{name: "3mib", in: "3mib", want: new(big.Int).Mul(big.NewInt(3), _MiByte)},
		{name: "1.5kb", in: "1.5kb", want: big.NewInt(1500)},
		{name: "1.5kib", in: "1.5kib", want: big.NewInt(1536)},
		{name: "10MB", in: "10MB", want: new(big.Int).Mul(big.NewInt(10), _MByte)},
		{name: "2KB", in: "2KB", want: new(big.Int).Mul(big.NewInt(2), _KByte)},
		{name: "1.5KiB", in: "1.5KiB", want: big.NewInt(1536)},
		{name: "bad with default", in: "bad", def: []*big.Int{def}, want: def},
		{name: "bad without default", in: "bad", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStringByteSize(tt.in, tt.def...)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("got %v, want nil", got)
				}
				return
			}
			if got == nil || got.Cmp(tt.want) != 0 {
				t.Fatalf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestParseStringTime(t *testing.T) {
	tests := []struct {
		name string
		in   string
		def  []time.Duration
		want time.Duration
	}{
		{name: "2s", in: "2s", want: 2 * time.Second},
		{name: "2S", in: "2S", want: 2 * time.Second},
		{name: "1H", in: "1H", want: time.Hour},
		{name: "3m", in: "3m", want: 3 * time.Minute},
		{name: "4h", in: "4h", want: 4 * time.Hour},
		{name: "1.5s", in: "1.5s", want: 1500 * time.Millisecond},
		{name: "0.5d", in: "0.5d", want: 12 * time.Hour},
		{name: "bad with default", in: "bad", def: []time.Duration{7 * time.Second}, want: 7 * time.Second},
		{name: "bad without default", in: "bad", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStringTime(tt.in, tt.def...)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseStringTimeOK_ZeroDurations(t *testing.T) {
	for _, s := range []string{"0s", "0d", "0w", "0y", "0h", "0m", "0ms"} {
		t.Run(s, func(t *testing.T) {
			got, ok := ParseStringTimeOK(s)
			if !ok {
				t.Fatalf("expected ok")
			}
			if got != 0 {
				t.Fatalf("got %v, want 0", got)
			}
		})
	}
	if _, ok := ParseStringTimeOK("bad"); ok {
		t.Fatalf("bad: expected !ok")
	}
	if _, ok := ParseStringTimeOK("0"); ok {
		t.Fatalf("bare 0 has no unit: expected !ok")
	}
}
