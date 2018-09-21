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
	if r < 1000000000 {
		r += 1000000000
	}
	return r
}
