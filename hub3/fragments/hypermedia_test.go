package fragments_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/delving/rapid-saas/hub3/fragments"
)

var _ = Describe("Hypermedia", func() {

	Describe("when creating now controls", func() {

		Context("from a http.Request", func() {

			base := "https://localhost:3000/fragments"
			query := "?object=true"
			r, err := http.NewRequest("GET", base+query+"&page=2", nil)

			It("should set the correct fullPath", func() {
				Expect(err).ToNot(HaveOccurred())
				fr := NewFragmentRequest()
				err := fr.ParseQueryString(r.URL.Query())
				hmd := NewHyperMediaDataSet(r, 295, fr)
				Expect(err).ToNot(HaveOccurred())
				Expect(hmd).ToNot(BeNil())
				Expect(hmd.DataSetURI).To(Equal(base))
				Expect(hmd.PagerURI).To(Equal(base + query + "&page=2"))
				Expect(hmd.TotalItems).To(Equal(int64(295)))
				Expect(hmd.CurrentPage).To(Equal(int32(2)))
				Expect(hmd.FirstPage).To(Equal(base + query + "&page=1"))
				Expect(hmd.PreviousPage).To(Equal(base + query + "&page=1"))
				Expect(hmd.NextPage).To(Equal(base + query + "&page=3"))
				Expect(hmd.ItemsPerPage).To(Equal(int64(FRAGMENT_SIZE)))
				Expect(hmd.HasNext()).To(BeFalse())
				Expect(hmd.HasPrevious()).To(BeTrue())
			})

			It("should create the controls", func() {
				fr := NewFragmentRequest()
				err := fr.ParseQueryString(r.URL.Query())
				hmd := NewHyperMediaDataSet(r, 395, fr)
				Expect(hmd).ToNot(BeNil())
				b, err := hmd.CreateControls()
				Expect(err).ToNot(HaveOccurred())
				Expect(b).ToNot(BeEmpty())
				Expect(hmd.HasNext()).To(BeTrue())
				Expect(hmd.HasPrevious()).To(BeTrue())
			})

		})
	})

})
