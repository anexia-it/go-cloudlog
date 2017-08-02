// Package cloudlog provides a CloudLog client library
package cloudlog

import (
	"crypto/tls"
	"encoding/json"
	"reflect"

	"time"

	"sync"

	"github.com/Shopify/sarama"
	multierror "github.com/hashicorp/go-multierror"
)

// CloudLog is the CloudLog object to send logs
type CloudLog struct {
	brokers       []string
	tlsConfig     *tls.Config
	producerMutex sync.RWMutex
	producer      sarama.SyncProducer
	saramaConfig  sarama.Config
	indexName     string
	sourceHost    string
	eventEncoder  EventEncoder
}

// NewCloudLog initializes a new CloudLog instance
func NewCloudLog(indexName string, options ...Option) (cl *CloudLog, err error) {
	if indexName == "" {
		err = ErrIndexNotDefined
		return
	}

	cl = &CloudLog{
		tlsConfig: &tls.Config{},
		indexName: indexName,
	}

	// When returning an error ensure that we return a nil value as *CloudLog
	defer func() {
		if err != nil {
			cl = nil
		}
	}()

	// Apply all options, default options first
	options = append(defaultOptions, options...)
	for _, opt := range options {
		if optErr := opt(cl); optErr != nil {
			err = multierror.Append(err, optErr)
		}
	}

	// At least one option caused an error, bail out
	if err != nil {
		return
	}

	// Enforce TLS and the specific kafka version
	cl.saramaConfig.Version = sarama.V0_10_2_0
	cl.saramaConfig.Net.TLS.Enable = true
	cl.saramaConfig.Net.TLS.Config = cl.tlsConfig

	return
}

// InitCloudLog validates and initializes the CloudLog client
func InitCloudLog(index string, ca string, cert string, key string) (*CloudLog, error) {
	return NewCloudLog(index, OptionCACertificateFile(ca), OptionClientCertificateFile(cert, key))
}

func (cl *CloudLog) getProducer() (sarama.SyncProducer, error) {
	var err error
	cl.producerMutex.RLock()
	defer cl.producerMutex.RUnlock()
	if cl.producer == nil {
		// Connection not yet established, establish connections now

		// First drop the rlock and obtain a wlock
		cl.producerMutex.RUnlock()
		cl.producerMutex.Lock()

		// Ensure that we drop the wlock again and obtain the rlock before returning,
		// so the deferred runlock will not cause a panic
		defer func() {
			cl.producerMutex.Unlock()
			cl.producerMutex.RLock()
		}()
		cl.producer, err = sarama.NewSyncProducer(cl.brokers, &cl.saramaConfig)
	}

	return cl.producer, err
}

// Close closes the connection
func (cl *CloudLog) Close() (err error) {
	cl.producerMutex.Lock()
	defer cl.producerMutex.Unlock()

	// Close should be a no-op if no connection has been established
	if cl.producer != nil {
		err = cl.producer.Close()
		cl.producer = nil
	}
	return
}

// PushEvents sends the supplied events to CloudLog
func (cl *CloudLog) PushEvents(events ...interface{}) (err error) {

	if len(events) == 0 {
		// Bail out early if no events have been passed in
		return
	}

	// Get the producer. This will lazily establish the connection on the first call
	var producer sarama.SyncProducer
	if producer, err = cl.getProducer(); err != nil {
		return
	}

	now := time.Now().UTC()
	timestampMillis := now.UnixNano() / int64(time.Millisecond)

	// Check if is slice
	if len(events) == 1 && reflect.TypeOf(events[0]).Kind() == reflect.Slice {
		var slice []interface{}
		val := reflect.ValueOf(events[0])
		for i := 0; i < val.Len(); i++ {
			slice = append(slice, val.Index(i).Interface())
		}
		events = slice
	}

	// Encode the events
	messages := make([]*sarama.ProducerMessage, len(events))
	for i, ev := range events {
		var eventMap map[string]interface{}
		if eventMap, err = cl.eventEncoder.EncodeEvent(ev); err != nil {
			return err
		}

		// if there is no timestamp field, set it to the current timestamp
		// otherwise try to convert it to epoch millis format
		if _, hasTimestamp := eventMap["timestamp"]; !hasTimestamp {
			eventMap["timestamp"] = timestampMillis
		} else {
			eventMap["timestamp"] = ConvertToTimestamp(eventMap["timestamp"])
		}

		eventMap["cloudlog_source_host"] = cl.sourceHost
		eventMap["cloudlog_client_type"] = "go-client-kafka"

		var eventData []byte
		eventData, err = json.Marshal(eventMap)
		if err != nil {
			return NewMarshalError(eventMap, err)
		}

		messages[i] = &sarama.ProducerMessage{
			Topic:     cl.indexName,
			Value:     sarama.StringEncoder(eventData),
			Timestamp: now,
		}
	}

	return producer.SendMessages(messages)
}

// PushEvent sends an event to CloudLog
func (cl *CloudLog) PushEvent(event interface{}) error {
	return cl.PushEvents(event)
}
