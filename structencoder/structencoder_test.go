package structencoder_test

import (
	"testing"

	"github.com/anexia-it/go-cloudlog/structencoder"
	"github.com/stretchr/testify/require"
	"gopkg.in/anexia-it/go-structmapper.v1"
)

type TestStruct struct {
	Property string `cloudlog:"test" cl:"test2"`
}

func TestStructEncoder_EncodeEvent(t *testing.T) {
	// Test with default tag name
	enc, err := structencoder.NewStructEncoder()
	require.NoError(t, err)
	require.NotNil(t, enc)

	m, err := enc.EncodeEvent(&TestStruct{"test0"})
	require.NoError(t, err)
	require.NotNil(t, m)
	require.EqualValues(t, map[string]interface{}{"test": "test0"}, m)

	// Test with custom tag name "cl"
	enc, err = structencoder.NewStructEncoder(structmapper.OptionTagName("cl"))
	require.NoError(t, err)
	require.NotNil(t, enc)

	m, err = enc.EncodeEvent(&TestStruct{"test0"})
	require.NoError(t, err)
	require.NotNil(t, m)
	require.EqualValues(t, map[string]interface{}{"test2": "test0"}, m)

	// Test if errors are reported through
	m, err = enc.EncodeEvent(0)
	require.EqualError(t, err, structmapper.ErrNotAStruct.Error())
	require.Nil(t, m)
}
