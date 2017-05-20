package cloudlog

import (
	"crypto/tls"

	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"
)

// rsaCertPEM is taken from crypto/tls/tls_test.go
var rsaCertPEM = `-----BEGIN CERTIFICATE-----
MIIB0zCCAX2gAwIBAgIJAI/M7BYjwB+uMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTIwOTEyMjE1MjAyWhcNMTUwOTEyMjE1MjAyWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANLJ
hPHhITqQbPklG3ibCVxwGMRfp/v4XqhfdQHdcVfHap6NQ5Wok/4xIA+ui35/MmNa
rtNuC+BdZ1tMuVCPFZcCAwEAAaNQME4wHQYDVR0OBBYEFJvKs8RfJaXTH08W+SGv
zQyKn0H8MB8GA1UdIwQYMBaAFJvKs8RfJaXTH08W+SGvzQyKn0H8MAwGA1UdEwQF
MAMBAf8wDQYJKoZIhvcNAQEFBQADQQBJlffJHybjDGxRMqaRmDhX0+6v02TUKZsW
r5QuVbpQhH6u+0UgcW0jp9QwpxoPTLTWGXEWBBBurxFwiCBhkQ+V
-----END CERTIFICATE-----
`

// esaKeyPEM is taken from crypto/tls/tls_test.go
var rsaKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBANLJhPHhITqQbPklG3ibCVxwGMRfp/v4XqhfdQHdcVfHap6NQ5Wo
k/4xIA+ui35/MmNartNuC+BdZ1tMuVCPFZcCAwEAAQJAEJ2N+zsR0Xn8/Q6twa4G
6OB1M1WO+k+ztnX/1SvNeWu8D6GImtupLTYgjZcHufykj09jiHmjHx8u8ZZB/o1N
MQIhAPW+eyZo7ay3lMz1V01WVjNKK9QSn1MJlb06h/LuYv9FAiEA25WPedKgVyCW
SmUwbPw8fnTcpqDWE3yTO3vKcebqMSsCIBF3UmVue8YU3jybC3NxuXq3wNm34R8T
xVLHwDXh/6NJAiEAl2oHGGLz64BuAfjKrqwz7qMYr9HCLIe/YsoWq/olzScCIQDi
D2lWusoe2/nEqfDVVWGWlyJ7yOmqaVm/iNUN9B2N2g==
-----END RSA PRIVATE KEY-----
`

func TestOptionTLSConfig(t *testing.T) {
	tlsConfig := &tls.Config{}
	cl := &CloudLog{}
	require.NoError(t, OptionTLSConfig(tlsConfig)(cl))
	require.EqualValues(t, tlsConfig, cl.tlsConfig)
}

func TestOptionBrokers(t *testing.T) {
	expectedBrokers := []string{"test0:443", "test1:443"}
	cl := &CloudLog{}
	require.NoError(t, OptionBrokers(expectedBrokers[0], expectedBrokers[1:]...)(cl))
	require.EqualValues(t, expectedBrokers, cl.brokers)
}

func TestOptionCACertificate(t *testing.T) {
	// Check if adding an invalid certificate fails, as expected
	cl := &CloudLog{
		tlsConfig: &tls.Config{},
	}

	require.EqualError(t, OptionCACertificate(nil)(cl), ErrCACertificateInvalid.Error())

	// Check if everything works using a valid certificate
	require.NoError(t, OptionCACertificate([]byte(rsaCertPEM))(cl))
	require.Len(t, cl.tlsConfig.RootCAs.Subjects(), 1)
}

func TestOptionCACertificateFile(t *testing.T) {
	cl := &CloudLog{
		tlsConfig: &tls.Config{},
	}

	tempDirPath, err := ioutil.TempDir("", "go-cloudlog-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDirPath)

	fileName := filepath.Join(tempDirPath, "ca.pem")

	// Use a non-existing file, which should return an error
	err = OptionCACertificateFile(fileName)(cl)
	require.True(t, os.IsNotExist(err))

	// Write the file and re-try, which should not cause any errors
	require.NoError(t, ioutil.WriteFile(fileName, []byte(rsaCertPEM), 0600))
	require.NoError(t, OptionCACertificateFile(fileName)(cl))
	require.Len(t, cl.tlsConfig.RootCAs.Subjects(), 1)
}

func TestOptionClientCertificates(t *testing.T) {
	cl := &CloudLog{
		tlsConfig: &tls.Config{},
	}

	// Check that this gives an error if the certificate has not been defined
	require.EqualError(t, OptionClientCertificates(nil)(cl), ErrCertificateMissing.Error())

	// Check that if a certificate is supplied it is used
	cert, err := tls.X509KeyPair([]byte(rsaCertPEM), []byte(rsaKeyPEM))
	require.NoError(t, err)

	require.NoError(t, OptionClientCertificates([]tls.Certificate{cert})(cl))
	require.EqualValues(t, []tls.Certificate{cert}, cl.tlsConfig.Certificates)
}

func TestOptionClientCertificateFile(t *testing.T) {
	cl := &CloudLog{
		tlsConfig: &tls.Config{},
	}

	tempDirPath, err := ioutil.TempDir("", "go-cloudlog-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDirPath)
	certFile := filepath.Join(tempDirPath, "cert.pem")
	keyFile := filepath.Join(tempDirPath, "key.pem")

	// Non-existent files
	err = OptionClientCertificateFile(certFile, keyFile)(cl)
	require.True(t, os.IsNotExist(err))

	// Non-existent key
	require.NoError(t, ioutil.WriteFile(certFile, []byte(rsaCertPEM), 0600))
	err = OptionClientCertificateFile(certFile, keyFile)(cl)
	require.True(t, os.IsNotExist(err))

	// Both existent
	require.NoError(t, ioutil.WriteFile(keyFile, []byte(rsaKeyPEM), 0600))
	require.NoError(t, OptionClientCertificateFile(certFile, keyFile)(cl))
	require.Len(t, cl.tlsConfig.Certificates, 1)

	// Validate that the certificate matches
	cert, err := tls.X509KeyPair([]byte(rsaCertPEM), []byte(rsaKeyPEM))
	require.NoError(t, err)
	require.EqualValues(t, []tls.Certificate{cert}, cl.tlsConfig.Certificates)
}

func TestOptionEventEncoder(t *testing.T) {
	cl := &CloudLog{}
	enc := NewAutomaticEventEncoder()
	require.NoError(t, OptionEventEncoder(enc)(cl))
	require.Equal(t, enc, cl.eventEncoder)
}

func TestOptionSaramaConfig(t *testing.T) {
	cl := &CloudLog{}
	sc := sarama.NewConfig()
	sc.ClientID = "test_client_id"
	require.NoError(t, OptionSaramaConfig(*sc)(cl))

	require.EqualValues(t, "test_client_id", cl.saramaConfig.ClientID)
}

func TestOptionSourceHost(t *testing.T) {
	cl := &CloudLog{}
	require.NoError(t, OptionSourceHost("test-host")(cl))
	require.EqualValues(t, "test-host", cl.sourceHost)
}
