// Package cloudlog provides a CloudLog client library
package cloudlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"time"
)

// CloudLog is the CloudLog object to send logs
type CloudLog struct {
	client   Client
	url      string
	token    string
	hostname string
	encoder  EventEncoder
}

// NewCloudLog initializes a new CloudLog instance with the default config
func NewCloudLog(indexName, token string) (cl *CloudLog, err error) {
	return NewCloudlogWithConfig(indexName, token, NewDefaultConfig())
}

// NewCloudlogWithConfig initializes a new CloudLog instance using the provided config
func NewCloudlogWithConfig(indexName, token string, config *Config) (cl *CloudLog, err error) {
	if indexName == "" {
		err = ErrIndexNotDefined
		return
	}

	url := fmt.Sprintf("https://api0401.bdp.anexia-it.com/v1/index/%s/data", indexName)

	cl = &CloudLog{
		url:      url,
		token:    token,
		client:   config.Client,
		hostname: config.Hostname,
		encoder:  config.Encoder,
	}

	return
}

// PushEvents sends the supplied events to CloudLog
func (cl *CloudLog) PushEvents(events ...interface{}) (err error) {
	return cl.push(events)
}

// PushEvent sends an event to CloudLog
func (cl *CloudLog) PushEvent(event interface{}) error {
	return cl.PushEvents(event)
}

func (cl *CloudLog) push(events []interface{}) (err error) {
	if len(events) == 0 {
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
	messages := make([]map[string]interface{}, len(events))
	for i, ev := range events {
		var eventMap map[string]interface{}
		if eventMap, err = cl.encoder.EncodeEvent(ev); err != nil {
			return err
		}

		// if there is no timestamp field, set it to the current timestamp
		// otherwise try to convert it to epoch millis format
		if _, hasTimestamp := eventMap["timestamp"]; !hasTimestamp {
			eventMap["timestamp"] = timestampMillis
		} else {
			eventMap["timestamp"] = ConvertToTimestamp(eventMap["timestamp"])
		}

		eventMap["cloudlog_source_host"] = cl.hostname
		eventMap["cloudlog_client_type"] = "go-client-rest"

		messages[i] = eventMap
	}

	err = cl.send(messages)
	return
}

func (cl *CloudLog) send(messages []map[string]interface{}) error {
	request := map[string]interface{}{
		"records": messages,
	}

	var eventData []byte
	eventData, err := json.Marshal(request)
	if err != nil {
		return NewMarshalError(request, err)
	}

	req, err := http.NewRequest(http.MethodPost, cl.url, bytes.NewReader(eventData))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", cl.token)

	resp, err := cl.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 201 {
		return fmt.Errorf("expecting StatusCode 201 but received %d", resp.StatusCode)
	}

	return nil
}
