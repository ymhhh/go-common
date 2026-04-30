package types

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// Contains returns true if the slice contains the value
func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Index returns the index of the first occurrence of value in slice, or -1 if not found.
func Index[T comparable](slice []T, value T) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

// Filter filters slice elements based on predicate function.
func Filter[T any](slice []T, fn func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map applies a function to each element and returns a new slice.
func Map[T any, R any](slice []T, fn func(T) R) []R {
	result := make([]R, 0, len(slice))
	for _, v := range slice {
		result = append(result, fn(v))
	}
	return result
}

// Reduce reduces slice to a single value using accumulator function.
func Reduce[T any, R any](slice []T, initial R, fn func(R, T) R) R {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// Remove removes the first occurrence of value from slice.
func Remove[T comparable](slice []T, value T) []T {
	index := Index(slice, value)
	if index == -1 {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// RemoveAll removes all occurrences of value from slice
func RemoveAll[T comparable](slice []T, value T) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}

// Unique removes duplicate elements from slice
func Unique[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[T]bool, len(slice))
	result := make([]T, 0, len(slice))

	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Chunk splits slice into chunks of specified size
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		panic("chunk size must be greater than 0")
	}

	if len(slice) == 0 {
		return nil
	}

	chunks := make([][]T, 0, (len(slice)+size-1)/size)
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// Reverse reverses the order of elements in slice
func Reverse[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// ReverseInPlace reverses slice in place
func ReverseInPlace[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Intersect returns elements that exist in both slices
func Intersect[T comparable](slice1, slice2 []T) []T {
	if len(slice1) == 0 || len(slice2) == 0 {
		return nil
	}

	set := make(map[T]bool, len(slice2))
	for _, v := range slice2 {
		set[v] = true
	}

	result := make([]T, 0)
	for _, v := range slice1 {
		if set[v] {
			result = append(result, v)
			delete(set, v) // Avoid duplicates
		}
	}
	return result
}

// Union returns all unique elements from both slices
func Union[T comparable](slice1, slice2 []T) []T {
	set := make(map[T]bool, len(slice1)+len(slice2))
	result := make([]T, 0, len(slice1)+len(slice2))

	for _, v := range slice1 {
		if !set[v] {
			set[v] = true
			result = append(result, v)
		}
	}

	for _, v := range slice2 {
		if !set[v] {
			set[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Difference returns elements in slice1 but not in slice2
func Difference[T comparable](slice1, slice2 []T) []T {
	if len(slice1) == 0 {
		return nil
	}
	if len(slice2) == 0 {
		return slice1
	}

	set := make(map[T]bool, len(slice2))
	for _, v := range slice2 {
		set[v] = true
	}

	result := make([]T, 0)
	for _, v := range slice1 {
		if !set[v] {
			result = append(result, v)
		}
	}
	return result
}

// Any checks if any element satisfies the predicate
func Any[T any](slice []T, fn func(T) bool) bool {
	for _, v := range slice {
		if fn(v) {
			return true
		}
	}
	return false
}

// All checks if all elements satisfy the predicate
func All[T any](slice []T, fn func(T) bool) bool {
	for _, v := range slice {
		if !fn(v) {
			return false
		}
	}
	return true
}

// First returns the first element that satisfies the predicate, or zero value and false
func First[T any](slice []T, fn func(T) bool) (T, bool) {
	var zero T
	for _, v := range slice {
		if fn(v) {
			return v, true
		}
	}
	return zero, false
}

// Count counts elements that satisfy the predicate
func Count[T any](slice []T, fn func(T) bool) int {
	count := 0
	for _, v := range slice {
		if fn(v) {
			count++
		}
	}
	return count
}

// Distinct returns distinct elements based on key function
func Distinct[T any, K comparable](slice []T, keyFn func(T) K) []T {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[K]bool, len(slice))
	result := make([]T, 0, len(slice))

	for _, v := range slice {
		key := keyFn(v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}
	return result
}

// Flatten flattens a slice of slices into a single slice
func Flatten[T any](slices [][]T) []T {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, 0, totalLen)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// Partition splits slice into two slices based on predicate
func Partition[T any](slice []T, fn func(T) bool) ([]T, []T) {
	trueSlice := make([]T, 0, len(slice))
	falseSlice := make([]T, 0, len(slice))

	for _, v := range slice {
		if fn(v) {
			trueSlice = append(trueSlice, v)
		} else {
			falseSlice = append(falseSlice, v)
		}
	}
	return trueSlice, falseSlice
}

var defaultRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// Shuffle shuffles slice using Fisher-Yates algorithm
func Shuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	for i := len(result) - 1; i > 0; i-- {
		j := defaultRand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// Take returns first n elements of slice
func Take[T any](slice []T, n int) []T {
	if n <= 0 {
		return nil
	}
	if n >= len(slice) {
		return slice
	}
	result := make([]T, n)
	copy(result, slice[:n])
	return result
}

// Drop returns slice with first n elements removed
func Drop[T any](slice []T, n int) []T {
	if n <= 0 {
		return slice
	}
	if n >= len(slice) {
		return nil
	}
	return slice[n:]
}

// Zip combines two slices into pairs
func Zip[T any, U any](slice1 []T, slice2 []U) []Pair[T, U] {
	minLen := len(slice1)
	if len(slice2) < minLen {
		minLen = len(slice2)
	}

	result := make([]Pair[T, U], minLen)
	for i := 0; i < minLen; i++ {
		result[i] = Pair[T, U]{First: slice1[i], Second: slice2[i]}
	}
	return result
}

// Pair represents a pair of values
type Pair[T any, U any] struct {
	First  T
	Second U
}

// Unzip separates pairs into two slices
func Unzip[T any, U any](pairs []Pair[T, U]) ([]T, []U) {
	if len(pairs) == 0 {
		return nil, nil
	}

	firsts := make([]T, len(pairs))
	seconds := make([]U, len(pairs))

	for i, p := range pairs {
		firsts[i] = p.First
		seconds[i] = p.Second
	}
	return firsts, seconds
}

// Convert converts slice from one type to another using converter function
func Convert[T any, R any](slice []T, converter func(T) R) []R {
	if slice == nil {
		return nil
	}
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = converter(v)
	}
	return result
}

// Sort sorts slice using less function.
func Sort[T any](slice []T, less func(T, T) bool) []T {
	sort.Slice(slice, func(i, j int) bool {
		return less(slice[i], slice[j])
	})
	return slice
}

// Sum calculates sum of numeric slice
func Sum[T Number](slice []T) T {
	var sum T
	for _, v := range slice {
		sum += v
	}
	return sum
}

// Number is a type constraint for numeric types
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

// Max returns maximum value in slice
func Max[T Number](slice []T) (T, error) {
	if len(slice) == 0 {
		var zero T
		return zero, fmt.Errorf("cannot find max of empty slice")
	}

	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max, nil
}

// Min returns minimum value in slice
func Min[T Number](slice []T) (T, error) {
	if len(slice) == 0 {
		var zero T
		return zero, fmt.Errorf("cannot find min of empty slice")
	}

	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min, nil
}

// Average calculates average of numeric slice
func Average[T Number](slice []T) (float64, error) {
	if len(slice) == 0 {
		return 0, fmt.Errorf("cannot calculate average of empty slice")
	}

	sum := Sum(slice)
	return float64(sum) / float64(len(slice)), nil
}
