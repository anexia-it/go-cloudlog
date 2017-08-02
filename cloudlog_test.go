package cloudlog

import (
	"errors"
	"testing"

	"encoding/json"
	"fmt"

	"os"
	"time"

	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"path/filepath"

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

	// Validate that the configuration has been applied as expected
	require.EqualValues(t, "test", cl.indexName)

	// Ensure the default sarama config has been applied
	require.EqualValues(t, time.Second*5, cl.saramaConfig.Net.DialTimeout)
	require.EqualValues(t, time.Second*30, cl.saramaConfig.Net.WriteTimeout)
	require.EqualValues(t, time.Second*30, cl.saramaConfig.Net.ReadTimeout)
	require.EqualValues(t, time.Second*10, cl.saramaConfig.Net.KeepAlive)
	require.EqualValues(t, 10, cl.saramaConfig.Net.MaxOpenRequests)
	require.EqualValues(t, sarama.WaitForAll, cl.saramaConfig.Producer.RequiredAcks)
	require.EqualValues(t, 10, cl.saramaConfig.Producer.Retry.Max)
	require.True(t, cl.saramaConfig.Producer.Return.Successes)
	require.True(t, cl.saramaConfig.Producer.Return.Errors)
	require.EqualValues(t, sarama.V0_10_2_0, cl.saramaConfig.Version)

	require.NotNil(t, cl.tlsConfig)
	require.NotNil(t, cl.eventEncoder)
	require.IsType(t, NewAutomaticEventEncoder(), cl.eventEncoder)

	hostname, err := os.Hostname()
	require.NoError(t, err)
	require.EqualValues(t, hostname, cl.sourceHost)

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

	// Passing no events should not do anything but return nil
	// We verify this by not configuring the mockEncoder at all, which would cause an error
	// if it was called
	require.NoError(t, cl.PushEvents())

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
	cl.eventEncoder = NewAutomaticEventEncoder()
	nowMillis := time.Now().UTC().UnixNano() / int64(time.Millisecond)
	mockProducer.EXPECT().SendMessages(gomock.Any()).Times(1).Do(func(msgs []*sarama.ProducerMessage) {
		require.Len(t, msgs, 3)
		for i, msg := range msgs {
			require.EqualValues(t, cl.indexName, msg.Topic)
			var msgData map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(msg.Value.(sarama.StringEncoder)), &msgData))

			// Ensure that all values have been set
			require.InDelta(t, nowMillis, msgData["timestamp"], float64(time.Second))
			require.EqualValues(t, fmt.Sprintf("test%d", i), msgData["message"])
			require.EqualValues(t, "go-client-kafka", msgData["cloudlog_client_type"])
			require.EqualValues(t, cl.sourceHost, msgData["cloudlog_source_host"])

		}
	}).Return(errors.New("test error"))
	require.EqualError(t, cl.PushEvents("test0", "test1", "test2"), "test error")

	// Test successful push of multiple events as slice
	cl.eventEncoder = NewAutomaticEventEncoder()
	mockProducer.EXPECT().SendMessages(gomock.Any()).Times(1).Do(func(msgs []*sarama.ProducerMessage) {
		require.Len(t, msgs, 3)
		for i, msg := range msgs {
			require.EqualValues(t, cl.indexName, msg.Topic)
			var msgData map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(msg.Value.(sarama.StringEncoder)), &msgData))

			// Ensure that all values have been set
			require.InDelta(t, nowMillis, msgData["timestamp"], float64(time.Second))
			require.EqualValues(t, fmt.Sprintf("test%d", i), msgData["message"])
			require.EqualValues(t, "go-client-kafka", msgData["cloudlog_client_type"])
			require.EqualValues(t, cl.sourceHost, msgData["cloudlog_source_host"])

		}
	}).Return(errors.New("test error"))
	require.EqualError(t, cl.PushEvents([]string{"test0", "test1", "test2"}), "test error")
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

	// Push a single simple string event
	nowMillis := time.Now().UTC().UnixNano() / int64(time.Millisecond)
	mockProducer.EXPECT().SendMessages(gomock.Any()).Times(1).Do(func(msgs []*sarama.ProducerMessage) {
		require.Len(t, msgs, 1)
		require.EqualValues(t, cl.indexName, msgs[0].Topic)
		var msgData map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(msgs[0].Value.(sarama.StringEncoder)), &msgData))

		// Ensure that all values have been set
		require.InDelta(t, nowMillis, msgData["timestamp"], float64(time.Second))
		require.EqualValues(t, "test0", msgData["message"])
		require.EqualValues(t, "go-client-kafka", msgData["cloudlog_client_type"])
		require.EqualValues(t, cl.sourceHost, msgData["cloudlog_source_host"])
	}).Return(errors.New("test error"))
	require.EqualError(t, cl.PushEvent("test0"), "test error")

	// Push an event with an existing timestamp
	expectedTimestamp := int64(14952277322252)
	mockProducer.EXPECT().SendMessages(gomock.Any()).Times(1).Do(func(msgs []*sarama.ProducerMessage) {
		require.Len(t, msgs, 1)
		require.EqualValues(t, cl.indexName, msgs[0].Topic)
		var msgData map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(msgs[0].Value.(sarama.StringEncoder)), &msgData))

		// Ensure that all values have been set
		require.EqualValues(t, expectedTimestamp, msgData["timestamp"])
		require.EqualValues(t, "test value", msgData["test_property"])
		require.EqualValues(t, "go-client-kafka", msgData["cloudlog_client_type"])
		require.EqualValues(t, cl.sourceHost, msgData["cloudlog_source_host"])
	}).Return(errors.New("test error 2"))
	require.EqualError(t, cl.PushEvent(map[string]interface{}{
		"test_property": "test value",
		"timestamp":     expectedTimestamp,
	}), "test error 2")

	// Push an event with an existing timestamp
	ts := time.Now()
	// Wait 250 ms to ensure the timestamp is not overridden
	time.Sleep(time.Millisecond * 250)
	mockProducer.EXPECT().SendMessages(gomock.Any()).Times(1).Do(func(msgs []*sarama.ProducerMessage) {
		require.Len(t, msgs, 1)
		require.EqualValues(t, cl.indexName, msgs[0].Topic)
		var msgData map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(msgs[0].Value.(sarama.StringEncoder)), &msgData))

		// Ensure that all values have been set
		require.EqualValues(t, ts.UTC().UnixNano()/int64(time.Millisecond), msgData["timestamp"])
		require.EqualValues(t, "test value 2", msgData["test_property"])
		require.EqualValues(t, "go-client-kafka", msgData["cloudlog_client_type"])
		require.EqualValues(t, cl.sourceHost, msgData["cloudlog_source_host"])
	}).Return(errors.New("test error 3"))
	require.EqualError(t, cl.PushEvent(map[string]interface{}{
		"test_property": "test value 2",
		"timestamp":     ts,
	}), "test error 3")
}

