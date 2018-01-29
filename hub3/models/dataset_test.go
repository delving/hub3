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
	"context"
	"time"

	c "bitbucket.org/delving/rapid/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dataset", func() {

	Context("When creating a dataset URI", func() {

		c.InitConfig()
		c.Config.RDF.RDFStoreEnabled = false
		c.Config.ElasticSearch.Enabled = false

		uri := createDatasetURI("test")
		It("should end with the spec", func() {
			Expect(uri).To(HaveSuffix("test"))
		})

		It("should contain the resource path and type", func() {
			Expect(uri).To(ContainSubstring("/resource/dataset/"))
		})

		It("should start with the RDF baseUrl from the configuration.", func() {
			baseURL := c.Config.RDF.BaseURL
			Expect(uri).To(HavePrefix(baseURL))
		})

	})

	Context("When creating a new Dataset", func() {
		spec := "test_spec"
		dataset := NewDataset(spec)
		It("should set the spec", func() {
			Expect(dataset).ToNot(BeNil())
			Expect(dataset.Spec).To(Equal(spec))
		})

		It("should set a datasetUri", func() {
			uri := dataset.URI
			Expect(uri).ToNot(BeEmpty())
			Expect(uri).To(Equal(createDatasetURI(spec)))
		})

		It("should set the creation time", func() {
			created := dataset.Created
			Expect(created).ToNot(BeNil())
			Expect(created.Day()).To(Equal(time.Now().Day()))
			Expect(created.Month()).To(Equal(time.Now().Month()))
			Expect(created.Year()).To(Equal(time.Now().Year()))
		})

		It("the creationd and modification time should be the same", func() {
			Expect(dataset.Created).To(Equal(dataset.Modified))
		})

		It("should set the revision to zero", func() {
			Expect(dataset.Revision).To(Equal(0))
		})

		It("should set deleted to be false", func() {
			Expect(dataset.Deleted).To(BeFalse())
		})

		It("should have access set to true", func() {
			Expect(dataset.Access.OAIPMH).To(BeTrue())
			Expect(dataset.Access.Search).To(BeTrue())
			Expect(dataset.Access.LOD).To(BeTrue())
		})

	})

	Context("When saving a DataSet", func() {
		spec := "test_spec"
		dataset := NewDataset(spec)

		It("should have nothing saved before save", func() {
			var ds []DataSet
			err := orm.All(&ds)
			Expect(err).To(BeNil())
			Expect(len(ds)).To(Equal(0))
		})

		It("should save a dataset without errors", func() {
			Expect(dataset.Save()).To(BeNil())
			var ds []DataSet
			err := orm.All(&ds)
			Expect(err).To(BeNil())
			Expect(len(ds)).To(Equal(1))
		})

		It("should be able to find it in the database", func() {
			var ds DataSet
			err := orm.One("Spec", spec, &ds)
			Expect(err).To(BeNil())
			Expect(ds.Created.Unix()).To(Equal(dataset.Created.Unix()))
			Expect(ds.Modified.UnixNano()).ToNot(Equal(dataset.Modified.UnixNano()))
			Expect(ds.Access.LOD).To(BeTrue())
		})

	})

	Context("When calling CreateDataSet", func() {

		It("should create a dataset when no dataset is present.", func() {
			ds, err := CreateDataSet("test3")
			Expect(err).ToNot(HaveOccurred())
			Expect(ds.Spec).To(Equal("test3"))
		})

	})

	Context("When calling GetOrCreateDataSet", func() {

		It("should create the datasets when no dataset is available", func() {
			ds, err := GetOrCreateDataSet("test2")
			Expect(err).ToNot(HaveOccurred())
			Expect(ds.Spec).To(Equal("test2"))
		})

		It("should not store the dataset again on Get", func() {
			datasetCount, err := ListDataSets()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(datasetCount) > 0).To(BeTrue())
			ds, err := GetOrCreateDataSet("test2")
			Expect(err).ToNot(HaveOccurred())
			Expect(ds.Spec).To(Equal("test2"))
			newCount, err := ListDataSets()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(datasetCount)).To(Equal(len(newCount)))
		})
	})

	Context("When calling ListDatasets", func() {

		It("should return an array of all stored datasets", func() {
			datasets, err := ListDataSets()
			Expect(err).ToNot(HaveOccurred())
			Expect(datasets).ToNot(BeEmpty())
		})

	})

	Context("When calling IncrementRevision", func() {

		It("should update the revision of the dataset by one", func() {
			ds, _ := GetOrCreateDataSet("test3")
			Expect(ds.Revision).To(Equal(0))
			err := ds.IncrementRevision()
			Expect(err).ToNot(HaveOccurred())
			ds, _ = GetOrCreateDataSet("test3")
			Expect(ds.Revision).To(Equal(1))
		})

		It("should have stored the dataset with the new revision", func() {
			ds, _ := GetOrCreateDataSet("test3")
			Expect(ds.Revision).To(Equal(1))
		})
	})

	// todo add code for removing datasets.
	Context("When calling delete", func() {

		It("should delete the dataset", func() {
			dataSets, err := ListDataSets()
			Expect(err).To(BeNil())
			dsNr := len(dataSets)
			dsName := "test4"
			ds, _ := GetOrCreateDataSet(dsName)
			Expect(ds).ToNot(BeNil())
			ds, err = GetDataSet(dsName)
			Expect(err).To(BeNil())
			ctx := context.Background()
			err = ds.Delete(ctx)
			Expect(err).To(BeNil())
			ds, err = GetDataSet(dsName)
			Expect(err).ToNot(BeNil())
			dataSets, err = ListDataSets()
			Expect(err).To(BeNil())
			Expect(dsNr).To(Equal(len(dataSets)))
		})

	})

})
