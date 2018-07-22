package ead_test

import (
	"encoding/xml"
	"io/ioutil"

	. "github.com/delving/rapid-saas/hub3/ead"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ead", func() {

	Describe("when converting to Nodes", func() {

		Context("for the dsc", func() {
			dsc := new(Cdsc)
			err := parseUtil(dsc, "ead.1.xml")

			It("should have a type", func() {
				Expect(err).ToNot(HaveOccurred())
				nl, err := dsc.NewNodeList()
				Expect(err).ToNot(HaveOccurred())
				Expect(nl.GetType()).To(Equal("combined"))
				//Expect(nl.GetNodes()).ToNot(HaveLen(1))
				Expect(nl.GetLabel()).To(HaveLen(1))
				Expect(dsc.Chead[0].Head).To(HavePrefix("Beschrijving"))
				Expect(nl.GetLabel()[0]).To(HavePrefix("Beschrijving"))
			})

			It("should have a header", func() {
				Expect(err).ToNot(HaveOccurred())
				nl, err := dsc.NewNodeList()
				Expect(err).ToNot(HaveOccurred())
				Expect(nl.GetLabel()).To(HaveLen(1))
				sourceHeader := dsc.Chead[0].Head
				Expect(nl.GetLabel()[0]).To(HavePrefix("Beschrijving"))
				Expect(nl.GetLabel()[0]).To(Equal(sourceHeader))
			})

			It("should have c-levels", func() {
				Expect(dsc.Cc01).To(HaveLen(1))
				nl, err := dsc.NewNodeList()
				Expect(err).ToNot(HaveOccurred())
				Expect(nl.GetNodes()).To(HaveLen(1))
				// TODO add nil check
			})
		})

		Context("for the c01", func() {
			c01 := new(Cc01)
			err := parseUtil(c01, "ead.2.xml")
			var node *Node

			It("should not throw an error on create", func() {
				Expect(err).ToNot(HaveOccurred())
				node, err = c01.NewNode()
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
				Expect(node.GetType()).To(Equal("series"))
			})

			It("should have a subType", func() {
				Expect(node.GetSubType()).To(BeEmpty())
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

	//Context("when being load from file", func() {

	//It("should create an EAD object", func() {
	//path, err := filepath.Abs("./test_data/ead/NL-HaNA_2.08.22.ead.xml")
	////Expect(err).ToNot(HaveOccurred())
	////rawEAD, err := ioutil.ReadFile(path)
	//Expect(err).ToNot(HaveOccurred())
	////Expect(rawEAD).ToNot(BeNil())
	//ead, err := ReadEAD(path)
	//Expect(err).ToNot(HaveOccurred())
	//Expect(ead).ToNot(BeNil())
	////fmt.Printf("%#v\n", ead.ArchDesc.DSC.Value)
	////fmt.Printf("%s\n", ead.String())
	////fmt.Printf("%s\n", ead.ToXML())
	//})

	//It("should read all the files in a directory", func() {
	//path, err := filepath.Abs("/mnt/usb1/ead-production/")
	//Expect(err).ToNot(HaveOccurred())
	//files, err := ioutil.ReadDir(path)
	//Expect(err).ToNot(HaveOccurred())
	//Expect(files).ToNot(BeEmpty())
	////for _, f := range files {
	////if strings.HasSuffix(f.Name(), "ead.xml") {
	//////ead, err := ReadEAD(fmt.Sprintf("/mnt/usb1/ead-production/%s", f.Name()))
	//////Expect(err).ToNot(HaveOccurred())
	//////Expect(ead).ToNot(BeNil())
	////fmt.Printf("%s\n", f.Name())
	////}
	////}
	//})

	//It("should serialize it to JSON", func() {
	//path, err := filepath.Abs("/mnt/usb1/ead-production/NL-HaNA_2.08.28.ead.xml")
	//Expect(err).ToNot(HaveOccurred())
	//ead, err := ReadEAD(path)
	//Expect(err).ToNot(HaveOccurred())
	//Expect(ead).ToNot(BeNil())
	//b, err := json.Marshal(ead)
	//Expect(err).ToNot(HaveOccurred())
	//err = ioutil.WriteFile("/tmp/test.ead", b, 0644)
	//Expect(err).ToNot(HaveOccurred())

	//})
	//})

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

//func parseEAD(t *testing.T, node interface{}, fName string) {
//dat, err := ioutil.ReadFile("test_data/ead/" + fName)
//if err != nil {
//t.Errorf("Unable to read file %s: %s", fName, err)
//t.Fail()
//}
//err = xml.Unmarshal(dat, node)
//if err != nil {
//t.Errorf("Unable to unmarshal EAD: %s", err)
//t.Fail()
//}
//}

//func TestCdsc_NewNodeList(t *testing.T) {
//tests := []struct {
//name    string
//fName   string
//want    *NodeList
//wantErr bool
//}{
//{
//"single header",
//"ead.1.xml",
////&NodeList{Type: "combined", Label: []string{"Beschrijving van de series en archiefbestanddelen"}},
//&NodeList{Type: "combined", Label: []string{""}},
//false,
//},
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//cdsc := new(Cdsc)
//parseEAD(t, cdsc, tt.fName)
//got, err := cdsc.NewNodeList()
//if (err != nil) != tt.wantErr {
//t.Errorf("Cdsc.NewNodeList() error = %v, wantErr %v", err, tt.wantErr)
//return
//}
//if !reflect.DeepEqual(got, tt.want) {
//t.Errorf("Cdsc.NewNodeList() = %v, want %v", got, tt.want)
//}
//})
//}
//}
