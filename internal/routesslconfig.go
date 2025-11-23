package internal

import (
	"encoding/base64"
	"errors"
	"log"
)

// RouteSSLConfig is a combination of an SSL cert and a key
type RouteSSLConfig struct {
	Cert string `yaml:"b64cert"`
	Key  string `yaml:"b64key"`
}

// Enabled returns wether this config is enabled (both cert and key are defined)
func (rsc RouteSSLConfig) Enabled() bool {
	if rsc.Cert != "" && rsc.Key != "" {
		return true
	}
	return false
}

// KeyBytes returns the bytes version of this key
func (rsc RouteSSLConfig) KeyBytes() ([]byte, error) {
	if !rsc.Enabled() {
		return nil, errors.New("cannot get CertBytes when SSL is not enabled")
	}

	return base64.StdEncoding.DecodeString(rsc.Key)
}

// MustKeyBytes returns the bytes value of this key, or logs a fatal message
func (rsc RouteSSLConfig) MustKeyBytes() []byte {
	kb, err := rsc.KeyBytes()
	if err != nil {
		globalHandler.log.Fatal("could not decrypt SSL key", err)
	}

	return kb
}

// CertBytes returns the bytes value of this cert
func (rsc RouteSSLConfig) CertBytes() ([]byte, error) {
	if !rsc.Enabled() {
		return nil, errors.New("cannot get CertBytes when SSL is not enabled")
	}

	return base64.StdEncoding.DecodeString(rsc.Cert)
}

// MustCertBytes returns the bytes value of this cert, or logs a fatal message
func (rsc RouteSSLConfig) MustCertBytes() []byte {
	cb, err := rsc.CertBytes()
	if err != nil {
		log.Fatal("could not decrypt SSL Cert", err)
	}

	return cb
}
