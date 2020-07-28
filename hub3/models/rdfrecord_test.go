// Copyright © 2017 Delving B.V. <info@delving.eu>
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

package models

import (
	"github.com/delving/hub3/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RDFRecord", func() {

	hubID := "test_spec_123"
	spec := "spec"
	Context("When creating a new RDFRecord", func() {
		It("should not be empty", func() {
			record := NewRDFRecord(
				hubID,
				spec,
			)
			Expect(record.HubID).ToNot(BeEmpty())
		})
	})

	Context("when saving an RDFRecord", func() {
		It("should store the record in BoltDB", func() {
			record := NewRDFRecord(hubID, spec)
			err := record.Save()
			Expect(err).ToNot(HaveOccurred())
			var response RDFRecord
			err = orm.One("HubID", record.HubID, &response)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should have the count of 1", func() {
			count := CountRDFRecords(spec)
			Expect(count).To(Equal(1))
		})
	})

	Context("Given an HubID", func() {
		record := RDFRecord{
			HubID: hubID,
		}
		orgID, spec, localID, err := record.ExtractHubID()

		It("should provide access to the localID", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(localID).To(Equal("123"))
		})

		It("should provide access to the OrgID", func() {
			Expect(orgID).To(Equal("test"))
		})

		It("should provide access to the spec", func() {
			Expect(spec).To(Equal("spec"))
		})

	})

	Context("Given an illegal HubID", func() {
		record := RDFRecord{
			HubID: "testspec_123",
		}
		orgID, _, _, err := record.ExtractHubID()

		It("should return an error", func() {
			Expect(orgID).To(BeEmpty())
			Expect(err).To(HaveOccurred())
		})

	})

	Context("When creating a source URI", func() {

		config.InitConfig()
		record := RDFRecord{
			HubID: "test_spec_123",
		}
		uri := record.createSourceURI()
		It("should start with the baseURI", func() {
			Expect(uri).ToNot(BeEmpty())
			Expect(uri).To(HavePrefix(config.Config.RDF.BaseURL))
		})

		It("should end with the localId", func() {
			Expect(uri).To(ContainSubstring("123"))
			Expect(uri).To(HaveSuffix("/123"))
		})

		It("should include the spec", func() {
			Expect(uri).To(ContainSubstring("/spec/"))
		})

	})

})
