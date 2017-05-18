package cloudlog

// EventEncoder defines the interface for encoding events
type EventEncoder interface {
	// EncodeEvent encodes the given event
	EncodeEvent(event interface{}) (map[string]interface{}, error)
}

// SimpleEventEncoder implements a simple event encoder
// This encoder only supports map[string]interface{}, string and []byte events.
// A more sophisticated encoder providing support for encoding structs as well is available
// from the structencoder sub-package.
type SimpleEventEncoder struct {
}

func (e *SimpleEventEncoder) EncodeEvent(event interface{}) (map[string]interface{}, error) {
	if enc, ok := event.(Event); ok {
		return enc.Encode(), nil
	}

	// Handle simple cases: map[string]interface{}, string, and []byte
	switch value := event.(type) {
	case map[string]interface{}:
		return value, nil
	case string:
		return map[string]interface{}{"message": value}, nil
	case []byte:
		return map[string]interface{}{"message": string(value)}, nil
	}

	return nil, NewUnsupportedEventType(event)
}
