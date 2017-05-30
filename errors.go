package cloudlog

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrCACertificateInvalid indicates that the supplied CA certificate is invalid
	ErrCACertificateInvalid = errors.New("CA certificate is invalid")

	// ErrCertificateMissing indicates that the client certificate is missing
	ErrCertificateMissing = errors.New("Client certificate is missing")

	// ErrIndexNotDefined indicates that the target index has not been defined
	ErrIndexNotDefined = errors.New("Target index is not defined")

	// ErrBrokersNotSpecified indicates that no brokers have been specified
	ErrBrokersNotSpecified = errors.New("Brokers not specified")
)

// EventEncodingError indicates that an event could not be encoded
type EventEncodingError struct {
	Message string
	Event   interface{}
}

func (e *EventEncodingError) Error() string {
	return e.Message
}

// IsEventEncodingError checks if a supplied error is an EventEncodingError
// and returns a boolean flag and the event that caused the error
func IsEventEncodingError(err error) (ok bool, event interface{}) {
	if e, ok := err.(*EventEncodingError); ok {
		return true, e.Event
	}

	return false, nil
}

// NewUnsupportedEventType constructs a new EventEncodingError that indicates that the
// supplied event type is unsupported
func NewUnsupportedEventType(event interface{}) *EventEncodingError {
	return &EventEncodingError{
		Message: fmt.Sprintf("Cannot encode event, type %s is unsupported",
			reflect.ValueOf(event).Type().String()),
		Event: event,
	}
}

// MarshalError represents a marshalling error
type MarshalError struct {
	// EventMap contains the events data map
	EventMap map[string]interface{}
	// Parent contains the parent error
	Parent error
}

func (e *MarshalError) Error() string {
	return fmt.Sprintf("Marshal of event failed: %s", e.Parent.Error())
}

// WrappedErrors returns the wrapped parent error
func (e *MarshalError) WrappedErrors() []error {
	return []error{e.Parent}
}

// NewMarshalError returns a new MarshalError
func NewMarshalError(eventMap map[string]interface{}, parent error) *MarshalError {
	return &MarshalError{
		EventMap: eventMap,
		Parent:   parent,
	}
}
