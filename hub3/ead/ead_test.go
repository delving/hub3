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

package ead_test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	. "github.com/delving/hub3/hub3/ead"
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
				Expect(nl.Type).To(Equal("combined"))
				//Expect(nl.Nodes).ToNot(HaveLen(1))
				Expect(nl.Label).To(HaveLen(1))
				Expect(dsc.Chead[0].Head).To(HavePrefix("Beschrijving"))
				Expect(nl.Label[0]).To(HavePrefix("Beschrijving"))
			})

			It("should have a header", func() {
				Expect(err).ToNot(HaveOccurred())
				cfg := NewNodeConfig(context.Background())
				nl, seen, err := dsc.NewNodeList(cfg)
				Expect(seen).To(Equal(uint64(1)))
				Expect(err).ToNot(HaveOccurred())
				Expect(nl.Label).To(HaveLen(1))
				sourceHeader := dsc.Chead[0].Head
				Expect(nl.Label[0]).To(HavePrefix("Beschrijving"))
				Expect(nl.Label[0]).To(Equal(sourceHeader))
			})

			It("should have c-levels", func() {
				Expect(dsc.Cc).To(HaveLen(1))
				cfg := NewNodeConfig(context.Background())
				nl, seen, err := dsc.NewNodeList(cfg)
				Expect(seen).To(Equal(uint64(1)))
				Expect(err).ToNot(HaveOccurred())
				Expect(nl.Nodes).To(HaveLen(1))
				// TODO add nil check
			})
		})

		Context("for the c01", func() {
			cc := new(Cc)
			err := parseUtil(cc, "ead.2.xml")
			cfg := NewNodeConfig(context.Background())
			var node *Node

			It("should not throw an error on create", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(cc.GetXMLName().Local).To(Equal("c"))
				node, err = NewNode(cc, []string{}, cfg)
				Expect(node.Order).To(Equal(uint64(1)))
				Expect(err).ToNot(HaveOccurred())
			})

			It("should have a cTag", func() {
				Expect(node).ToNot(BeNil())
				Expect(node.CTag).ToNot(BeEmpty())
				Expect(node.CTag).To(Equal("c"))
				Expect(node.Depth).To(Equal(int32(1)))
			})

			It("should set the depth", func() {
				Expect(node.Depth).To(Equal(int32(1)))
			})

			It("should have a type", func() {
				Expect(cc.GetAttrlevel()).To(Equal("series"))
				//Expect(node.Type).To(Equal("series"))
			})

			It("should have a subType", func() {
				Expect(node.SubType).To(BeEmpty())
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
				Expect(header.Label).To(HaveLen(1))
				Expect(header.Label[0]).To(Equal("Octrooi verleend door de Staten-Generaal betreffende de alleenhandel ten oosten van Kaap de Goede Hoop en ten westen van de Straat van Magallanes voor de duur van 21 jaar"))
				Expect(header.GetTreeLabel()).To(Equal("Octrooi verleend door de Staten-Generaal betreffende de alleenhandel ten oosten van Kaap de Goede Hoop en ten westen van de Straat van Magallanes voor de duur van 21 jaar"))
			})

			It("should not have date as label", func() {
				Expect(header.DateAsLabel).To(BeFalse())
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
				Expect(header.Label).To(HaveLen(1))
				Expect(header.Label[0]).To(Equal("ca. 1839 new books."))
			})

			It("should have date as label", func() {
				Expect(header.DateAsLabel).To(BeTrue())
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
				Expect(header.Physdesc).ToNot(BeEmpty())
			})

			It("should have not have date as label", func() {
				Expect(header.DateAsLabel).To(BeFalse())
			})

			It("should have a label", func() {
				Expect(header.Label).ToNot(BeEmpty())
			})

			It("should have a date", func() {
				Expect(header.Date).ToNot(BeEmpty())
			})

			Context("when extracting a NodeDate", func() {

				unitDate := did.Cunitdate[0]
				nDate, err := unitDate.NewNodeDate()

				It("should not thrown an error on creation", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(nDate).ToNot(BeNil())
				})

				It("should have an calendar", func() {
					Expect(nDate.Calendar).To(Equal(unitDate.Attrcalendar))
				})

				It("should have an era", func() {
					Expect(nDate.Era).To(Equal(unitDate.Attrera))
				})

				It("should have a normal string", func() {
					Expect(nDate.Normal).To(Equal(unitDate.Attrnormal))
				})

				It("should have the date as string", func() {
					Expect(nDate.Label).To(Equal(unitDate.Unitdate))
				})

			})

			Context("when extracting a handle unitID", func() {

				unitID := did.Cunitid[1]
				id, err := unitID.NewNodeID()

				It("should have an ID", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(id).ToNot(BeNil())
					Expect(id.ID).To(Equal(unitID.Unitid))
				})

				It("should have a Type", func() {
					Expect(id.Type).To(Equal(unitID.Attrtype))
				})

				It("should have an audience", func() {
					Expect(id.Audience).To(Equal(unitID.Attraudience))
				})

			})

			Context("when extracting from an ABS unitID", func() {

				unitID := did.Cunitid[0]
				id, err := unitID.NewNodeID()

				It("should have an ID", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(id).ToNot(BeNil())
					Expect(id.ID).To(Equal(unitID.Unitid))
				})

				It("should have a TypeID", func() {
					Expect(id.TypeID).To(Equal(unitID.Attridentifier))
				})

				It("should have a Type", func() {
					Expect(id.Type).To(Equal(unitID.Attrtype))
				})

				It("should extract the nodeIDs", func() {
					nodeIDs, inventoryNumber, err := did.NewNodeIDs()
					Expect(err).ToNot(HaveOccurred())
					Expect(nodeIDs).ToNot(BeEmpty())
					Expect(inventoryNumber).ToNot(BeEmpty())
				})
			})

			Context("when extracting nodeIDs from various types", func() {
				var createUnitIDs = func(id string, nodeType string) []*Cunitid {
					var cus []*Cunitid
					cu := &Cunitid{
						Unitid:   id,
						Attrtype: nodeType,
					}
					cus = append(cus, cu)
					return cus
				}
				tests := []struct {
					name   string
					cdid   Cdid
					wantID string
				}{
					{
						"Should extract series_code type unit id",
						Cdid{Cunitid: createUnitIDs("100", "series_code")},
						"100",
					},
					{
						"Should not extract unknown type",
						Cdid{Cunitid: createUnitIDs("69", "unknown_type")},
						"",
					},
					{
						"Should extract dashes from blank types",
						Cdid{Cunitid: createUnitIDs("---", "blank")},
						"---",
					},
					{
						"Should extract analoog",
						Cdid{Cunitid: createUnitIDs("10000", "analoog")},
						"10000",
					},
					{
						"Should extract Born Digital type",
						Cdid{Cunitid: createUnitIDs("DC-2015/446", "BD")},
						"DC-2015/446",
					},
					{
						"Should extract from empty type even when silly",
						Cdid{Cunitid: createUnitIDs("miauw", "")},
						"miauw",
					},
				}
				for _, tt := range tests {
					tt := tt
					It(tt.name, func() {
						_, inventoryID, _ := tt.cdid.NewNodeIDs()
						Expect(inventoryID).To(Equal(tt.wantID))
					})
				}
			})

		})

		Context("for the p in the dsc", func() {
			dsc := new(Cdsc)
			err := parseUtil(dsc, "ead.dsc.xml")
			It("should not throw an error on create", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should create a c level series from the p", func() {
				cfg := NewNodeConfig(context.Background())
				nodes, nodeCount, err := dsc.NewNodeList(cfg)
				Expect(nodes).ToNot(BeNil())
				Expect(nodeCount).To(Equal(uint64(2)))
				Expect(err).To(BeNil())
				first := nodes.Nodes[0]
				Expect(first.CTag).To(Equal("c"))
				Expect(first.Type).To(Equal("file"))
				Expect(first.Header.Label[0]).To(Equal("Let op: deze inventaris is alleen voor het inzien van de dossiers. Kijk in de index voor het zoeken op naam. https://www.nationaalarchief.nl/onderzoeken/index/nt00446."))
			})
		})

	})

	Context("when being load from file", func() {

		It("should create an EAD object", func() {
			path, err := filepath.Abs("./testdata/ead/NL-HaNA_2.08.22.ead.xml")
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
			path, err := filepath.Abs("./testdata/ead/NL-HaNA_2.08.22.ead.xml")
			Expect(err).ToNot(HaveOccurred())
			ead, err := ReadEAD(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(ead).ToNot(BeNil())
			b, err := json.Marshal(ead)
			Expect(err).ToNot(HaveOccurred())
			f, err := ioutil.TempFile("/tmp", "test.ead")
			Expect(err).ToNot(HaveOccurred())

			_, err = fmt.Fprint(f, b)
			Expect(err).ToNot(HaveOccurred())

		})
	})

})

