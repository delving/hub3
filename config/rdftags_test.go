package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/delving/rapid-saas/config"
)

var _ = Describe("Rdftags", func() {

	Describe("when creating a new one", func() {

		Context("from configuration", func() {

			c := &RawConfig{
				RDFTag: RDFTag{
					Label:     []string{"http://purl.org/dc/elements/1.1/title"},
					Thumbnail: []string{"http://xmlns.com/foaf/0.1/depiction"},
				},
			}
			tm := NewRDFTagMap(c)

			It("should create a tagMap", func() {
				Expect(tm).ToNot(BeNil())
				Expect(tm.Len()).To(Equal(2))
			})

			It("should return a label for a URI", func() {
				label, ok := tm.Get("http://xmlns.com/foaf/0.1/depiction")
				Expect(ok).To(BeTrue())
				Expect(label).To(ContainElement("thumbnail"))
				Expect(label).To(HaveLen(1))
			})

		})
	})

})
