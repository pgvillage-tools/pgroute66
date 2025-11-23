package internal

import (
	"encoding/base64"

	"github.com/mannemsolutions/pgroute66/pkg/pg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	var (
		password    = "s3cr3t_p@$$w0rd"
		b64password = base64.StdEncoding.EncodeToString([]byte(password))
		config      = RouteConfig{
			Hosts: RouteHostsConfig{
				"h1": pg.Dsn{"b64password": b64password},
				"h2": pg.Dsn{"password": password},
			},
		}
	)
	Context("initializing", func() {
		It("should work as expected", func() {
			h := NewPgRouteHandler(config)
			Expect(h).NotTo(BeNil())
			Expect(h.config).To(Equal(config))
			Expect(h.config.Hosts).To(HaveLen(2))
			for _, hostname := range []string{"h1", "h2"} {
				Expect(h.config.Hosts).To(HaveKey(hostname))
				host := h.config.Hosts["h1"]
				Expect(host).To(HaveKey("password"))
				Expect(host).NotTo(HaveKey("b64password"))

				Expect(h.connections).To(HaveKey(hostname))
				// conn := h.connections[hostname]
			}
		})
	})
})
