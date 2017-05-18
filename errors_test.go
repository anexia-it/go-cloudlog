package cloudlog_test

import (
	"errors"
	"testing"

	"github.com/anexia-it/go-cloudlog"
	"github.com/stretchr/testify/require"
)

func TestEventEncodingError_Error(t *testing.T) {
	e := &cloudlog.EventEncodingError{
		Message: "test",
	}

	require.EqualValues(t, e.Message, e.Error())
}

func TestIsEventEncodingError(t *testing.T) {
	// Regular error
	ok, event := cloudlog.IsEventEncodingError(errors.New("test"))
	require.False(t, ok)
	require.Nil(t, event)

	// Event encoding error
	e := &cloudlog.EventEncodingError{
		Event: "test event",
	}
	ok, event = cloudlog.IsEventEncodingError(e)
	require.True(t, ok)
	require.EqualValues(t, e.Event, event)
}

func TestMarshalError_Error(t *testing.T) {
	me := &cloudlog.MarshalError{
		Parent: errors.New("test error"),
	}

	require.EqualError(t, me, "Marshal of event failed: test error")
}

func TestMarshalError_WrappedErrors(t *testing.T) {
	parent := errors.New("test error")
	me := &cloudlog.MarshalError{
		Parent: parent,
	}

	wrapped := me.WrappedErrors()
	require.Len(t, wrapped, 1)
	require.EqualError(t, wrapped[0], parent.Error())
}
