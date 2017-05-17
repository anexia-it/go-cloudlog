// Package cloudlog provides a CloudLog client library
package cloudlog

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"time"

	"github.com/Shopify/sarama"
)

// CloudLog is the CloudLog object to send logs
type CloudLog struct {
	Index    string
	CAFile   string
	CertFile string
	KeyFile  string
	Producer sarama.SyncProducer
}

// Default broker list
var brokers = []string{"anx-bdp-broker0401.bdp.anexia-it.com:443", "anx-bdp-broker0402.bdp.anexia-it.com:443", "anx-bdp-broker0403.bdp.anexia-it.com:443"}

// InitCloudLog validates and initalizes the CloudLog client
func InitCloudLog(index string, ca string, cert string, key string) (*CloudLog, error) {

	// create CloudLog struct
	c := &CloudLog{
		Index:    index,
		CAFile:   ca,
		CertFile: cert,
		KeyFile:  key,
	}

	// try to connect
	err := c.connect()

	// check error
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Close closes the connection
func (c *CloudLog) Close() error {
	err := c.Producer.Close()
	if err != nil {
		return errors.New("error while closing producer: " + err.Error())
	}
	return nil
}

// PushEvent sends an event to CloudLog
func (c *CloudLog) PushEvent(event string) error {
	return c.PushEvents([]string{event})
}

// PushEvents sends one or more events to CloudLog
func (c *CloudLog) PushEvents(events []string) error {

	var messages []*sarama.ProducerMessage

	for _, event := range events {
		messages = append(messages, &sarama.ProducerMessage{
			Topic:     c.Index,
			Value:     sarama.StringEncoder(event),
			Timestamp: time.Now(),
		})
	}

	err := c.Producer.SendMessages(messages)
	if err != nil {
		return err
	}

	return nil
}

// Try to establish a producer to CloudLog
func (c *CloudLog) connect() error {

	var err error

	tlsConfig, err := createTLSConfiguration(c)
	if err != nil {
		return errors.New("invalid tls configuration: " + err.Error())
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_10_2_0
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = tlsConfig
	c.Producer, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return errors.New("producer could not be created: " + err.Error())
	}

	return nil
}

// Create TLS config from ca, cert and key file
func createTLSConfiguration(c *CloudLog) (*tls.Config, error) {

	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, err
	}

	caCert, err := ioutil.ReadFile(c.CAFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	t := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	return t, nil
}
