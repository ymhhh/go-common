package types

import (
	"encoding/json"
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

	if got, err := ToInt64(uint64(1<<63 - 1)); err != nil || got != 1<<63-1 {
		t.Fatalf("max uint64 fitting int64: got=%d err=%v", got, err)
	}
	if got, err := ToInt64(uint64(1 << 63)); err == nil {
		t.Fatalf("uint64 overflow should error, got=%d", got)
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

	maxInt := uint64(^uint(0) >> 1)
	if got, err := ToInt(maxInt); err != nil || got != int(maxInt) {
		t.Fatalf("max uint64 fitting int: got=%d err=%v", got, err)
	}
	if got, err := ToInt(maxInt + 1); err == nil {
		t.Fatalf("uint64 overflow should error, got=%d", got)
	}

	// invalid type
	if _, err := ToInt([]int{1}); err == nil {
		t.Fatalf("expected error for slice")
	}
}
