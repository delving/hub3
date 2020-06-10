// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/delving/hub3/hub3/fragments"
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