func TestInitCloudLog(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-cloudlog-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Write files
	caPath := filepath.Join(tmpDir, "ca.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")
	certPath := filepath.Join(tmpDir, "cert.pem")
	require.NoError(t, ioutil.WriteFile(caPath, []byte(rsaCertPEM), 0600))
	require.NoError(t, ioutil.WriteFile(certPath, []byte(rsaCertPEM), 0600))
	require.NoError(t, ioutil.WriteFile(keyPath, []byte(rsaKeyPEM), 0600))

	cl, err := InitCloudLog("testIndex", caPath, certPath, keyPath)
	require.NoError(t, err)
	require.NotNil(t, cl)
	require.EqualValues(t, "testIndex", cl.indexName)
	require.NotNil(t, cl.tlsConfig)
	require.NotNil(t, cl.tlsConfig.RootCAs)
	require.Len(t, cl.tlsConfig.RootCAs.Subjects(), 1)

	require.Len(t, cl.tlsConfig.Certificates, 1)
	cert := cl.tlsConfig.Certificates[0]
	require.Len(t, cert.Certificate, 1)
	rsaCertPEMBlock, _ := pem.Decode([]byte(rsaCertPEM))
	require.NotNil(t, rsaCertPEMBlock)
	require.NotNil(t, rsaCertPEMBlock.Bytes)
	require.EqualValues(t, rsaCertPEMBlock.Bytes, cert.Certificate[0])

	require.IsType(t, &rsa.PrivateKey{}, cert.PrivateKey)
	privateKey := cert.PrivateKey.(*rsa.PrivateKey)
	rsaKeyPEMBlock, _ := pem.Decode([]byte(rsaKeyPEM))
	require.NotNil(t, rsaKeyPEMBlock)
	require.NotNil(t, rsaKeyPEMBlock.Bytes)
	require.EqualValues(t, rsaKeyPEMBlock.Bytes, x509.MarshalPKCS1PrivateKey(privateKey))

}
