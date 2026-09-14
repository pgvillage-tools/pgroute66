package v1

import (
	"encoding/base64"
	"errors"
	"log"
)

// SSLConfig is a combination of an SSL cert and a key
type SSLConfig struct {
	Cert string `yaml:"b64cert"`
	Key  string `yaml:"b64key"`
}

// Enabled returns wether this config is enabled (both cert and key are defined)
func (rsc SSLConfig) Enabled() bool {
	if rsc.Cert != "" && rsc.Key != "" {
		return true
	}
	return false
}

// KeyBytes returns the bytes version of this key
func (rsc SSLConfig) KeyBytes() ([]byte, error) {
	if !rsc.Enabled() {
		return nil, errors.New("cannot get CertBytes when SSL is not enabled")
	}

	return base64.StdEncoding.DecodeString(rsc.Key)
}

// CertBytes returns the bytes value of this cert
func (rsc SSLConfig) CertBytes() ([]byte, error) {
	if !rsc.Enabled() {
		return nil, errors.New("cannot get CertBytes when SSL is not enabled")
	}

	return base64.StdEncoding.DecodeString(rsc.Cert)
}

// MustCertBytes returns the bytes value of this cert, or logs a fatal message
func (rsc SSLConfig) MustCertBytes() []byte {
	cb, err := rsc.CertBytes()
	if err != nil {
		log.Fatal("could not decrypt SSL Cert", err)
	}

	return cb
}
