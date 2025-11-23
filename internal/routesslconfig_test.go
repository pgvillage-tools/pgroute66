package internal

import (
	"encoding/base64"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routesslconfig", func() {
	Context("a valid RouteSSLConfig is defined", func() {
		const (
			cert = "--- cert ---"
			key  = "--- key ---"
		)
		It("should be enabled", func() {
			var (
				rsc = RouteSSLConfig{
					Cert: base64.StdEncoding.EncodeToString([]byte(cert)),
					Key:  base64.StdEncoding.EncodeToString([]byte(key)),
				}
			)
			Expect(rsc.Enabled()).To(BeTrue())
		})
	})
})
