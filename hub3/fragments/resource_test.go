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
	"reflect"
	"testing"

	"github.com/delving/hub3/config"
	. "github.com/delving/hub3/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource", func() {

	Describe("when creating a resource map", func() {

		It("return an empty map when the graph is empty", func() {
			rm, err := NewResourceMap(r.NewGraph(""))
			Expect(err).To(HaveOccurred())
			Expect(rm.Resources()).To(BeEmpty())
		})

		It("return an non empty map when the graph is not empty", func() {
			fb, err := testDataGraph(false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fb).ToNot(BeNil())
			rm, err := NewResourceMap(fb.Graph)
			Expect(err).ToNot(HaveOccurred())
			Expect(rm).ToNot(BeNil())
			rs := rm.Resources()
			Expect(rs).ToNot(BeEmpty())
			Expect(rs).To(HaveLen(12))
			Expect(rs).To(HaveKey("http://data.jck.nl/resource/aggregation/jhm-foto/F900893"))
		})

		It("should have a FragmentResource for each map key", func() {
			fb, err := testDataGraph(false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fb).ToNot(BeNil())
			rm, err := NewResourceMap(fb.Graph)
			Expect(err).ToNot(HaveOccurred())
			Expect(rm.Resources()).ToNot(BeEmpty())

			subject := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893"
			fr, ok := rm.GetResource(subject)
			Expect(ok).To(BeTrue())
			Expect(fr.ID).To(Equal(subject))
			Expect(fr.Types).To(ContainElement("http://www.openarchives.org/ore/terms/Aggregation"))
			Expect(fr.Types).To(HaveLen(1))
			Expect(fr.ObjectIDs()).To(HaveLen(6))
			// todo properly check for not referring to itself
			Expect(fr.ObjectIDs()).ToNot(ContainElement(subject))
			Expect(fr.Predicates()).To(HaveLen(6))
		})
	})

	Describe("when appending a triple", func() {

		It("should add the subject to the resource map", func() {
			rm := NewEmptyResourceMap()
			Expect(rm.Resources()).To(BeEmpty())
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				NSRef("book"),
			)
			err := rm.AppendTriple(t, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(rm.Resources()).To(HaveLen(1))
			Expect(rm.Resources()).To(HaveKey(t.GetSubjectID()))
			fr, ok := rm.Resources()[t.GetSubjectID()]
			Expect(ok).To(BeTrue())
			Expect(fr.Types).To(HaveLen(1))
		})

		It("should add the subject only once", func() {
			rm := NewEmptyResourceMap()
			Expect(rm.Resources()).To(BeEmpty())
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				NSRef("book"),
			)
			err := rm.AppendTriple(t, false)
			Expect(err).ToNot(HaveOccurred())
			err = rm.AppendTriple(t, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(rm.Resources()).To(HaveLen(1))
		})

		It("should add not add objectIDS for rdfType", func() {
			rm := NewEmptyResourceMap()
			Expect(rm.Resources()).To(BeEmpty())
			subject := NSRef("1")
			t := r.NewTriple(
				subject,
				r.NewResource(RDFType),
				NSRef("book"),
			)
			err := rm.AppendTriple(t, false)
			Expect(err).ToNot(HaveOccurred())

			entry, ok := rm.Resources()[r.GetResourceID(subject)]
			Expect(ok).To(BeTrue())
			Expect(entry.ObjectIDs()).To(HaveLen(0))
		})

		It("should add objectIDS for resources", func() {
			rm := NewEmptyResourceMap()
			Expect(rm.Resources()).To(BeEmpty())
			subject := NSRef("1")
			t := r.NewTriple(
				subject,
				NSRef("title"),
				NSRef("myBook"),
			)
			err := rm.AppendTriple(t, false)
			Expect(err).ToNot(HaveOccurred())

			err = rm.AppendTriple(t, false)
			Expect(err).ToNot(HaveOccurred())

			entry, ok := rm.Resources()[r.GetResourceID(subject)]
			Expect(ok).To(BeTrue())
			Expect(entry.ObjectIDs()).To(HaveLen(1))
		})

	})

	Describe("when creating a fragment entry", func() {

		It("should return an ID for a resource", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				NSRef("book"),
			)
			entry, id := CreateFragmentEntry(t, false, 0)
			Expect(id).ToNot(BeEmpty())
			Expect(id).To(Equal(r.GetResourceID(t.Object)))
			Expect(entry.ID).To(Equal(id))
			Expect(entry.Triple).ToNot(BeEmpty())
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.DataType).To(BeEmpty())
			Expect(entry.Value).To(BeEmpty())
			Expect(entry.EntryType).To(Equal("Resource"))
		})

		It("should return an ID for a BlankNode", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				r.NewBlankNode("book"),
			)
			entry, id := CreateFragmentEntry(t, false, 0)
			Expect(id).ToNot(BeEmpty())
			Expect(id).To(Equal(r.GetResourceID(t.Object)))
			Expect(id).To(HavePrefix("_:"))
			Expect(id).To(Equal("_:book"))
			Expect(entry.ID).To(Equal(id))
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.DataType).To(BeEmpty())
			Expect(entry.Value).To(BeEmpty())
			Expect(entry.EntryType).To(Equal("Bnode"))
		})

		It("should return no ID for a Literal", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				r.NewLiteral("book"),
			)
			entry, id := CreateFragmentEntry(t, false, 0)
			Expect(id).To(BeEmpty())
			Expect(entry.ID).To(BeEmpty())

			Expect(entry.Value).To(Equal("book"))
			Expect(entry.DataType).To(BeEmpty())
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.EntryType).To(Equal("Literal"))
		})

		It("should have a language when the triple has a language", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				r.NewLiteralWithLanguage("book", "en"),
			)
			entry, id := CreateFragmentEntry(t, false, 0)
			Expect(id).To(BeEmpty())
			Expect(entry.ID).To(BeEmpty())

			Expect(entry.Value).To(Equal("book"))
			Expect(entry.DataType).To(BeEmpty())
			Expect(entry.Language).To(Equal("en"))
			Expect(entry.EntryType).To(Equal("Literal"))
		})

		It("should have a datatype for non-string", func() {
			t := r.NewTriple(
				NSRef("1"),
				r.NewResource(RDFType),
				r.NewLiteralWithDatatype("1", r.NewResource("http://www.w3.org/2001/XMLSchema#decimal")),
			)
			entry, id := CreateFragmentEntry(t, false, 0)
			Expect(id).To(BeEmpty())
			Expect(entry.ID).To(BeEmpty())

			Expect(entry.Value).To(Equal("1"))
			Expect(entry.DataType).ToNot(BeEmpty())
			Expect(entry.Language).To(BeEmpty())
			Expect(entry.EntryType).To(Equal("Literal"))
		})
	})

	Describe("when creating FragmentReferrerContext", func() {

		Context("and determining the level", func() {
			fb, _ := testDataGraph(false)
			rm, _ := NewResourceMap(fb.Graph)
			subject := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893"

			It("should not have 0 as level", func() {
				fb, err := testDataGraph(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(fb).ToNot(BeNil())
				rm, err := NewResourceMap(fb.Graph)
				Expect(err).ToNot(HaveOccurred())
				Expect(rm.Resources()).ToNot(BeEmpty())

				fr, ok := rm.GetResource(subject)
				Expect(ok).To(BeTrue())

				level := fr.GetLevel()
				Expect(level).To(Equal(int32(1)))
			})
			It("should throw an error when the subject is unknown", func() {
				Expect(rm).ToNot(BeNil())
				_, err := rm.SetContextLevels("urn:unknown")
				Expect(err).To(HaveOccurred())
			})

			config.InitConfig()
			It("should determine its level by the number of context is has", func() {
				Expect(rm).ToNot(BeNil())
				_, err := rm.SetContextLevels(subject)
				Expect(err).ToNot(HaveOccurred())

				providedCHO, ok := rm.GetResource("http://data.jck.nl/resource/document/jhm-foto/F900893")
				Expect(providedCHO).ToNot(BeNil())
				Expect(ok).To(BeTrue())
				Expect(providedCHO.Context).To(HaveLen(1))
				Expect(providedCHO.Context[0].GetSubjectClass()).To(HaveLen(1))
				Expect(providedCHO.Context[0].Level).To(Equal(int32(1)))
				Expect(providedCHO.GetLevel()).To(Equal(int32(2)))
				label, lang := providedCHO.GetLabel()
				Expect(label).To(Equal(""))
				Expect(lang).To(Equal(""))

				skosConcept, ok := rm.GetResource("http://data.jck.nl/resource/skos/thesau/90000072")
				Expect(skosConcept).ToNot(BeNil())
				Expect(ok).To(BeTrue())
				Expect(skosConcept.Context).To(HaveLen(2))
				Expect(skosConcept.GetLevel()).To(Equal(int32(3)))
				Expect(skosConcept.Context[1].Level).To(Equal(int32(2)))
				Expect(skosConcept.Context[1].GetSubjectClass()).To(HaveLen(1))
				Expect(skosConcept.Context[0].Level).To(Equal(int32(1)))
				Expect(skosConcept.Context[0].GetSubjectClass()).To(HaveLen(1))
				Expect(config.Config.RDFTag.Label).To(HaveLen(2))
				label, lang = skosConcept.GetLabel()
				Expect(label).To(Equal("grafsteen"))
				Expect(lang).To(Equal("nl"))
			})
		})
	})

	Describe("when creating a ResultSummary", func() {

		Context("from a resource entry", func() {

			It("should only set a field once", func() {
				entry1 := &ResourceEntry{
					Value: "test1",
					Tags:  []string{"title"},
				}
				entry2 := &ResourceEntry{
					Value: "test2",
					Tags:  []string{"title"},
				}
				sum := &ResultSummary{}
				Expect(sum.Title).To(BeEmpty())
				Expect(sum.Thumbnail).To(BeEmpty())
				sum.AddEntry(entry1)
				Expect(sum.Title).To(Equal("test1"))
				sum.AddEntry(entry2)
				Expect(sum.Thumbnail).To(BeEmpty())
				Expect(sum.Title).To(Equal("test1"))
				Expect(sum.Thumbnail).To(BeEmpty())
			})
		})
	})

	Describe("when creating a Header", func() {

		fb, _ := testDataGraph(false)
		//rm, _ := NewResourceMap(fb.Graph)
		//subject := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893"

		Context("from a FragmentGraph", func() {

			//skosConcept, _ := rm.GetResource("http://data.jck.nl/resource/skos/thesau/90000072")
			//entry := skosConcept.Predicates["http://www.w3.org/2004/02/skos/core#prefLabel"]
			header := fb.FragmentGraph().CreateHeader("fragment")

			It("should set the OrgID", func() {
				Expect(header.OrgID).To(Equal("hub3"))
			})

			It("should set the spec", func() {
				Expect(header.Spec).To(Equal("test-spec"))
			})

			It("should set the Revision", func() {
				Expect(header.Revision).To(Equal(int32(1)))
			})

			It("should set the hubID", func() {
				Expect(header.HubID).ToNot(Equal(""))
			})

			It("should have no tags", func() {
				Expect(header.Tags).To(BeEmpty())
			})

			It("should have a docType", func() {
				Expect(header.GetDocType()).To(Equal("fragment"))
			})

		})

		Context("and adding Tags", func() {

			It("should only add a tag", func() {
				header := fb.FragmentGraph().CreateHeader("")
				Expect(header.Tags).To(BeEmpty())
				header.AddTags("tag1")
				Expect(header.Tags).ToNot(BeEmpty())
				Expect(header.Tags).To(HaveLen(1))

			})

			It("should not add a tag twice", func() {
				header := fb.FragmentGraph().CreateHeader("")
				Expect(header.Tags).To(BeEmpty())
				header.AddTags("tag1", "tag2")
				header.AddTags("tag1")
				Expect(header.Tags).ToNot(BeEmpty())
				Expect(header.Tags).To(HaveLen(2))

			})

		})

	})

})

