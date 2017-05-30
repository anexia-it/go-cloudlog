package cloudlog

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"os"

	"github.com/Shopify/sarama"
)

// Option defines the type used for applying options to CloudLog
type Option func(*CloudLog) error

// OptionBrokers defines the list of event brokers to use
func OptionBrokers(brokers ...string) Option {
	return func(cl *CloudLog) error {
		if len(brokers) == 0 {
			return ErrBrokersNotSpecified
		}
		cl.brokers = brokers
		return nil
	}
}

// OptionTLSConfig defines the TLS configuration CloudLog uses
func OptionTLSConfig(tlsConfig *tls.Config) Option {
	return func(cl *CloudLog) error {
		cl.tlsConfig = tlsConfig
		return nil
	}
}

// OptionCACertificate sets the CA certificate CloudLog uses
func OptionCACertificate(pemBlock []byte) Option {
	return func(cl *CloudLog) error {
		if cl.tlsConfig.RootCAs == nil {
			cl.tlsConfig.RootCAs = x509.NewCertPool()
		}

		if ok := cl.tlsConfig.RootCAs.AppendCertsFromPEM(pemBlock); !ok {
			return ErrCACertificateInvalid
		}

		return nil
	}
}

// OptionCACertificateFile loads the CA certificate from the supplied paths and
// configures CloudLog to use this CA certificate
func OptionCACertificateFile(path string) Option {
	return func(cl *CloudLog) error {
		pemBlock, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		// Call into OptionCACertificate
		return OptionCACertificate(pemBlock)(cl)
	}
}

// OptionClientCertificates configures CloudLog to use the supplied tls.Certificates
// as client certificates
func OptionClientCertificates(certs []tls.Certificate) Option {
	return func(cl *CloudLog) error {
		if len(certs) == 0 {
			return ErrCertificateMissing
		}
		cl.tlsConfig.Certificates = certs
		return nil
	}
}

// OptionClientCertificateFile configures CloudLog to use the certificate
// and key contained in the supplied paths as client certificate
func OptionClientCertificateFile(certFile, keyFile string) Option {
	return func(cl *CloudLog) error {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)

		if err != nil {
			return err
		}
		return OptionClientCertificates([]tls.Certificate{cert})(cl)
	}
}

// OptionEventEncoder configures the EventEncoder to use for encoding events
func OptionEventEncoder(encoder EventEncoder) Option {
	return func(cl *CloudLog) error {
		cl.eventEncoder = encoder
		return nil
	}
}

// OptionSaramaConfig sets the sarama configuration
func OptionSaramaConfig(config sarama.Config) Option {
	return func(cl *CloudLog) error {
		cl.saramaConfig = config
		return nil
	}
}

// OptionSourceHost configures the sources' hostname
func OptionSourceHost(hostname string) Option {
	return func(cl *CloudLog) error {
		cl.sourceHost = hostname
		return nil
	}
}

// DefaultBrokerAddresses defines the default broker addresses
var DefaultBrokerAddresses = []string{
	"anx-bdp-broker0401.bdp.anexia-it.com:443",
	"anx-bdp-broker0402.bdp.anexia-it.com:443",
	"anx-bdp-broker0403.bdp.anexia-it.com:443",
}

// defaultOptions defines the default options which are applied to a new CloudLog instance
var defaultOptions = []Option{
	OptionBrokers(DefaultBrokerAddresses...),
	OptionEventEncoder(NewAutomaticEventEncoder()),
}

func init() {
	// Use the system's hostname as the default source host
	hostname, _ := os.Hostname()
	defaultOptions = append(defaultOptions, OptionSourceHost(hostname))
}
