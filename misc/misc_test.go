package misc

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test01(t *testing.T) {
	_ = fmt.Println
	type Case struct {
		Slice0  []int64
		Slice1  []int64
		IsEqual bool
	}
	for i, c := range []Case{
		Case{nil, nil, true},
		Case{[]int64{0, 3, 7, 1, 9}, nil, false},
		Case{nil, []int64{4, 9, 2, 5, 1}, false},
		Case{[]int64{8, 6, 9, 3, 2}, []int64{8, 6, 9, 3, 2}, true},
		Case{[]int64{7, 8, 3}, []int64{7, 8, 3, 7, 6}, false},
		Case{[]int64{8, 6, 9, 3, 2}, []int64{8, 6, 9}, false},
		Case{[]int64{0, 2, 8, 8, 2, 1}, []int64{0, 2, 8, 8, 2, 1}, true},
	} {
		expectation := CheckInt64sIsEqual(c.Slice0, c.Slice1)
		if expectation != c.IsEqual {
			t.Error(i)
		}
	}
}

func Test02(t *testing.T) {
	type Case struct {
		Slice  []int64
		Sorted []int64
	}
	for i, c := range []Case{
		Case{[]int64{0, 3, 7, 1, 9}, []int64{0, 1, 3, 7, 9}},
		Case{[]int64{4, 9, 2, 5, 1}, []int64{1, 2, 4, 5, 9}},
		Case{[]int64{0, 2, 8, 8, 2, 1}, []int64{0, 1, 2, 2, 8, 8}},
		Case{[]int64{7, 8, 3}, []int64{3, 7, 8}},
		Case{[]int64{5}, []int64{5}},
	} {
		expectation := SortedInt64s(c.Slice)
		if !CheckInt64sIsEqual(expectation, c.Sorted) {
			t.Error(i, expectation)
		}
	}
}

func Test03(t *testing.T) {
	j := `{
        "f1": 1, 
        "f2": "true", 
        "f3": true
    }`
	var data map[string]interface{}
	e := json.Unmarshal([]byte(j), &data)
	if e != nil {
		t.Error(e)
	}
	f1 := ReadFloat64(data, "f1")
	f11 := ReadInt64(data, "f1")
	f2 := ReadString(data, "f2")
	f22 := ReadBool(data, "f2")
	f3 := ReadBool(data, "f3")
	f33 := ReadString(data, "f3")
	f4 := ReadFloat64(data, "f4")
	if f1 != 1 {
		t.Error()
	}
	if f11 != 1 {
		t.Error()
	}
	if f2 != "true" {
		t.Error()
	}
	if f22 != false {
		t.Error()
	}
	if f3 != true {
		t.Error()
	}
	if f33 != "" {
		t.Error()
	}
	if f4 != 0 {
		t.Error()
	}
}