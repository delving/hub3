package rdf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/delving/hub3/ikuzo/domain"
)

var ErrRecordNotFound = errors.New("record not found")

// Record is the primary grouping of triples to represent a search record.
// This is a replacement of the old fragments.FragmentGraph.
// lod.RecordStore interface interacts as a storage layer for these records
//
// TODO(kiivihal): implement index record
type Record struct {
	ID         domain.HubID
	CreatedAt  time.Time
	ModifiedAt time.Time
	Hash       string
	Deleted    bool
	Version    int32
	GraphData  []byte
	GraphURI   string
	MimeType   string
}

func path(hubID domain.HubID, version int32) string {
	parts := []string{hubID.OrgID, hubID.DatasetID}
	if version > 0 {
		parts = append(parts, "records-tmp", fmt.Sprintf("%d", version))
	} else {
		parts = append(parts, "records")
	}

	parts = append(parts, hubID.String()+".json")

	return filepath.Join(parts...)
}

func (r *Record) Path() string {
	return path(r.ID, 0)
}

func (r *Record) tmpPath(version int32) string {
	return path(r.ID, version)
}

func ReadRecord(hubID domain.HubID, basePath string) (*Record, error) {
	fname := filepath.Join(basePath, path(hubID, 0))

	file, err := os.Open(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("unable to open file %s; %w", fname, err)
	}

	defer file.Close()

	var rec Record

	err = json.NewDecoder(file).Decode(&rec)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal file %s; %w", fname, err)
	}

	return &rec, nil
}

func (r *Record) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func (r *Record) Merge(previous *Record) (bool, error) {
	return false, nil
}

// Write will write to path and append the relative path from r.Path()
// When the r.Hash is not identical a new record will be written, otherwise
// only the Version will be updated
func (r *Record) Write(path string) error {
	fname := filepath.Join(path, r.Path())

	prev, err := ReadRecord(r.ID, path)
	if err != nil {
		if !errors.Is(err, ErrRecordNotFound) {
			return err
		}

		if err := ensureDir(filepath.Dir(fname)); err != nil {
			return err
		}
		prev = r
	}

	if prev != nil {
		r.Merge(prev)
	}

	if r.Hash == prev.Hash {
		prev.Version = r.Version
	} else {
		prev.Hash = r.Hash
		prev.GraphData = r.GraphData
	}

	b, err := json.Marshal(prev)
	if err != nil {
		return err
	}

	return os.WriteFile(fname, b, os.ModePerm)
}

func (r *Record) WriteTmp(path string) error {
	// ensure dir
	return nil
}

func (r *Record) Graph() (*Graph, error) {
	return nil, nil
}

func ensureDir(dirPath string) error {
	err := os.MkdirAll(dirPath, os.ModePerm)

	if err == nil || os.IsExist(err) {
		return nil
	}

	return err
}
