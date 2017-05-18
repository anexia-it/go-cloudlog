package cloudlog

import (
	"errors"
	"testing"

	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"
)

func MockOptionWithError(_ *CloudLog) error {
	return errors.New("mock option error")
}

func TestNewCloudLog(t *testing.T) {
	// Missing indexName
	cl, err := NewCloudLog("")
	require.EqualError(t, err, ErrIndexNotDefined.Error())
	require.Nil(t, cl)

	// Default options, mock broker
	cl, err = NewCloudLog("test")
	require.NoError(t, err)
	require.NotNil(t, cl)
	// Ensure cl.Close() is called
	cl.Close()

	// Option that returns an error
	cl, err = NewCloudLog("test", MockOptionWithError)
	require.Error(t, err)
	require.Nil(t, cl)

	require.IsType(t, &multierror.Error{}, err)
	errorWrapper := err.(*multierror.Error)
	require.Len(t, errorWrapper.WrappedErrors(), 1)
	require.EqualError(t, errorWrapper.WrappedErrors()[0], "mock option error")
}

func TestCloudLog_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := &CloudLog{}

	// No producer set, should be a no-op
	require.NoError(t, cl.Close())

	// Set mock producer
	mockProducer := NewMockSyncProducer(ctrl)
	mockProducer.EXPECT().Close().Return(errors.New("test error"))
	cl.producer = mockProducer
	require.EqualError(t, cl.Close(), "test error")
	require.Nil(t, cl.producer)
}

func TestCloudLog_PushEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := &CloudLog{
		sourceHost: "test-host",
	}
	// Set mock producer
	mockProducer := NewMockSyncProducer(ctrl)
	cl.producer = mockProducer

	// Set mock encoder
	mockEncoder := NewMockEventEncoder(ctrl)
	cl.eventEncoder = mockEncoder

	// Test failure in EncodeEvent
	mockEncoder.EXPECT().EncodeEvent("test event").Times(1).Return(nil, errors.New("test error"))
	require.EqualError(t, cl.PushEvents("test event"), "test error")

	// Test failure in JSON marshalling
	expectedMap := map[string]interface{}{
		"test": func() {},
	}
	mockEncoder.EXPECT().EncodeEvent("test event 2").Times(1).Return(expectedMap, nil)

	err := cl.PushEvents("test event 2")
	require.Error(t, err)
	require.IsType(t, &MarshalError{}, err)
	require.EqualValues(t, expectedMap, err.(*MarshalError).EventMap)

	// Test successful push of multiple events
	cl.eventEncoder = &SimpleEventEncoder{}
	mockProducer.EXPECT().SendMessages(gomock.Any()).Times(1).Do(func(msgs []*sarama.ProducerMessage) {
		require.Len(t, msgs, 3)
		for i, msg := range msgs {
			require.EqualValues(t, cl.indexName, msg.Topic)
			var msgData map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(msg.Value.(sarama.StringEncoder)), &msgData))
			require.EqualValues(t, fmt.Sprintf("test%d", i), msgData["message"])
			require.EqualValues(t, "go-client", msgData["cloudlog_client_type"])
			require.EqualValues(t, cl.sourceHost, msgData["cloudlog_source_host"])

		}
	}).Return(errors.New("test error"))
	require.EqualError(t, cl.PushEvents("test0", "test1", "test2"), "test error")
}

func TestCloudLog_PushEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cl := &CloudLog{
		sourceHost: "test-host",
	}
	// Set mock producer
	mockProducer := NewMockSyncProducer(ctrl)
	cl.producer = mockProducer
	cl.eventEncoder = &SimpleEventEncoder{}
	mockProducer.EXPECT().SendMessages(gomock.Any()).Times(1).Do(func(msgs []*sarama.ProducerMessage) {
		require.Len(t, msgs, 1)
		require.EqualValues(t, cl.indexName, msgs[0].Topic)
		var msgData map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(msgs[0].Value.(sarama.StringEncoder)), &msgData))
		require.EqualValues(t, "test0", msgData["message"])
		require.EqualValues(t, "go-client", msgData["cloudlog_client_type"])
		require.EqualValues(t, cl.sourceHost, msgData["cloudlog_source_host"])
	}).Return(errors.New("test error"))

	require.EqualError(t, cl.PushEvent("test0"), "test error")
}
