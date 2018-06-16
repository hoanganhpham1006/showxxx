package misc

import (
	"errors"
	"sort"
)

func CheckInt64sIsEqual(a []int64, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type IncreasingOrder []int64

func (a IncreasingOrder) Len() int               { return len(a) }
func (a IncreasingOrder) Less(i int, j int) bool { return a[i] < a[j] }
func (a IncreasingOrder) Swap(i int, j int)      { a[i], a[j] = a[j], a[i] }

// return a new sorted list in increasing order
func SortedInt64s(a []int64) []int64 {
	result := make([]int64, len(a))
	copy(result, a)
	sort.Sort(IncreasingOrder(result))
	return result
}

// Return the lowest index of arg1 in where arg0 is found,
// If not found return -1
func FindStringInSlice(sub string, list []string) int {
	for index, element := range list {
		if sub == element {
			return index
		}
	}
	return -1
}

// Return the lowest index of arg1 in where arg0 is found,
// If not found return -1
func FindInt64InSlice(sub int64, list []int64) int {
	for index, element := range list {
		if sub == element {
			return index
		}
	}
	return -1
}

// Return the index where to insert item x in list a, assuming a is sorted desc.
// The return value i is such that:
// all e in a[:i] have e >= x, and all e in a[i:] have e < x.
// Optional args lo (default 0) and hi (default len(a)) bound the
// slice of a to be searched.
func bisectRight(a []float64, x float64, lo int, hi int) (int, error) {
	if a == nil {
		return 0, errors.New("a is nil")
	}
	if lo < 0 {
		return 0, errors.New("lo must be non-negative")
	}
	for lo < hi {
		mid := (lo + hi) / 2
		if x > a[mid] {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo, nil
}

// Return the index where to insert item x in list a, assuming a is sorted desc
func BisectRight(a []float64, x float64) (int, error) {
	return bisectRight(a, x, 0, len(a))
}
