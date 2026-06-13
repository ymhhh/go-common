package types

import (
	"encoding/json"
	"math"
	"testing"
)

func TestToInt64(t *testing.T) {
	// nil
	if got, err := ToInt64(nil); err != nil || got != 0 {
		t.Fatalf("nil: got=%d err=%v", got, err)
	}

	// ints
	if got, err := ToInt64(int64(12)); err != nil || got != 12 {
		t.Fatalf("int64: got=%d err=%v", got, err)
	}
	if got, err := ToInt64(int32(12)); err != nil || got != 12 {
		t.Fatalf("int32: got=%d err=%v", got, err)
	}
	if got, err := ToInt64(int(12)); err != nil || got != 12 {
		t.Fatalf("int: got=%d err=%v", got, err)
	}

	// string number
	if got, err := ToInt64("34"); err != nil || got != 34 {
		t.Fatalf("string: got=%d err=%v", got, err)
	}

	// json.Number is type string kind
	if got, err := ToInt64(json.Number("56")); err != nil || got != 56 {
		t.Fatalf("json.Number: got=%d err=%v", got, err)
	}
	if got, err := ToInt64(json.Number("1e6")); err != nil || got != 1_000_000 {
		t.Fatalf("json.Number exponent: got=%d err=%v", got, err)
	}
	if _, err := ToInt64(json.Number("1.5")); err == nil {
		t.Fatalf("expected error for fractional json.Number")
	}

	// unsigned values beyond int64 must not wrap negative
	if _, err := ToInt64(uint64(math.MaxInt64) + 1); err == nil {
		t.Fatalf("expected error for overflowing uint64")
	}

	// integral floats are accepted, but unsafe float-to-int truncation is not
	if got, err := ToInt64(float64(78)); err != nil || got != 78 {
		t.Fatalf("integral float64: got=%d err=%v", got, err)
	}
	if _, err := ToInt64(12.5); err == nil {
		t.Fatalf("expected error for fractional float64")
	}
	if _, err := ToInt64(math.Inf(1)); err == nil {
		t.Fatalf("expected error for infinite float64")
	}
	if _, err := ToInt64(float64(math.MaxInt64)); err == nil {
		t.Fatalf("expected error for out-of-range float64")
	}

	// invalid type
	if _, err := ToInt64(true); err == nil {
		t.Fatalf("expected error for bool")
	}
}

func TestToInt(t *testing.T) {
	// nil
	if got, err := ToInt(nil); err != nil || got != 0 {
		t.Fatalf("nil: got=%d err=%v", got, err)
	}

	// ints
	if got, err := ToInt(int64(12)); err != nil || got != 12 {
		t.Fatalf("int64: got=%d err=%v", got, err)
	}
	if got, err := ToInt(int(12)); err != nil || got != 12 {
		t.Fatalf("int: got=%d err=%v", got, err)
	}

	// string number
	if got, err := ToInt("34"); err != nil || got != 34 {
		t.Fatalf("string: got=%d err=%v", got, err)
	}

	// json.Number
	if got, err := ToInt(json.Number("56")); err != nil || got != 56 {
		t.Fatalf("json.Number: got=%d err=%v", got, err)
	}
	if got, err := ToInt(json.Number("1e6")); err != nil || got != 1_000_000 {
		t.Fatalf("json.Number exponent: got=%d err=%v", got, err)
	}
	if _, err := ToInt(json.Number("1.5")); err == nil {
		t.Fatalf("expected error for fractional json.Number")
	}

	// values beyond int must not silently wrap when narrowed
	if _, err := ToInt(uint64(math.MaxInt) + 1); err == nil {
		t.Fatalf("expected error for overflowing uint64")
	}

	// integral floats are accepted, but unsafe float-to-int truncation is not
	if got, err := ToInt(float32(78)); err != nil || got != 78 {
		t.Fatalf("integral float32: got=%d err=%v", got, err)
	}
	if _, err := ToInt(float32(12.5)); err == nil {
		t.Fatalf("expected error for fractional float32")
	}
	if _, err := ToInt(math.NaN()); err == nil {
		t.Fatalf("expected error for NaN float64")
	}
	if _, err := ToInt(-float64(math.MinInt)); err == nil {
		t.Fatalf("expected error for out-of-range float64")
	}

	// invalid type
	if _, err := ToInt([]int{1}); err == nil {
		t.Fatalf("expected error for slice")
	}
}
