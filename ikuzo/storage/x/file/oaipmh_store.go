package file

import (
	"context"
	"encoding/json"
	stdliberrors "errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
)

// var _ oaipmh.Store = (*OAIPMHStore)(nil)

type OAIPMHStore struct {
	Path string // TODO(kiivihal): replace with bucket later
}

type repoConfig struct{}

type setConfig struct {
	Name             string
	MetadataPrefixes []string
	Prefix           string
}

func NewOAIPMHStore() (*OAIPMHStore, error) {
	return &OAIPMHStore{}, nil
}

func (o *OAIPMHStore) readSetConfig(path, spec string) (setConfig, error) {
	var cfg setConfig

	f, err := os.Open(filepath.Join(path, spec, ".config.json"))
	if err != nil {
		return cfg, err
	}

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (o *OAIPMHStore) ListSets(ctx context.Context, q *oaipmh.RequestConfig) (sets []oaipmh.Set, errors []oaipmh.Error, err error) {
	rawSets, err := o.filterSets(q.OrgID, "", "")
	if err != nil {
		return sets, errors, err
	}

	for spec, cfg := range rawSets {
		sets = append(sets, oaipmh.Set{
			SetSpec: spec,
			SetName: cfg.Name,
		})
	}

	return sets, errors, err
}

func (o *OAIPMHStore) filterSets(orgID, specPrefix, metadataPrefix string) (sets map[string]setConfig, err error) {
	path := filepath.Join(o.Path, orgID)

	sets = make(map[string]setConfig)

	datasets, err := os.ReadDir(path)
	if err != nil {
		return sets, err
	}

	for _, dataset := range datasets {
		if !dataset.IsDir() {
			continue
		}

		cfg, cfgErr := o.readSetConfig(path, dataset.Name())
		if cfgErr != nil {
			if stdliberrors.Is(cfgErr, os.ErrNotExist) {
				continue
			}

			return sets, cfgErr
		}

		if specPrefix != "" {
			comp := strings.EqualFold

			if strings.Contains(dataset.Name(), "_") {
				specPrefix = specPrefix + "_"
				comp = strings.HasPrefix
			}

			if !comp(dataset.Name(), specPrefix) {
				continue
			}
		}

		if metadataPrefix != "" {
			var found bool

			for _, prefix := range cfg.MetadataPrefixes {
				if strings.EqualFold(metadataPrefix, prefix) {
					found = true
				}
			}

			if !found {
				continue
			}
		}

		sets[dataset.Name()] = cfg
	}

	return sets, nil
}

func (o *OAIPMHStore) ListIdentifiers(ctx context.Context, q *oaipmh.RequestConfig) (
	headers []oaipmh.Header, errors []oaipmh.Error, err error,
) {
	path := filepath.Join(o.Path, q.OrgID, q.DatasetID, q.FirstRequest.MetadataPrefix)

	files, err := os.ReadDir(path)
	if err != nil {
		if stdliberrors.Is(err, os.ErrNotExist) {
			errors = append(errors, oaipmh.ErrNoMetadataFormats)
			return headers, errors, err
		}
	}

	q.TotalSize = len(files)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, infoErr := file.Info()
		if infoErr != nil {
			return headers, errors, infoErr
		}

		headers = append(headers, oaipmh.Header{
			Identifier: file.Name(),
			DateStamp:  info.ModTime().Format(oaipmh.TimeFormat),
			SetSpec:    []string{},
			Status:     "",
		})
	}

	return headers, errors, err
}

// func (o *OAIPMHStore) ListRecords(ctx context.Context, q *oaipmh.QueryConfig) (records []oaipmh.Record, err error, errors []oaipmh.Error) {
// return records, err, errors
// }

// func (o *OAIPMHStore) GetRecord(ctx context.Context, q *oaipmh.QueryConfig) (record oaipmh.Record, err error, errors []oaipmh.Error) {
// return record, err, errors
// }
