package misc

import (
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
	v, _ := vi.(float64)
	return v
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
