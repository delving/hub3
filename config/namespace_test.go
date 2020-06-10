// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config_test

import (
	c "github.com/delving/hub3/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespace", func() {

	Describe("Has a NameSpaceMap", func() {

		Context("When creating a New NameSpaceMap", func() {

			It("Should create a Map", func() {
				nsMap := c.NewNameSpaceMap()
				Expect(nsMap).ToNot(BeNil())
			})
		})

		Context("When adding a key to the NamespaceMap", func() {

			It("should have no items", func() {
				nsMap := c.NewNameSpaceMap()
				prefix2base, base2prefix := nsMap.Len()
				Expect(prefix2base).To(Equal(0))
				Expect(prefix2base).To(Equal(base2prefix))
			})

			It("Should add the key Map", func() {
				nsMap := c.NewNameSpaceMap()
				nsMap.Add("dc", "http://purl.org/dc/elements/1.1/")
				Expect(nsMap).ToNot(BeNil())
				prefix2base, base2prefix := nsMap.Len()
				Expect(prefix2base).To(Equal(1))
				Expect(prefix2base).To(Equal(base2prefix))
			})

			It("Should not add the key twice", func() {
				nsMap := c.NewNameSpaceMap()
				nsMap.Add("dc", "http://purl.org/dc/elements/1.1/")
				nsMap.Add("dc", "http://purl.org/dc/elements/1.1/")
				Expect(nsMap).ToNot(BeNil())
				prefix2base, base2prefix := nsMap.Len()
				Expect(prefix2base).To(Equal(1))
				Expect(prefix2base).To(Equal(base2prefix))
			})
		})

		Context("When retrieving from the NameSpaceMap", func() {

			It("should return not ok when a key is not found", func() {
				nsMap := c.NewNameSpaceMap()
				base, ok := nsMap.GetBaseURI("dc")
				Expect(base).To(BeEmpty())
				Expect(ok).To(BeFalse())
				prefix, ok := nsMap.GetPrefix("http://purl.org/dc/elements/1.1/")
				Expect(prefix).To(BeEmpty())
				Expect(ok).To(BeFalse())
			})

			It("should return ok when the key is found", func() {
				nsMap := c.NewNameSpaceMap()
				nsMap.Add("dc", "http://purl.org/dc/elements/1.1/")
				base, ok := nsMap.GetBaseURI("dc")
				Expect(base).ToNot(BeEmpty())
				Expect(ok).To(BeTrue())
				prefix, ok := nsMap.GetPrefix("http://purl.org/dc/elements/1.1/")
				Expect(prefix).ToNot(BeEmpty())
				Expect(ok).To(BeTrue())
			})
		})

		Context("When deleting a key", func() {

			It("should remove the key from the prefix map", func() {
				nsMap := c.NewNameSpaceMap()
				nsMap.Add("dc", "http://purl.org/dc/elements/1.1/")
				prefix2base, base2prefix := nsMap.Len()
				Expect(prefix2base).To(Equal(1))
				Expect(prefix2base).To(Equal(base2prefix))
				nsMap.DeletePrefix("dc")
				prefix2base, base2prefix = nsMap.Len()
				Expect(prefix2base).To(Equal(0))
				Expect(prefix2base).To(Equal(base2prefix))
			})

			It("should remove the key from the prefix map", func() {
				nsMap := c.NewNameSpaceMap()
				nsMap.Add("dc", "http://purl.org/dc/elements/1.1/")
				prefix2base, base2prefix := nsMap.Len()
				Expect(prefix2base).To(Equal(1))
				Expect(prefix2base).To(Equal(base2prefix))
				nsMap.DeleteBaseURI("http://purl.org/dc/elements/1.1/")
				prefix2base, base2prefix = nsMap.Len()
				Expect(prefix2base).To(Equal(0))
				Expect(prefix2base).To(Equal(base2prefix))
			})
		})

		Context("When initialised from Config", func() {

			It("Should Contain the same number of entries as the config list", func() {

			})
		})

	})

	Describe("Should be able to deal with namespace from a uri", func() {

		Context("When given an URI as a string", func() {

			It("Should split when given an URI with a #", func() {
				rdfType := "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
				base, name := c.SplitURI(rdfType)
				Expect(name).To(Equal("type"))
				Expect(base).To(HaveSuffix("#"))
			})

			It("Should split when given an URI with a #", func() {
				dcSubject := "http://purl.org/dc/elements/1.1/subject"
				base, name := c.SplitURI(dcSubject)
				Expect(name).To(Equal("subject"))
				Expect(base).To(HaveSuffix("/"))
			})
		})

		Context("when given a URI", func() {

			nsMap := c.NewNameSpaceMap()
			nsMap.Add("dc", "http://purl.org/dc/elements/1.1/")

			It("should return the search label", func() {
				dcSubject := "http://purl.org/dc/elements/1.1/subject"
				label, err := nsMap.GetSearchLabel(dcSubject)
				Expect(err).ToNot(HaveOccurred())
				Expect(label).ToNot(BeEmpty())
				Expect(label).To(Equal("dc_subject"))
			})

			It("Add a default prefix when namespace is not found", func() {
				dcSubject := "http://purl.org/dc/elements/1.3/subject"
				label, err := nsMap.GetSearchLabel(dcSubject)
				Expect(err).ToNot(HaveOccurred())
				Expect(label).To(HaveSuffix("_subject"))
			})
		})

	})
})
