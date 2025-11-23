package pg

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Functions", func() {
	Context("identifierNameSQL", func() {
		It("should work as expected", func() {
			for _, test := range []struct {
				in  string
				out string
			}{
				{in: "1", out: `"1"`},
				{in: `name_with_"`, out: `"name_with_"""`},
				{in: `a b c d`, out: `"a b c d"`},
				{in: `this is a "word"`, out: `"this is a ""word"""`},
			} {
				Expect(identifierNameSQL(test.in)).To(Equal(test.out))
			}
		})
	})
	Context("connectStringValue", func() {
		It("should work as expected", func() {
			for _, test := range []struct {
				in  string
				out string
			}{
				{in: `1`, out: `'1'`},
				{in: `''`, out: `'\'\''`},
				{in: `a b c d`, out: `'a b c d'`},
				{in: `this is a 'word'`, out: `'this is a \'word\''`},
			} {
				Expect(connectStringValue(test.in)).To(Equal(test.out))
			}
		})
	})
})
