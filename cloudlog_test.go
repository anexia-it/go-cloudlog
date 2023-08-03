package cloudlog

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func MockOptionWithError(_ *CloudLog) error {
	return errors.New("mock option error")
}

type MockClient struct {
	m map[string]interface{}
}

func (mc *MockClient) Do(req *http.Request) (resp *http.Response, err error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}

	resp = &http.Response{
		Body: io.NopCloser(strings.NewReader("")),
	}
	resp.StatusCode = 201
	mc.m = m
	return
}

func TestPushSimpleEvent(t *testing.T) {
	req := require.New(t)

	conf := NewDefaultConfig()
	mc := &MockClient{}
	conf.Client = mc
	conf.Hostname = "test-host"
	cl, err := NewCloudlogWithConfig("abc123", "token", conf)
	req.NoError(err)

	//simple event
	err = cl.PushEvent("test message")
	req.NoError(err)

	m := mc.m

	records := m["records"].([]interface{})
	req.Equal(1, len(records))
	record1 := records[0].(map[string]interface{})
	req.Equal("test message", record1["message"])
	req.Equal("test-host", record1["cloudlog_source_host"])
	req.Equal("go-client-rest", record1["cloudlog_client_type"])
	req.InDelta(time.Now().Unix()*1000, record1["timestamp"], 1000)
}

func TestPushEvent(t *testing.T) {
	req := require.New(t)

	conf := NewDefaultConfig()
	mc := &MockClient{}
	conf.Client = mc
	conf.Hostname = "test-host"
	cl, err := NewCloudlogWithConfig("abc123", "token", conf)
	req.NoError(err)

	//simple event
	input := map[string]interface{}{
		"message": "test message",
		"value":   1,
	}
	err = cl.PushEvent(input)
	req.NoError(err)

	m := mc.m

	records := m["records"].([]interface{})
	req.Equal(1, len(records))
	record1 := records[0].(map[string]interface{})
	req.Equal("test message", record1["message"])

	req.Equal(1.0, record1["value"])

	req.Equal("test-host", record1["cloudlog_source_host"])
	req.Equal("go-client-rest", record1["cloudlog_client_type"])
	req.InDelta(time.Now().Unix()*1000, record1["timestamp"], 1000)
}
