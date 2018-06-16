package hub3

import (
	"io/ioutil"

	"github.com/caltechlibrary/ead2002"
)

func ReadEAD(path string) (*ead2002.EAD, error) {
	rawEAD, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ead2002.Parse(rawEAD)
}
