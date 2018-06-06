package misc

import (
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
