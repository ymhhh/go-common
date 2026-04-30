package types

import (
	"testing"
)

func TestContains_Index_Remove_Unique(t *testing.T) {
	s := []int{1, 2, 3, 2}
	if !Contains(s, 2) {
		t.Fatalf("Contains")
	}
	if Index(s, 3) != 2 {
		t.Fatalf("Index")
	}
	if got := Remove(s, 2); len(got) != 3 || !Contains(got, 2) {
		t.Fatalf("Remove: %#v", got)
	}
	if got := RemoveAll(s, 2); len(got) != 2 {
		t.Fatalf("RemoveAll: %#v", got)
	}
	if got := Unique([]string{"a", "b", "a"}); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("Unique: %#v", got)
	}
}

func TestChunk_PanicsOnBadSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = Chunk([]int{1, 2}, 0)
}

func TestSort_MutatesSlice(t *testing.T) {
	s := []int{3, 1, 2}
	_ = Sort(s, func(a, b int) bool { return a < b })
	if s[0] != 1 || s[1] != 2 || s[2] != 3 {
		t.Fatalf("sort: %#v", s)
	}
}

func TestConvert_NilSlice(t *testing.T) {
	var in []int
	out := Convert(in, func(v int) string { return "" })
	if out != nil {
		t.Fatalf("Convert nil: %#v", out)
	}
}

func TestSum_Max_Min_Average(t *testing.T) {
	if got := Sum([]int{1, 2, 3}); got != 6 {
		t.Fatalf("Sum: %v", got)
	}
	if _, err := Max([]int{}); err == nil {
		t.Fatalf("Max empty: expected err")
	}
	if _, err := Min([]int{}); err == nil {
		t.Fatalf("Min empty: expected err")
	}
	if _, err := Average([]float64{}); err == nil {
		t.Fatalf("Average empty: expected err")
	}
	if avg, err := Average([]float64{1, 2, 3}); err != nil || avg != 2 {
		t.Fatalf("Average: %v err=%v", avg, err)
	}
}

func TestZip_Unzip(t *testing.T) {
	a := []int{1, 2}
	b := []string{"x", "y"}
	pairs := Zip(a, b)
	if len(pairs) != 2 || pairs[0].First != 1 || pairs[0].Second != "x" {
		t.Fatalf("Zip: %#v", pairs)
	}
	aa, bb := Unzip(pairs)
	if len(aa) != 2 || len(bb) != 2 || aa[1] != 2 || bb[1] != "y" {
		t.Fatalf("Unzip: %#v %#v", aa, bb)
	}
}
