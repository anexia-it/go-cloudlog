package cloudlog_test

import (
	"testing"

	"github.com/anexia-it/go-cloudlog"
	"github.com/stretchr/testify/require"
)

type mockEvent struct {
	value int
}

func (m *mockEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"value": m.value,
	}
}

func TestSimpleEventEncoder_EncodeEvent(t *testing.T) {
	enc := &cloudlog.SimpleEventEncoder{}

	// Test type that implements the Event interface
	ev := &mockEvent{
		value: 12345,
	}

	m, err := enc.EncodeEvent(ev)
	require.NoError(t, err)
	require.EqualValues(t, ev.Encode(), m)

	// Test pass-through of map[string]interface{}
	expected := map[string]interface{}{
		"test1": 1,
		"test2": map[string]interface{}{
			"test2.1": 2.1,
		},
		"test3": "3",
	}

	m, err = enc.EncodeEvent(expected)
	require.NoError(t, err)
	require.EqualValues(t, expected, m)

	// Test string
	stringInput := "test string"
	expected = map[string]interface{}{
		"message": stringInput,
	}
	m, err = enc.EncodeEvent(stringInput)
	require.NoError(t, err)
	require.EqualValues(t, expected, m)

	// Test []byte
	bytesInput := []byte("test bytes")
	expected = map[string]interface{}{
		"message": string(bytesInput),
	}
	m, err = enc.EncodeEvent(bytesInput)
	require.NoError(t, err)
	require.EqualValues(t, expected, m)

	// Test unsupported type
	m, err = enc.EncodeEvent(0)
	require.Error(t, err)
	require.Nil(t, m)
	isEventEncodingError, eventData := cloudlog.IsEventEncodingError(err)
	require.True(t, isEventEncodingError)
	require.EqualValues(t, 0, eventData)
}

type TestStruct struct {
	Property string `cloudlog:"test" cl:"test2"`
}

func TestStructEncoder_EncodeEvent(t *testing.T) {
	// Test with default tag name
	enc, err := cloudlog.NewStructEncoder()
	require.NoError(t, err)
	require.NotNil(t, enc)

	m, err := enc.EncodeEvent(&TestStruct{"test0"})
	require.NoError(t, err)
	require.NotNil(t, m)
	require.EqualValues(t, map[string]interface{}{"test": "test0"}, m)
}
