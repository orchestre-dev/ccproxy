package utils

import (
	"encoding/json"
	"time"
)

// ToJSONString converts an interface{} to a JSON string
func ToJSONString(v interface{}) string {
	if v == nil {
		return "{}"
	}

	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}

	return string(data)
}

// GetTimestamp returns the current Unix timestamp
func GetTimestamp() int64 {
	return time.Now().Unix()
}
