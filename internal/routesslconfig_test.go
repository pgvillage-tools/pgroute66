package internal

import (
	"encoding/base64"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routesslconfig", func() {
	Context("a valid RouteSSLConfig is defined", func() {
		var (
			cert = []byte("--- cert ---")
			key  = []byte("--- key ---")
			rsc  = RouteSSLConfig{
				Cert: base64.StdEncoding.EncodeToString([]byte(cert)),
				Key:  base64.StdEncoding.EncodeToString([]byte(key)),
			}
		)
		It("should be enabled", func() {
			Expect(rsc.Enabled()).To(BeTrue())
		})
		It("should not panic on key / cert", func() {
			Expect(func() { _ = rsc.MustCertBytes() }).NotTo(Panic())
			Expect(func() { _ = rsc.MustKeyBytes() }).NotTo(Panic())
		})
		It("should deliver cert bytes properly", func() {
			bytes, err := rsc.CertBytes()
			Expect(bytes).To(Equal(cert))
			Expect(err).NotTo(HaveOccurred())
		})
		It("should deliver key bytes properly", func() {
			bytes, err := rsc.KeyBytes()
			Expect(bytes).To(Equal(key))
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("an invalid RouteSSLConfig is defined", func() {
		var rsc = RouteSSLConfig{}
		It("should be disabled", func() {
			Expect(rsc.Enabled()).To(BeFalse())
		})
		// We should test this too, but before we can we should replace all
		// logging implementation with zerolog
		It("should panic on key / cert", func() {
			Expect(func() { _ = rsc.MustCertBytes() }).To(Panic())
			Expect(func() { _ = rsc.MustKeyBytes() }).To(Panic())
		})
		It("should not deliver cert bytes", func() {
			bytes, err := rsc.CertBytes()
			Expect(bytes).To(BeEmpty())
			Expect(err).To(HaveOccurred())
		})
		It("should not deliver key bytes", func() {
			bytes, err := rsc.KeyBytes()
			Expect(bytes).To(BeEmpty())
			Expect(err).To(HaveOccurred())
		})
	})
})
