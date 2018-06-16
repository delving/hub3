package hub3_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/caltechlibrary/ead2002"
	//. "github.com/delving/rapid-saas/hub3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ead", func() {

	Context("when being load from file", func() {

		It("should create an EAD object", func() {
			path, err := filepath.Abs("./models/test_data/ead/NL-HaNA_2.08.22.ead.xml")
			Expect(err).ToNot(HaveOccurred())
			rawEAD, err := ioutil.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(rawEAD).ToNot(BeNil())
			ead, err := ead2002.Parse(rawEAD)
			Expect(err).ToNot(HaveOccurred())
			Expect(ead).ToNot(BeNil())
			//fmt.Printf("%#v\n", ead.ArchDesc.DSC.Value)
			//fmt.Printf("%s\n", ead.String())
			//fmt.Printf("%s\n", ead.ToXML())
		})

		It("should read all the files in a directory", func() {
			path, err := filepath.Abs("./models/test_data/ead/")
			Expect(err).ToNot(HaveOccurred())
			files, err := ioutil.ReadDir(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(files).ToNot(BeEmpty())
			for _, f := range files {
				//ead, err := ReadEAD(fmt.Sprintf("/mnt/usb2/ead-production/%s", f.Name()))
				//Expect(err).ToNot(HaveOccurred())
				//Expect(ead).ToNot(BeNil())
				fmt.Printf("%s\n", f.Name())
			}
		})
	})

})
