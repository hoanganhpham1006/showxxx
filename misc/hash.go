package misc

import (
	// "fmt"
	"hash/fnv"
)

func HashStringToInt64(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	r := int64(h.Sum64())
	if r < 0 {
		r = -r
	}
	return r
}