func parseUtil(node interface{}, fName string) error {
	dat, err := ioutil.ReadFile("testdata/ead/" + fName)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(dat, node)
	if err != nil {
		return err
	}
	return nil
}

func TestNodeDate_ValidDateNormal(t *testing.T) {
	type fields struct {
		Calendar string
		Era      string
		Normal   string
		Label    string
		Type     string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"correct date range",
			fields{Normal: "1990-01-01/1995-10-31"},
			false,
		},
		{
			"correct year range",
			fields{Normal: "1990/1995"},
			false,
		},
		{
			"single date",
			fields{Normal: "1990-01-01"},
			false,
		},
		{
			"wrong range",
			fields{Normal: "2000-12-01/1990-01-01"},
			true,
		},
		{
			"partial start",
			fields{Normal: "1990-01-01/"},
			false,
		},
		{
			"partial end",
			fields{Normal: "/1990-01-01"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nd := &NodeDate{
				Calendar: tt.fields.Calendar,
				Era:      tt.fields.Era,
				Normal:   tt.fields.Normal,
				Label:    tt.fields.Label,
				Type:     tt.fields.Type,
			}
			if err := nd.ValidDateNormal(); (err != nil) != tt.wantErr {
				t.Errorf("NodeDate.ValidDateNormal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
