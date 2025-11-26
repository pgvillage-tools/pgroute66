package pg

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Query", func() {
	var query = Query{}.
	Select("relname").
	qry := strings.Join([]string{
		"select relname",
		"from pg_class",
		"where relname = $1",
		"and relnamespace in",
		"(select oid from pg_namespace where nspname=$2)",
	}, "\n")

	When("building SQL", func() {
		It("should work as expected", func() {
		})
	})
})
