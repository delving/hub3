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

package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/delving/hub3/config"
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