func TestCreateDateRange(t *testing.T) {
	type args struct {
		period string
	}
	tests := []struct {
		name    string
		args    args
		want    IndexRange
		wantErr bool
	}{
		{"simple year",
			args{period: "1980"},
			IndexRange{
				Greater: "1980-01-01",
				Less:    "1980-12-31",
			},
			false,
		},
		{"/ period",
			args{period: "1980/1985"},
			IndexRange{
				Greater: "1980-01-01",
				Less:    "1985-12-31",
			},
			false,
		},
		{"padded / period",
			args{period: "1980 / 1985"},
			IndexRange{
				Greater: "1980-01-01",
				Less:    "1985-12-31",
			},
			false,
		},
		{"full year period",
			args{period: "1793-05-13/1794-01-11"},
			IndexRange{
				Greater: "1793-05-13",
				Less:    "1794-01-11",
			},
			false,
		},
		{"mixed years",
			args{period: "1793-05/1794"},
			IndexRange{
				Greater: "1793-05-01",
				Less:    "1794-12-31",
			},
			false,
		},
		{"feb year",
			args{period: "1778/1781-02"},
			IndexRange{
				Greater: "1778-01-01",
				Less:    "1781-02-28",
			},
			false,
		},
		//{"- period",
		//args{period: "1980-1985"},
		//IndexRange{
		//Greater: "1980-01-01",
		//Less:    "1985-12-31",
		//},
		//false,
		//},
		//{"padded - period",
		//args{period: "1980 - 1985"},
		//IndexRange{
		//Greater: "1980-01-01",
		//Less:    "1985-12-31",
		//},
		//false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateDateRange(tt.args.period)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDateRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateDateRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndexRange_Valid(t *testing.T) {
	type fields struct {
		Greater string
		Less    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"simple",
			fields{
				Less:    "zzz",
				Greater: "aaa",
			},
			false,
		},
		{
			"simple reverse",
			fields{
				Less:    "aaa",
				Greater: "zzz",
			},
			true,
		},
		{
			"date correct",
			fields{
				Less:    "1900-10-11",
				Greater: "1850-01-01",
			},
			false,
		},
		{
			"date incorrect",
			fields{
				Greater: "1900-10-11",
				Less:    "1850-01-01",
			},
			true,
		},
		{
			"date partial",
			fields{
				Less:    "1900-10-11",
				Greater: "1850",
			},
			false,
		},
		{
			"numeric",
			fields{
				Less:    "1000",
				Greater: "1",
			},
			false,
		},
		{
			"partial less empty",
			fields{
				Less:    "",
				Greater: "100",
			},
			true,
		},
		{
			"partial greater empty",
			fields{
				Less:    "100",
				Greater: "",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := IndexRange{
				Greater: tt.fields.Greater,
				Less:    tt.fields.Less,
			}
			if err := ir.Valid(); (err != nil) != tt.wantErr {
				t.Errorf("IndexRange.Valid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTreeQuery_IsExpanded(t *testing.T) {
	type fields struct {
		Label    string
		UnitID   string
		IsPaging bool
		Query    string
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"not expanded",
			fields{},
			false,
		},
		{
			"is paging",
			fields{
				IsPaging: true,
			},
			true,
		},
		{
			"has Label",
			fields{
				Label: "tree",
			},
			true,
		},
		{
			"has UnitID",
			fields{
				UnitID: "123",
			},
			true,
		},
		{
			"has query",
			fields{
				Query: "my query",
			},
			true,
		},
		{
			"has query with paging",
			fields{
				Query:    "my query",
				IsPaging: true,
			},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tq := &TreeQuery{
				Label:    tt.fields.Label,
				UnitID:   tt.fields.UnitID,
				IsPaging: tt.fields.IsPaging,
				Query:    tt.fields.Query,
			}
			if got := tq.IsExpanded(); got != tt.want {
				t.Errorf("TreeQuery.IsExpanded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTreeQuery_IsNavigatedQuery(t *testing.T) {
	type fields struct {
		CLevel           string
		Leaf             string
		Parent           string
		Type             []string
		Depth            []string
		FillTree         bool
		ChildCount       string
		Label            string
		Spec             string
		UnitID           string
		CursorHint       int32
		MimeType         []string
		HasRestriction   bool
		HasDigitalObject bool
		Page             []int32
		PageSize         int32
		AllParents       bool
		IsPaging         bool
		IsSearch         bool
		PageMode         string
		Query            string
		WithFields       bool
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{

		{
			"not a navigated query",
			fields{},
			false,
		},
		{
			"only unitID is not a navigated query",
			fields{UnitID: "123"},
			false,
		},
		{
			"UnitID && Label is a navigated query",
			fields{
				UnitID: "123",
				Label:  "my label",
			},
			true,
		},
		{
			"UnitID && Query is a navigated query",
			fields{
				UnitID: "123",
				Query:  "navigated query",
			},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tq := &TreeQuery{
				CLevel:           tt.fields.CLevel,
				Leaf:             tt.fields.Leaf,
				Parent:           tt.fields.Parent,
				Type:             tt.fields.Type,
				Depth:            tt.fields.Depth,
				FillTree:         tt.fields.FillTree,
				ChildCount:       tt.fields.ChildCount,
				Label:            tt.fields.Label,
				Spec:             tt.fields.Spec,
				UnitID:           tt.fields.UnitID,
				CursorHint:       tt.fields.CursorHint,
				MimeType:         tt.fields.MimeType,
				HasRestriction:   tt.fields.HasRestriction,
				HasDigitalObject: tt.fields.HasDigitalObject,
				Page:             tt.fields.Page,
				PageSize:         tt.fields.PageSize,
				AllParents:       tt.fields.AllParents,
				IsPaging:         tt.fields.IsPaging,
				IsSearch:         tt.fields.IsSearch,
				PageMode:         tt.fields.PageMode,
				Query:            tt.fields.Query,
				WithFields:       tt.fields.WithFields,
			}
			if got := tq.IsNavigatedQuery(); got != tt.want {
				t.Errorf("TreeQuery.IsNavigatedQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
