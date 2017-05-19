package cloudlog_test

import (
	"testing"
	"time"

	"github.com/anexia-it/go-cloudlog"
	"github.com/stretchr/testify/require"
)

func TestConvertToTimestamp(t *testing.T) {
	// Test case: int64 pass-through
	expectedInt64 := int64(12345)
	require.EqualValues(t, expectedInt64, cloudlog.ConvertToTimestamp(expectedInt64))

	// Test case: conversion of a time.Time
	now := time.Now()
	expectedInt64 = now.UTC().UnixNano() / int64(time.Millisecond)
	require.EqualValues(t, expectedInt64, cloudlog.ConvertToTimestamp(now))

	// Test case: conversion of a *time.Time
	require.EqualValues(t, expectedInt64, cloudlog.ConvertToTimestamp(&now))

	// Test case: *time.Time nil value
	require.EqualValues(t, int64(0), cloudlog.ConvertToTimestamp((*time.Time)(nil)))

	// Test case: other, unsupported values
	require.EqualValues(t, int64(0), cloudlog.ConvertToTimestamp(uint64(12345)))
	require.EqualValues(t, int64(0), cloudlog.ConvertToTimestamp(float64(1234.5)))
	require.EqualValues(t, int64(0), cloudlog.ConvertToTimestamp(float32(1234.5)))
	require.EqualValues(t, int64(0), cloudlog.ConvertToTimestamp(nil))
	require.EqualValues(t, int64(0), cloudlog.ConvertToTimestamp("test"))
}
