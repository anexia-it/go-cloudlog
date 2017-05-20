package cloudlog

import (
	"reflect"
	"time"
)

var int64Type = reflect.TypeOf(int64(0))

// ConvertToTimestamp takes an empty interface and tries to convert the value
// to a timestamp as expected by CloudLog (Unix millisecond timestamp).
// Besides being a no-op for int64 values, this function is able to convert
// time.Time values correctly.
//
// If conversion is not possible this function returns the original value.
func ConvertToTimestamp(value interface{}) interface{} {
	switch v := value.(type) {
	case int64:
		// Simple case: already an int64, no-op
		return v
	case time.Time:
		// Convert time.Time
		return v.UTC().UnixNano() / int64(time.Millisecond)
	case *time.Time:
		if v != nil {
			return v.UTC().UnixNano() / int64(time.Millisecond)
		}
	}

	// Conversion not possible, return original value
	return value
}
