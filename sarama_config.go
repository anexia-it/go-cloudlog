package cloudlog

import (
	"time"

	"github.com/Shopify/sarama"
)

var defaultSaramaConfig = sarama.NewConfig()

func init() {
	defaultSaramaConfig.Net.DialTimeout = time.Second * 5
	defaultSaramaConfig.Net.WriteTimeout = time.Second * 30
	defaultSaramaConfig.Net.ReadTimeout = time.Second * 30
	defaultSaramaConfig.Net.KeepAlive = time.Second * 10
	defaultSaramaConfig.Net.MaxOpenRequests = 10
	defaultSaramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	defaultSaramaConfig.Producer.Retry.Max = 10
	defaultSaramaConfig.Producer.Return.Successes = true
	defaultSaramaConfig.Producer.Return.Errors = true
	defaultSaramaConfig.Version = sarama.V0_10_2_0

	// Update the default options just after initializing defaultSaramaConfig
	defaultOptions = append(defaultOptions, OptionSaramaConfig(GetDefaultSaramaConfig()))
}

// GetDefaultSaramaConfig returns a copy of the default sarama config.
// The configuration returned by this function should be used as a basline configuration
// for modifications and changes.
func GetDefaultSaramaConfig() sarama.Config {
	return *defaultSaramaConfig
}
