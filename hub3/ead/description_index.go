package ead

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/service/x/search"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
)

var (
	ErrNoDescriptionIndex = errors.New("no index created for EAD description")
)

const (
	startHighlightTag    = "em"
	hightlightStyleClass = "dhcl"
)

type DescriptionIndex struct {
	spec string
	ti   *memory.TextIndex
}

func NewDescriptionIndex(spec string) *DescriptionIndex {
	return &DescriptionIndex{
		spec: spec,
		ti:   memory.NewTextIndex(),
	}
}

func (di *DescriptionIndex) CreateFrom(desc *Description) error {
	for _, item := range desc.Item {
		if item.Text != "" {
			err := di.ti.AppendString(item.Text, int(item.Order))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (di *DescriptionIndex) Write() error {
	err := os.MkdirAll(GetDataPath(di.spec), os.ModePerm)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	err = di.ti.Encode(&buf)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		getIndexPath(di.spec),
		buf.Bytes(),
		0644,
	)
}

func (di *DescriptionIndex) Search(qt *search.QueryTerm) (*search.Matches, error) {
	return di.ti.Search(qt)
}

func (di *DescriptionIndex) SearchWithString(query string) (*search.Matches, error) {
	rlog := config.Config.Logger.With().
		Str("application", "hub3").
		Str("search.type", "request builder").
		Logger()

	queryParser, parseErr := search.NewQueryParser()
	if parseErr != nil {
		rlog.Error().Err(parseErr).
			Str("subquery", "description").
			Msg("unable to create search.QueryParser.")

		return nil, parseErr
	}

	qt, queryErr := queryParser.Parse(query)
	if queryErr != nil {
		rlog.Error().Err(queryErr).
			Str("subquery", "description").
			Msg("unable to parse query into search.QueryTerm")

		return nil, queryErr
	}

	return di.Search(qt)
}

func (di *DescriptionIndex) HighlightMatches(hits *search.Matches, items []*DataItem, filter bool) []*DataItem {
	matches := []*DataItem{}

	for _, item := range items {
		hasDoc := hits.HasDocID(int(item.Order))
		if filter && !hasDoc {
			continue
		}

		if !hasDoc {
			matches = append(matches, item)
			continue
		}

		tok := search.NewTokenizer()
		ts := tok.ParseString(item.Text, int(item.Order))

		item.Text = ts.Highlight(hits.Vectors(), startHighlightTag, hightlightStyleClass)

		matches = append(matches, item)
	}

	return matches
}

func GetDescriptionIndex(spec string) (*DescriptionIndex, error) {
	indexPath := getIndexPath(spec)
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil, ErrNoDescriptionIndex
	}

	r, err := os.Open(indexPath)
	if err != nil {
		return nil, err
	}

	ti, err := memory.DecodeTextIndex(r)
	if err != nil {
		return nil, err
	}

	di := NewDescriptionIndex(spec)
	di.ti = ti

	return di, nil
}

func getIndexPath(spec string) string {
	return path.Join(GetDataPath(spec), "description_index.gob")
}

func getDescriptionPath(spec string) string {
	return path.Join(GetDataPath(spec), "description.gob")
}

func GetDataPath(spec string) string {
	return path.Join(config.Config.EAD.CacheDir, spec)
}
