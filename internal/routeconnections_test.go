package internal

import (
	"fmt"

	"github.com/mannemsolutions/pgroute66/pkg/pg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routeconnections", Ordered, func() {
	rcs := RouteConnections{}
	for i := 0; i < 10; i++ {
		host := fmt.Sprintf("db-%d", i)
		rcs[host] = pg.NewConn(pg.Dsn{"host": host})
	}
	BeforeAll(func() {
	})
	Context("a valid RouteConnections is defined", func() {
		It("should filter properly", func() {
			host1and5 := rcs.FilteredConnections([]string{"db-1", "db-5"})
			Expect(host1and5).To(HaveKey("db-1"))
			Expect(host1and5).To(HaveKey("db-5"))
			Expect(host1and5).NotTo(HaveKey("db-0"))
			Expect(host1and5).NotTo(HaveKey("db-6"))
			Expect(host1and5).NotTo(HaveKey("db-9"))
		})
	})
})
