package types

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

type namedInt int

func TestToInt64(t *testing.T) {
	tests := []struct {
		name    string
		in      any
		want    int64
		wantErr string
	}{
		{name: "nil", in: nil, want: 0},
		{name: "int64", in: int64(12), want: 12},
		{name: "int32", in: int32(12), want: 12},
		{name: "int", in: int(12), want: 12},
		{name: "string", in: "34", want: 34},
		{name: "json.Number", in: json.Number("56"), want: 56},
		{name: "json.Number exponent", in: json.Number("1e6"), want: 1_000_000},
		{name: "named int", in: namedInt(9), want: 9},
		{name: "json.Number fractional", in: json.Number("1.5"), wantErr: "cannot convert"},
		{name: "uint64 overflow", in: uint64(math.MaxInt64) + 1, wantErr: "cannot convert"},
		{name: "integral float64", in: float64(78), want: 78},
		{name: "fractional float64", in: 12.5, wantErr: "cannot convert"},
		{name: "inf float64", in: math.Inf(1), wantErr: "cannot convert"},
		{name: "maxint64 float64", in: float64(math.MaxInt64), wantErr: "cannot convert"},
		{name: "bool", in: true, wantErr: "cannot convert bool to int64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt64(tt.in)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err=%v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got=%d err=%v, want %d", got, err, tt.want)
			}
		})
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name    string
		in      any
		want    int
		wantErr string
	}{
		{name: "nil", in: nil, want: 0},
		{name: "int64", in: int64(12), want: 12},
		{name: "int", in: int(12), want: 12},
		{name: "string", in: "34", want: 34},
		{name: "json.Number", in: json.Number("56"), want: 56},
		{name: "json.Number exponent", in: json.Number("1e6"), want: 1_000_000},
		{name: "named int", in: namedInt(9), want: 9},
		{name: "json.Number fractional", in: json.Number("1.5"), wantErr: "cannot convert"},
		{name: "uint64 overflow", in: uint64(math.MaxInt) + 1, wantErr: "cannot convert"},
		{name: "integral float32", in: float32(78), want: 78},
		{name: "fractional float32", in: float32(12.5), wantErr: "cannot convert"},
		{name: "NaN", in: math.NaN(), wantErr: "cannot convert"},
		{name: "out of range float64", in: -float64(math.MinInt), wantErr: "cannot convert"},
		{name: "slice", in: []int{1}, wantErr: "cannot convert []int to int"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt(tt.in)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err=%v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got=%d err=%v, want %d", got, err, tt.want)
			}
		})
	}
}
