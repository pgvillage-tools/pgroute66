package pg

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Override", func() {
	When("GetOverride", func() {
		var (
			key1 = OverrideKey{Query: "select 1"}
			key2 = OverrideKey{Query: "select 2"}
			o    = Overrides{key1.Hash(): OverrideResult{}}
		)
		It("should work as expected", func() {
			Expect(o.GetOverride(key1)).NotTo(BeNil())
		})
		It("should panic when the wrong override is requested", func() {
			Expect(func() { _ = o.GetOverride(key2) }).To(Panic())
		})
	})
})
