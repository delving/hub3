package models

import (
	"time"

	"bitbucket.org/delving/rapid/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dataset", func() {

	Context("When creating a dataset URI", func() {

		config.InitConfig()
		uri := createDatasetURI("test")
		It("should end with the spec", func() {
			Expect(uri).To(HaveSuffix("test"))
		})

		It("should contain the resource path and type", func() {
			Expect(uri).To(ContainSubstring("/resource/dataset/"))
		})

		It("should start with the RDF baseUrl from the configuration.", func() {
			baseUrl := config.Config.RDF.BaseUrl
			Expect(uri).To(HavePrefix(baseUrl))
		})

	})

	Context("When creating a new Dataset", func() {
		spec := "test"
		dataset := NewDataset(spec)
		It("should set the spec", func() {
			Expect(dataset).ToNot(BeNil())
			Expect(dataset.Spec).To(Equal(spec))
		})

		It("should set a datasetUri", func() {
			uri := dataset.Uri
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

	})

})
