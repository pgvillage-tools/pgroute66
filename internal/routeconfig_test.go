package internal

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"

	"github.com/mannemsolutions/pgroute66/pkg/pg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteConfig", Ordered, func() {
	var (
		tmpDir string
	)
	BeforeAll(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "pgroute66_config")
		Expect(err).NotTo(HaveOccurred())
	})
	AfterAll(func() {
		os.RemoveAll(tmpDir)
	})
	Context("proper config", func() {
		const config string = `
---
hosts:
  host1:
    host: pgroute66-postgres-1
  host2:
    host: pgroute66-postgres-2

host_groups:
  cluster:
    - host1
    - host2
    - host3


loglevel: debug

bind: 0.0.0.0

port: 8443
ssl:
  b64cert: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2d0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktrd2dn
  b64key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2d0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktrd2dn
`
		It("should succeed", func() {
			configFile := path.Join(tmpDir, "config.yaml")
			Expect(os.WriteFile(configFile, []byte(config), 0o600)).NotTo(HaveOccurred())
			config, err := NewConfigFromFile(configFile, true)
			Expect(config).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Ssl.CertBytes()).NotTo(BeNil())
			Expect(config.Ssl.KeyBytes()).NotTo(BeNil())
			Expect(config.Hosts).To(HaveLen(2))
			Expect(config.LogLevel).To(Equal(debugLoglevel))
			Expect(config.BindTo()).To(Equal("0.0.0.0:8443"))
		})
	})
	Context("debug", func() {
		It("should work as expected", func() {
			debugConfigFile := path.Join(tmpDir, "debug_config.yaml")
			for _, test := range []struct {
				debug    bool
				yml      string
				expected string
			}{
				{debug: false, yml: "loglevel: info", expected: infoLoglevel},
				{debug: false, yml: "loglevel: debug", expected: debugLoglevel},
				{debug: false, yml: "", expected: ""},
				{debug: true, yml: "loglevel: info", expected: debugLoglevel},
				{debug: true, yml: "loglevel: debug", expected: debugLoglevel},
				{debug: true, yml: "", expected: debugLoglevel},
			} {
				Expect(os.WriteFile(debugConfigFile, []byte(test.yml), 0o600)).NotTo(HaveOccurred())
				config, err := NewConfigFromFile(debugConfigFile, test.debug)
				Expect(config).To(Equal(RouteConfig{LogLevel: test.expected}))
				Expect(config.Debug()).To(Equal(test.expected == debugLoglevel))
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
	Context("missing config", func() {
		It("should not succeed", func() {
			missingConfigFile := path.Join(tmpDir, "missing_config.yaml")
			config, err := NewConfigFromFile(missingConfigFile, true)
			Expect(config).To(Equal(RouteConfig{}))
			Expect(err).To(HaveOccurred())
		})
	})
	Context("improper config", func() {
		It("should not succeed", func() {
			const config string = `<>`
			brokenConfigFile := path.Join(tmpDir, "broken_config.yaml")
			Expect(os.WriteFile(brokenConfigFile, []byte(config), 0o600)).NotTo(HaveOccurred())
			brokenConfig, err := NewConfigFromFile(brokenConfigFile, true)
			Expect(brokenConfig).To(Equal(RouteConfig{}))
			Expect(err).To(HaveOccurred())
		})
	})
	Context("improper permissions", func() {
		It("should not succeed", func() {
			const config string = ``
			inaccessibleConfigFile := path.Join(tmpDir, "inaccessible_config.yaml")
			Expect(os.WriteFile(inaccessibleConfigFile, []byte(config), 0o000)).NotTo(HaveOccurred())
			inAccessibleConfig, err := NewConfigFromFile(inaccessibleConfigFile, true)
			Expect(inAccessibleConfig).To(Equal(RouteConfig{}))
			Expect(err).To(HaveOccurred())
		})
	})
	Context("GroupHosts", func() {
		It("should work as expected", func() {
			var config = RouteConfig{
				Hosts: RouteHostsConfig{
					"h1": pg.Dsn{"host": "h1"},
					"h2": pg.Dsn{"host": "h2"},
					"h3": pg.Dsn{"host": "h3"},
					"h4": pg.Dsn{"host": "h4"},
				},
				Groups: RouteHostGroups{
					"c1": RouteHostGroup{"h1", "h2"},
					"c2": RouteHostGroup{"h3", "h4"},
				},
			}
			for _, test := range []struct {
				group string
				hosts []string
			}{
				{group: "all", hosts: []string{"h1", "h2", "h3", "h4"}},
				{group: "c1", hosts: []string{"h1", "h2"}},
				{group: "c2", hosts: []string{"h3", "h4"}},
				{group: "c3", hosts: []string{}},
			} {
				g := config.GroupHosts(test.group)
				Expect(g).To(ContainElements(test.hosts))
			}
		})
	})
	Context("BindTo", func() {
		It("should work as expected", func() {
			const (
				h1    = "s1.local"
				h2    = "1.2.3.4"
				port1 = 1234
				port2 = 5678
			)
			for _, test := range []struct {
				host     string
				port     int
				ssl      bool
				expected string
			}{
				{host: "", port: port1, expected: fmt.Sprintf("%s:%d", "localhost", port1)},
				{host: h1, port: port1, expected: fmt.Sprintf("%s:%d", h1, port1)},
				{host: h2, port: port2, expected: fmt.Sprintf("%s:%d", h2, port2)},
				{host: h2, port: 0, expected: fmt.Sprintf("%s:%d", h2, defaultNoSSLPort)},
				{host: h2, port: 0, ssl: true, expected: fmt.Sprintf("%s:%d", h2, defaultSSLPort)},
			} {
				config := RouteConfig{Bind: test.host, Port: test.port}
				if test.ssl {
					config.Ssl.Cert = base64.StdEncoding.EncodeToString([]byte("--- cert ---"))
					config.Ssl.Key = base64.StdEncoding.EncodeToString([]byte("--- key ---"))
				}
				Expect(config.BindTo()).To(Equal(test.expected))
			}
		})
	})
})
