package misc

import (
	"encoding/base64"
	"fmt"
	"time"
)

// get value from map field as string
func ReadString(data map[string]interface{}, field string) string {
	vi := data[field]
	v, _ := vi.(string)
	return v
}

// get value from map field as float64
func ReadFloat64(data map[string]interface{}, field string) float64 {
	vi := data[field]
	_ = fmt.Println
	v, isOk := vi.(float64)
	if isOk {
		return v
	}
	v2, isOk := vi.(int64)
	if isOk {
		return float64(v2)
	}
	v3, _ := vi.(int)
	return float64(v3)
}

// get value from map field as int64
func ReadInt64(data map[string]interface{}, field string) int64 {
	return int64(ReadFloat64(data, field))
}

// get value from map field as bool
func ReadBool(data map[string]interface{}, field string) bool {
	vi := data[field]
	v, _ := vi.(bool)
	return v
}

// get value from map field as time.Time
func ReadTime(data map[string]interface{}, field string) time.Time {
	vi := data[field]
	vs, _ := vi.(string)
	v, _ := time.Parse(time.RFC3339Nano, vs)
	return v
}

// get value from map string base64 encoded field to []byte,
// return []bytes{} if error occured
func ReadBytes(data map[string]interface{}, field string) []byte {
	vi := data[field]
	vs, _ := vi.(string)
	vbs, _ := base64.StdEncoding.DecodeString(vs)
	return vbs
}
