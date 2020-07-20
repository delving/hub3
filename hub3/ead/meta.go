package ead

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/models"
)

var (
	ErrNoFileNotFound = errors.New("file not found")
)

// Meta holds all processing information for an EAD Archive
type Meta struct {
	OrgID            string
	DatasetID        string
	Label            string
	Period           []string
	Inventories      int
	DigitalObjects   int
	RecordsPublished int
	MetsFiles        int
	Created          time.Time
	Updated          time.Time
	TimesUploaded    int
	DaoStats         models.DaoStats
}

func (m *Meta) Write() error {
	err := os.MkdirAll(GetDataPath(m.DatasetID), os.ModePerm)
	if err != nil {
		return err
	}

	m.Updated = time.Now()
	m.TimesUploaded++

	var buf bytes.Buffer

	err = m.encode(&buf)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		getMetaPath(m.DatasetID),
		buf.Bytes(),
		os.ModePerm,
	)
}

func (m *Meta) encode(w io.Writer) error {
	e := gob.NewEncoder(w)

	err := e.Encode(m)
	if err != nil {
		return fmt.Errorf("unable to marshall ead.Meta to GOB; %w", err)
	}

	return nil
}

func decodeMeta(r io.Reader) (*Meta, error) {
	var meta Meta

	d := gob.NewDecoder(r)

	err := d.Decode(&meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

func GetOrCreateMeta(spec string) (*Meta, bool, error) {
	meta, err := GetMeta(spec)
	if err == ErrNoFileNotFound {
		return &Meta{DatasetID: spec, Created: time.Now()}, true, nil
	}

	return meta, false, err
}

func getRemoteMeta(spec string) (*Meta, error) {
	var meta Meta

	var netClient = &http.Client{
		Timeout: time.Second * 1,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/ead/%s/meta", c.Config.DataNodeURL, spec), nil)
	if err != nil {
		return nil, err
	}

	resp, err := netClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decodeErr := json.NewDecoder(resp.Body).Decode(&meta)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &meta, nil

}

func GetMeta(spec string) (*Meta, error) {
	if !c.Config.IsDataNode() {
		return getRemoteMeta(spec)
	}

	metaPath := getMetaPath(spec)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return nil, ErrNoFileNotFound
	}

	r, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}

	meta, err := decodeMeta(r)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func getMetaPath(spec string) string {
	return path.Join(GetDataPath(spec), "meta.gob")
}
