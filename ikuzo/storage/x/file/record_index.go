package file

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
)

type metrics struct {
	Records      int64
	Deleted      int64
	EarliestDate time.Time
	LatestDate   time.Time
}

type index struct {
	OrgID           string
	DatasetID       string
	DataSets        []oaipmh.Set // TODO(kiivihal): maybe one
	AllowedFormats  []oaipmh.MetadataFormat
	SourceFormat    SourceFormat
	SourceExtension string
	lookUp          map[string]*recordPointer
	records         []*recordPointer // sorted by last modified
	files           map[string]*os.File
	m               *metrics
}

/*
use generic bucket (gocloud) to store all data to disk
this makes it possible to run this in a loadbalanced mode possible
*/
func newIndex(orgID, datasetID string, format SourceFormat) (*index, error) {
	idx := &index{
		OrgID:           orgID,
		DatasetID:       datasetID,
		SourceFormat:    format,
		SourceExtension: "",
		lookUp:          map[string]*recordPointer{},
		files:           map[string]*os.File{},
	}

	switch format {
	case FormatRaw:
	case FormatNTriples:
	case FormatEAD:
	default:
		return nil, fmt.Errorf("unsupported source format: %s", format)
	}

	return idx, nil
}

func (idx *index) Close() error {
	for _, f := range idx.files {
		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (idx *index) sortRecords() {
	sort.Slice(idx.records, func(i, j int) bool {
		return idx.records[i].LastModified.Before(idx.records[j].LastModified)
	})
}

func (idx *index) ParseNarthexNtriples(r io.Reader, fname, fhash string) error {
	var (
		lines           int
		offset          int
		recordOffset    int
		recordLength    int
		recordStartLine int
		records         int
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines++

		line := scanner.Bytes()
		if len(line) == 0 {
			offset++ // empty line still has \n
			continue
		}

		if IsRecordSeparator(line) {
			records++

			sep, err := NewRecordSeparator(line)
			if err != nil {
				return err
			}

			rp := sep.newRecordPointer(fname, fhash)
			rp.Offset = int64(recordOffset)
			rp.Length = int64(recordLength)
			rp.StartLine = int64(recordStartLine)
			rp.Lines = int64(lines - recordStartLine)

			idx.records = append(idx.records, rp)
			idx.lookUp[sep.HubID] = rp

			recordOffset += (recordLength + 1 + len(line))
			recordStartLine = lines + 1
			recordLength = 0

			continue
		}

		offset += len(line) + 1
		recordLength += len(line) + 1
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func readIndex(orgID, datasetID string) (*index, error) {
	path := filepath.Join(orgID, datasetID, configDirFname, indexFname)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	idx, err := decodeIndex(f)
	if err != nil {
		return nil, err
	}

	return idx, nil
}

func (idx *index) Write() error {
	dirPath := filepath.Join(idx.OrgID, idx.DatasetID, configDirFname)

	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	err = idx.encode(&buf)
	if err != nil {
		return err
	}

	path := filepath.Join(dirPath, indexFname)

	return os.WriteFile(
		path,
		buf.Bytes(),
		os.ModePerm,
	)
}

func (idx *index) encode(w io.Writer) error {
	e := gob.NewEncoder(w)

	err := e.Encode(idx)
	if err != nil {
		return fmt.Errorf("unable to marshal file.index; %w", err)
	}

	return nil
}

func decodeIndex(r io.Reader) (*index, error) {
	var idx index

	d := gob.NewDecoder(r)

	err := d.Decode(&idx)
	if err != nil {
		return nil, err
	}

	return &idx, nil
}
