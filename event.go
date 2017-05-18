package cloudlog

// Event defines the interface events may optionally implement to provide their own
// encoding logic
type Event interface {
	// Encode encodes the given event to a map[string]interface{}
	Encode() map[string]interface{}
}
