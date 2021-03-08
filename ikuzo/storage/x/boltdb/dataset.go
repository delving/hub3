package boltdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/asdine/storm"
	"github.com/delving/hub3/ikuzo/service/x/dataset"
	"github.com/rs/zerolog/log"
)

var (
	dbName = "hub3_bbolt.db"
)

type dataSetWrapper struct {
	DataSetID       string `json:"spec" storm:"id,index,unique"`
	dataset.DataSet `json:"dataset" storm:"inline"`
}

func NewDataSetStore(path string) (dataset.Store, error) {
	if path == "" {
		currentPath := getPath()
		path = filepath.Join(currentPath, dbName)
	}

	if !strings.HasSuffix(path, ".db") {
		path = fmt.Sprintf("%s.db", path)
	}

	db, err := storm.Open(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to open the BoltDB database file.")
	}

	store := dataSetStore{
		orm:  db,
		path: path,
	}

	log.Info().
		Str("full_path", path).
		Str("db_name", db.Bolt.Path()).
		Msg("starting boldDB")

	return &store, nil
}

type dataSetStore struct {
	orm  *storm.DB
	path string
}

func (d *dataSetStore) Delete(ctx context.Context, ds *dataset.DataSet) error {
	// TODO(kiivihal): implement me
	return nil
}

func (d *dataSetStore) List(ctx context.Context) ([]*dataset.DataSet, error) {
	// TODO(kiivihal): implement me
	return []*dataset.DataSet{}, nil
}

func (d *dataSetStore) One(ctx context.Context, orgID, spec string) (*dataset.DataSet, error) {
	// TODO(kiivihal): implement me
	return nil, nil
}

func (d *dataSetStore) Reset(ctx context.Context) error {
	if err := d.Shutdown(ctx); err != nil {
		return fmt.Errorf("bbolt: unable to reset storm dataset storage; %w", err)
	}

	if err := os.Remove(d.path); err != nil {
		return fmt.Errorf("bbolt: unable to delete database file %s; %w", d.path, err)
	}

	return nil
}

func (d *dataSetStore) Save(ctx context.Context, ds *dataset.DataSet) error {
	// TODO(kiivihal): implement me
	return nil
}

func (d *dataSetStore) Shutdown(ctx context.Context) error {
	err := d.orm.Close()
	if err != nil {
		return fmt.Errorf("unable to shutdown bbolt; %w", err)
	}

	return nil
}

func getPath() string {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	return filepath.Dir(ex)
}
