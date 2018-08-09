package misc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func Test01(t *testing.T) {
	rand.Seed(time.Now().Unix())
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

func Test04(t *testing.T) {
	//	fmt.Println("NextDay00 ", NextDay00())
	//	fmt.Println("NextDay00 ", NextWeek00())
	//	fmt.Println("NextDay00 ", NextMonth00())
	//	for {
	//		fmt.Println(time.Now())
	//		time.Sleep(1 * time.Second)
	//	}
}

func Test05(t *testing.T) {
	list := CreateLimitedList(3)
	for i := 0; i < 10; i++ {
		list.Append(fmt.Sprintf("%v", i))
	}
	if len(list.Elements) != 3 {
		t.Error()
		return
	}
	if !((list.Elements[0] == "7") && (list.Elements[1] == "8") &&
		(list.Elements[2] == "9")) {
		t.Error(list)
	}
}

func Test06(t *testing.T) {
	list := []int64{10, 11, 12, 13, 14, 15, 16, 17}
	for i := 0; i < 10; i++ {
		//		fmt.Println(ChoiceInt64s(list))
	}
	list = []int64{}
	_ = ChoiceInt64s(list)
}

func TestNormalDistribution(t *testing.T) {

}
