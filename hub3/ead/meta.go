package ead

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/delving/hub3/hub3/models"
)

var ErrFileNotFound = errors.New("file not found")

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
	Revision         int32
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

	return os.WriteFile(
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
	if err == ErrFileNotFound {
		return &Meta{DatasetID: spec, Created: time.Now()}, true, nil
	}

	return meta, false, err
}

func GetMeta(spec string) (*Meta, error) {
	metaPath := getMetaPath(spec)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return nil, ErrFileNotFound
	}

	r, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	meta, err := decodeMeta(r)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func getMetaPath(spec string) string {
	return path.Join("/"+GetDataPath(spec), "meta.gob")
}
