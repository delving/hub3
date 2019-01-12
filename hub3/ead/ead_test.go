package ead_test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "github.com/delving/rapid-saas/hub3/ead"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ead", func() {

	Describe("when creating a node configuration", func() {

		It("should initialise the NodeCounter", func() {
			cfg := NewNodeConfig(context.Background())
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Counter).ToNot(BeNil())
		})

		It("should increment the counter by one", func() {
			cfg := NewNodeConfig(context.Background())
			Expect(cfg.Counter.GetCount()).To(BeZero())
			cfg.Counter.Increment()
			Expect(cfg.Counter.GetCount()).ToNot(BeZero())
			Expect(cfg.Counter.GetCount()).To(Equal(uint64(1)))
		})

	})

	Describe("when converting to Nodes", func() {

		Context("for the dsc", func() {
			dsc := new(Cdsc)
			err := parseUtil(dsc, "ead.1.xml")

			It("should have a type", func() {
				Expect(err).ToNot(HaveOccurred())
				cfg := NewNodeConfig(context.Background())
				nl, seen, err := dsc.NewNodeList(cfg)
				Expect(seen).To(Equal(uint64(1)))
				Expect(err).ToNot(HaveOccurred())
				Expect(nl.GetType()).To(Equal("combined"))
				//Expect(nl.GetNodes()).ToNot(HaveLen(1))
				Expect(nl.GetLabel()).To(HaveLen(1))
				Expect(dsc.Chead[0].Head).To(HavePrefix("Beschrijving"))
				Expect(nl.GetLabel()[0]).To(HavePrefix("Beschrijving"))
			})

			It("should have a header", func() {
				Expect(err).ToNot(HaveOccurred())
				cfg := NewNodeConfig(context.Background())
				nl, seen, err := dsc.NewNodeList(cfg)
				Expect(seen).To(Equal(uint64(1)))
				Expect(err).ToNot(HaveOccurred())
				Expect(nl.GetLabel()).To(HaveLen(1))
				sourceHeader := dsc.Chead[0].Head
				Expect(nl.GetLabel()[0]).To(HavePrefix("Beschrijving"))
				Expect(nl.GetLabel()[0]).To(Equal(sourceHeader))
			})

			It("should have c-levels", func() {
				Expect(dsc.Nested).To(HaveLen(1))
				cfg := NewNodeConfig(context.Background())
				nl, seen, err := dsc.NewNodeList(cfg)
				Expect(seen).To(Equal(uint64(1)))
				Expect(err).ToNot(HaveOccurred())
				Expect(nl.GetNodes()).To(HaveLen(1))
				// TODO add nil check
			})
		})

		Context("for the c01", func() {
			c01 := new(Cc01)
			err := parseUtil(c01, "ead.2.xml")
			cfg := NewNodeConfig(context.Background())
			var node *Node

			It("should not throw an error on create", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(c01.GetXMLName().Local).To(Equal("c01"))
				node, err = NewNode(c01, []string{}, 0, cfg)
				Expect(node.Order).To(Equal(uint64(1)))
				Expect(err).ToNot(HaveOccurred())
			})

			It("should have a cTag", func() {
				Expect(node).ToNot(BeNil())
				Expect(node.GetCTag()).ToNot(BeEmpty())
				Expect(node.GetCTag()).To(Equal("c01"))
				Expect(node.GetDepth()).To(Equal(int32(1)))
			})

			It("should set the depth", func() {
				Expect(node.GetDepth()).To(Equal(int32(1)))
			})

			It("should have a type", func() {
				Expect(c01.GetAttrlevel()).To(Equal("series"))
				//Expect(node.GetType()).To(Equal("series"))
			})

			It("should have a subType", func() {
				Expect(node.GetSubType()).To(BeEmpty())
			})
		})

		Context("for the complex date did", func() {
			did := new(Cdid)
			err := parseUtil(did, "ead.diddate2.xml")
			var header *Header

			It("should not throw an error on create", func() {
				Expect(err).ToNot(HaveOccurred())
				header, err = did.NewHeader()
				Expect(err).ToNot(HaveOccurred())
				Expect(header).ToNot(BeNil())
			})

			It("should have date as label", func() {
				Expect(header.GetLabel()).To(HaveLen(1))
				Expect(header.GetLabel()[0]).To(Equal("Octrooi verleend door de Staten-Generaal betreffende de alleenhandel ten oosten van Kaap de Goede Hoop en ten westen van de Straat van Magallanes voor de duur van 21 jaar"))
				Expect(header.GetTreeLabel()).To(Equal("1 Octrooi verleend door de Staten-Generaal betreffende de alleenhandel ten oosten van Kaap de Goede Hoop en ten westen van de Straat van Magallanes voor de duur van 21 jaar"))
			})

			It("should not have date as label", func() {
				Expect(header.GetDateAsLabel()).To(BeFalse())
			})

		})

		Context("for the date did", func() {
			did := new(Cdid)
			err := parseUtil(did, "ead.diddate.xml")
			var header *Header

			It("should not throw an error on create", func() {
				Expect(err).ToNot(HaveOccurred())
				header, err = did.NewHeader()
				Expect(err).ToNot(HaveOccurred())
				Expect(header).ToNot(BeNil())
			})

			It("should have date as label", func() {
				Expect(header.GetLabel()).To(HaveLen(1))
				Expect(header.GetLabel()[0]).To(Equal("ca. 1839 new books."))
			})

			It("should have date as label", func() {
				Expect(header.GetDateAsLabel()).To(BeTrue())
			})
		})

		Context("for the did", func() {
			did := new(Cdid)
			err := parseUtil(did, "ead.did.xml")
			var header *Header

			It("should not throw an error on create", func() {
				Expect(err).ToNot(HaveOccurred())
				header, err = did.NewHeader()
				Expect(err).ToNot(HaveOccurred())
				Expect(header).ToNot(BeNil())
			})

			It("should have a physdesc", func() {
				Expect(header.GetPhysdesc()).ToNot(BeEmpty())
			})

			It("should have not have date as label", func() {
				Expect(header.GetDateAsLabel()).To(BeFalse())
			})

			It("should have a label", func() {
				Expect(header.GetLabel()).ToNot(BeEmpty())
			})

			It("should have a date", func() {
				Expect(header.GetDate()).ToNot(BeEmpty())
			})

			Context("when extracting a NodeDate", func() {

				unitDate := did.Cunitdate[0]
				nDate, err := unitDate.NewNodeDate()

				It("should not thrown an error on creation", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(nDate).ToNot(BeNil())
				})

				It("should have an calendar", func() {
					Expect(nDate.GetCalendar()).To(Equal(unitDate.Attrcalendar))
				})

				It("should have an era", func() {
					Expect(nDate.GetEra()).To(Equal(unitDate.Attrera))
				})

				It("should have a normal string", func() {
					Expect(nDate.GetNormal()).To(Equal(unitDate.Attrnormal))
				})

				It("should have the date as string", func() {
					Expect(nDate.GetLabel()).To(Equal(unitDate.Date))
				})

			})

			Context("when extracting a handle unitID", func() {

				unitID := did.Cunitid[1]
				id, err := unitID.NewNodeID()

				It("should have an ID", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(id).ToNot(BeNil())
					Expect(id.ID).To(Equal(unitID.ID))
				})

				It("should have a Type", func() {
					Expect(id.GetType()).To(Equal(unitID.Attrtype))
				})

				It("should have an audience", func() {
					Expect(id.GetAudience()).To(Equal(unitID.Attraudience))
				})

			})

			Context("when extracting from an ABS unitID", func() {

				unitID := did.Cunitid[0]
				id, err := unitID.NewNodeID()

				It("should have an ID", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(id).ToNot(BeNil())
					Expect(id.ID).To(Equal(unitID.ID))
				})

				It("should have a TypeID", func() {
					Expect(id.GetTypeID()).To(Equal(unitID.Attridentifier))
				})

				It("should have a Type", func() {
					Expect(id.GetType()).To(Equal(unitID.Attrtype))
				})

				It("should extract the nodeIDs", func() {
					nodeIDs, inventoryNumber, err := did.NewNodeIDs()
					Expect(err).ToNot(HaveOccurred())
					Expect(nodeIDs).ToNot(BeEmpty())
					Expect(inventoryNumber).ToNot(BeEmpty())
				})
			})

		})

	})

	Context("when being load from file", func() {

		It("should create an EAD object", func() {
			path, err := filepath.Abs("./test_data/ead/NL-HaNA_2.08.22.ead.xml")
			//Expect(err).ToNot(HaveOccurred())
			//rawEAD, err := ioutil.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			//Expect(rawEAD).ToNot(BeNil())
			ead, err := ReadEAD(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(ead).ToNot(BeNil())
			fmt.Printf("%#v\n", ead.Ceadheader.Ceadid.EadID)
			//fmt.Printf("%s\n", ead.String())
			//fmt.Printf("%s\n", ead.ToXML())
		})

		//It("should read all the files in a directory", func() {
		//path, err := filepath.Abs("/mnt/usb1/ead-production/")
		//Expect(err).ToNot(HaveOccurred())
		//files, err := ioutil.ReadDir(path)
		//Expect(err).ToNot(HaveOccurred())
		//Expect(files).ToNot(BeEmpty())
		//for _, f := range files {
		//if strings.HasSuffix(f.Name(), "ead.xml") {
		//ead, err := ReadEAD(fmt.Sprintf("/mnt/usb1/ead-production/%s", f.Name()))
		//Expect(err).ToNot(HaveOccurred())
		//Expect(ead).ToNot(BeNil())
		//fmt.Printf("%s\n", f.Name())
		//cfg := NewNodeConfig(context.Background())
		//cfg.Spec = strings.TrimSuffix(f.Name(), ".ead.xml")
		//cfg.OrgID = c.Config.OrgID
		////cfg.Revision = int32(ds.Revision)
		//basePath := fmt.Sprintf("%s/%s", c.Config.EAD.CacheDir, cfg.Spec)

		//nl, _, err := ead.Carchdesc.Cdsc.NewNodeList(cfg)
		//Expect(nl).ToNot(BeNil())
		//Expect(err).ToNot(HaveOccurred())
		//Expect(ead).ToNot(BeNil())
		//if len(cfg.Errors) > 0 {
		//d, err := cfg.ErrorToCSV()
		//Expect(err).ToNot(HaveOccurred())
		//err = ioutil.WriteFile(basePath+".duplicate.csv", d, 0644)
		//Expect(err).ToNot(HaveOccurred())
		//}
		//}
		//}
		//})

		It("should serialize it to JSON", func() {
			path, err := filepath.Abs("./test_data/ead/NL-HaNA_2.08.22.ead.xml")
			Expect(err).ToNot(HaveOccurred())
			ead, err := ReadEAD(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(ead).ToNot(BeNil())
			b, err := json.Marshal(ead)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile("/tmp/test.ead", b, 0644)
			Expect(err).ToNot(HaveOccurred())

		})
	})

})

func parseUtil(node interface{}, fName string) error {
	dat, err := ioutil.ReadFile("test_data/ead/" + fName)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(dat, node)
	if err != nil {
		return err
	}
	return nil
}
