package zdatabase

import (
	"fmt"
	"strconv"
	"testing"
)

func Test01(t *testing.T) {
	aKey := "a"
	aVal := float64(1.5)
	aValS := fmt.Sprintf("%v", aVal)

	SaveGlobalVar(aKey, aValS)
	a2S := LoadGlobalVar(aKey)
	a2, e := strconv.ParseFloat(a2S, 64)
	if e != nil {
		t.Error(e)
	}
	b := 2 * a2
	if b != 3 {
		t.Error()
	}
}
