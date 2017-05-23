package cloudlog

import structmapper "gopkg.in/anexia-it/go-structmapper.v1"

// DefaultTagName defines the default tag name to use
const DefaultTagName = "cloudlog"

// EventEncoder defines the interface for encoding events
type EventEncoder interface {
	// EncodeEvent encodes the given event
	EncodeEvent(event interface{}) (map[string]interface{}, error)
}

// AutomaticEventEncoder tries to find the right encoder for the given input
type AutomaticEventEncoder struct {
	Encoders []EventEncoder
}

// NewAutomaticEventEncoder returns a new encoder that supports all available encoders
func NewAutomaticEventEncoder() *AutomaticEventEncoder {
	encoder := &AutomaticEventEncoder{}
	structEncoder, _ := NewStructEncoder()
	encoder.Encoders = []EventEncoder{&SimpleEventEncoder{}, structEncoder}
	return encoder
}

// EncodeEvent encodes the given event
func (e *AutomaticEventEncoder) EncodeEvent(event interface{}) (map[string]interface{}, error) {
	for _, encoder := range e.Encoders {
		result, err := encoder.EncodeEvent(event)
		if err == nil {
			return result, nil
		}
	}
	return nil, NewUnsupportedEventType(event)
}

// StructEncoder implements an encoder which can encode structs using gopkg.in/anexia-it/go-structmapper.v1
type StructEncoder struct {
	mapper *structmapper.Mapper
}

// EncodeEvent encodes the given event
func (e *StructEncoder) EncodeEvent(event interface{}) (m map[string]interface{}, err error) {
	if m, err = e.mapper.ToMap(event); err != nil {
		m = nil
		return
	}
	m, err = structmapper.ForceStringMapKeys(m)
	return
}

// NewStructEncoder returns a new encoder that supports structs
func NewStructEncoder(options ...structmapper.Option) (*StructEncoder, error) {
	mapper, err := structmapper.NewMapper(append([]structmapper.Option{
		structmapper.OptionTagName(DefaultTagName)}, options...)...)
	if err != nil {
		return nil, err
	}

	return &StructEncoder{
		mapper: mapper,
	}, nil
}

// SimpleEventEncoder implements a simple event encoder
// This encoder only supports map[string]interface{}, string and []byte events.
// A more sophisticated encoder providing support for encoding structs as well is available
// from the structencoder sub-package.
type SimpleEventEncoder struct {
}

// EncodeEvent encodes the given event
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
