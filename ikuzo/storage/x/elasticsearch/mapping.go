package elasticsearch

import (
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/storage/x/elasticsearch/mapping"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

func IsMappingValid(es *elasticsearch.Client, indexName string) (bool, error) {
	res, err := es.Indices.GetMapping(es.Indices.GetMapping.WithIndex(indexName))
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	json := read(res.Body)

	m := gjson.Parse(json).Map()

	v, ok := m[indexName]
	if !ok {
		return false, nil
	}

	return mapping.ValidMapping(v.Get("mappings").Raw), nil
}

func MappingUpdate(es *elasticsearch.Client, indexName, esMapping string) error {
	resp, err := es.Indices.PutMapping([]string{indexName}, strings.NewReader(esMapping))
	if err != nil {
		return fmt.Errorf("unable to update mapping; %w", err)
	}

	if resp.HasWarnings() {
		log.Warn().Msgf("mapping update warnings: %#v", resp.Warnings())
	}

	return nil
}
