package pg

import (
	"context"
	"fmt"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conn", func() {
	const (
		host1 = "s1.local"
		port1 = "2345"
	)
	var (
		ctx = context.Background()
	)
	Context("NewConn", func() {
		It("should work as expected", func() {
			conn := NewConn(Dsn{
				"host": host1,
				"port": port1,
			})
			Expect(conn).NotTo(BeNil())
			Expect(conn.connParams).To(HaveKey("host"))
			Expect(conn.connParams).To(HaveKey("port"))
			Expect(conn.endpoint).To(Equal(fmt.Sprintf("%s:%s", host1, port1)))
			Expect(conn.DSN()).To(Equal(fmt.Sprintf("host='%s' port='%s'", host1, port1)))
		})
		It("should try to connect", func() {
			tmpDir, tmpErr := os.MkdirTemp("", "pg_conn")
			Expect(tmpErr).NotTo(HaveOccurred())
			defer os.RemoveAll(tmpDir)
			conn := NewConn(Dsn{
				"host": path.Join(tmpDir, "does_not_exist"),
				"port": "1",
			})
			err := conn.Connect(ctx)
			Expect(err.Error()).To(ContainSubstring("failed to connect to "))
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		})
		It("should not connect with overrides", func() {
			var (
				key = OverrideKey{
					Query: "select 1",
				}
				hash     = key.Hash()
				override = OverrideResult{
					rows: Result{map[string]any{"": "primary"}}}
			)
			conn := NewConn(Dsn{
				"port": "1",
			})
			conn.Override(Overrides{hash: override})
			Expect(conn.Connect(ctx)).NotTo(HaveOccurred())
			Expect(conn.conn).To(BeNil())
		})
		It("should error on invalid dsn", func() {
			conn := NewConn(Dsn{
				"invalid": "key",
			})
			err := conn.Connect(ctx)
			Expect(err.Error()).To(ContainSubstring("failed to connect to "))
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		})
	})
	Context("Host", func() {
		It("should work as expected", func() {
			const (
				pgHostEnvKey = "PGHOST"
				host2        = "s2.local"
			)
			orgPgHost := os.Getenv(pgHostEnvKey)
			defer os.Setenv(pgHostEnvKey, orgPgHost)
			for _, envVar := range []string{"", host2} {
				os.Setenv(pgHostEnvKey, envVar)
				for _, host := range []string{"", "localhost", host1} {
					conn := Conn{connParams: Dsn{"host": host}}
					if host == "" {
						host = envVar
					}
					if host == "" {
						host = "localhost"
					}
					Expect(conn.Host()).To(Equal(host))
				}
			}
		})
	})
	Context("Port", func() {
		It("should work as expected", func() {
			const (
				pgPortEnvKey = "PGPORT"
				port2        = "3456"
			)
			orgPgPort := os.Getenv(pgPortEnvKey)
			defer os.Setenv(pgPortEnvKey, orgPgPort)
			for _, envVar := range []string{"", port2} {
				os.Setenv(pgPortEnvKey, envVar)
				for _, port := range []string{"", "1234", port1} {
					conn := Conn{connParams: Dsn{"port": port}}
					if port == "" {
						port = envVar
					}
					if port == "" {
						port = "5432"
					}
					Expect(conn.Port()).To(Equal(port))
				}
			}
		})
	})
	Context("IsPrimary", func() {
		It("should work as expected", func() {
			for _, test := range []struct {
				result   OverrideResult
				expected bool
			}{
				{expected: true, result: OverrideResult{
					rows: Result{map[string]any{"": "primary"}}}},
				{result: OverrideResult{rows: Result{}}},
			} {
				var (
					key = OverrideKey{
						Query: "select 'primary' where not pg_is_in_recovery()",
					}
					hash = key.Hash()
					conn = Conn{}
				)
				conn.Override(Overrides{hash: test.result})
				isPrimary, err := conn.IsPrimary(ctx)
				Expect(isPrimary).To(Equal(test.expected))
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
	Context("IsStandby", func() {
		It("should work as expected", func() {
			for _, test := range []struct {
				result   OverrideResult
				expected bool
			}{
				{expected: true, result: OverrideResult{
					rows: Result{map[string]any{"": "standby"}}}},
				{result: OverrideResult{rows: Result{}}},
			} {
				var (
					key = OverrideKey{
						Query: "select 'standby' where pg_is_in_recovery()",
					}
					hash = key.Hash()
					conn = Conn{
						overrides: Overrides{hash: test.result},
					}
				)
				isStandby, err := conn.IsStandby(ctx)
				Expect(isStandby).To(Equal(test.expected))
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})
