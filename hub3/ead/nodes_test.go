package ead_test

import (
	"context"
	"testing"

	"github.com/matryer/is"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/delving/hub3/config"
	. "github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/hub3/fragments"
)

var _ = Describe("Nodes", func() {

	config.InitConfig()

	Describe("converting to RDF", func() {

		Context("from a header", func() {
			dsc := new(Cdsc)
			err := parseUtil(dsc, "ead.1.xml")
			var nl *NodeList
			cfg := NewNodeConfig(context.Background())
			cfg.OrgID = "test"
			cfg.Spec = "test-spec"
			cfg.Revision = int32(38)
			cfg.CreateTree = CreateTree
			var h *fragments.Header

			It("should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
				nl, _, err = dsc.NewNodeList(cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(nl).ToNot(BeNil())
				fg, _, _ := nl.Nodes[0].FragmentGraph(cfg)
				h = fg.Meta
				Expect(h).ToNot(BeNil())
			})

			It("should have an OrgID", func() {
				Expect(h.GetOrgID()).To(Equal("test"))
			})

			It("should have a spec", func() {
				Expect(h.GetSpec()).To(Equal("test-spec"))
			})

			It("should have the right revision", func() {
				Expect(h.GetRevision()).To(Equal(int32(38)))
			})

			It("should have a hubID", func() {
				Expect(h.GetHubID()).To(Equal("test_test-spec_A"))
			})

			It("should have an EAD docType", func() {
				Expect(h.GetDocType()).To(Equal("graph"))
			})

			It("should set a modified time", func() {
				Expect(h.GetModified()).ToNot(BeZero())
			})

			It("should set the entryURI", func() {
				Expect(h.GetEntryURI()).ToNot(BeEmpty())
				Expect(h.GetEntryURI()).To(Equal("http://data.hub3.org/hub3/archive/test-spec/A"))
			})

			It("should have a NamedGraphURI", func() {
				Expect(h.GetNamedGraphURI()).To(HavePrefix(h.GetEntryURI()))
				Expect(h.GetNamedGraphURI()).To(HaveSuffix("/graph"))
			})
		})

		Context("from a single cLevel node", func() {
			dsc := new(Cdsc)
			err := parseUtil(dsc, "ead.1.xml")
			var nl *NodeList
			cfg := NewNodeConfig(context.Background())
			cfg.OrgID = "test"
			cfg.Spec = "test_spec"
			cfg.Revision = int32(38)
			cfg.CreateTree = CreateTree

			It("should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
				nl, _, err = dsc.NewNodeList(cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(nl).ToNot(BeNil())
			})

			It("should convert only the main body to RDF", func() {
				node := nl.Nodes[0]
				Expect(node.Type).To(Equal("series"))
				fr, _, err := node.FragmentGraph(cfg)
				Expect(err).ToNot(HaveOccurred())
				s := fr.GetAboutURI()
				Expect(s).To(Equal("http://data.hub3.org/hub3/archive/test_spec/A"))
			})

			It("should set the meta header", func() {
				node := nl.Nodes[0]
				fr, _, err := node.FragmentGraph(cfg)
				Expect(err).ToNot(HaveOccurred())
				h := fr.Meta
				Expect(h).ToNot(BeNil())
			})

			It("should have resources", func() {
				node := nl.Nodes[0]
				fr, _, err := node.FragmentGraph(cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(fr).ToNot(BeNil())
				// TODO enable later again
				//Expect(fr.Resources).ToNot(BeNil())
				//Expect(fr.Resources).To(HaveLen(1))
			})

		})

		Context("when creating triples with parents", func() {
			cc := new(Cc)
			err := parseUtil(cc, "ead.4.xml")
			cfg := NewNodeConfig(context.Background())
			cfg.AddLabel("c1", "c1 label")
			cfg.AddLabel("c2", "c2 label")
			cfg.OrgID = "test"
			cfg.Spec = "test_spec"
			cfg.Revision = int32(38)
			var node *Node

			It("it should not throw an error", func() {
				parentIDs := []string{"c1", "c2"}
				node, err = NewNode(cc, parentIDs, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(node).ToNot(BeNil())
			})

			It("should have parentIDS", func() {
				triples := node.Triples(cfg)
				Expect(triples).ToNot(BeEmpty())
				Expect(triples).To(HaveLen(11))
			})

		})

		Context("when creating triples from a Header", func() {
			cc := new(Cc)
			err := parseUtil(cc, "ead.4.xml")
			cfg := NewNodeConfig(context.Background())
			cfg.OrgID = "test"
			cfg.Spec = "test_spec"
			cfg.Revision = int32(38)
			var h *Header
			var node *Node

			It("it should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
				parentIDs := []string{"c1", "c2"}
				node, err = NewNode(cc, parentIDs, cfg)
				Expect(err).ToNot(HaveOccurred())

				h = node.Header
				Expect(h).ToNot(BeNil())
			})

			It("should generate a subject", func() {
				s := node.GetSubject(cfg)
				Expect(s).ToNot(BeEmpty())
			})

			Context("check unit date", func() {
				var date *NodeDate

				It("should have a date", func() {
					dates := h.Date
					Expect(dates).ToNot(BeEmpty())
					Expect(dates).To(HaveLen(1))
					date = dates[0]
					Expect(date).ToNot(BeNil())
				})

			})

			Context("check unitIDs", func() {
				var id *NodeID

				It("should have unitIDs", func() {
					ids := h.ID
					Expect(ids).ToNot(BeEmpty())
					Expect(ids).To(HaveLen(2))
					id = ids[0]
					Expect(id).ToNot(BeNil())
				})

			})

		})
	})

})

// nolint:gocritic
func TestParseNonNumbered(t *testing.T) {
	is := is.New(t)

	ead := new(Cead)
	err := parseUtil(ead, "4.ZHPB2.xml")
	is.NoErr(err)

	dsc := ead.Carchdesc.Cdsc

	cfg := NewNodeConfig(context.Background())

	nl, processed, err := dsc.NewNodeList(cfg)
	is.NoErr(err)

	t.Logf("processed %d", processed)
	is.True(processed == uint64(641))
	is.True(nl != nil)
}
