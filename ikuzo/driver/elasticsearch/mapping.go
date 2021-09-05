package elasticsearch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch/internal/mapping"
	"github.com/tidwall/gjson"
)

func (c *Client) createDefaultMappings(orgCfg domain.OrganizationConfig, withAlias, withReset bool) (indexNames []string, err error) {
	mappings := map[string]func(shards, replicas int) string{}

	for _, indexType := range orgCfg.ElasticSearch.IndexTypes {
		switch indexType {
		case "v1":
			mappings[orgCfg.GetV1IndexName()] = mapping.V1ESMapping
		case "v2":
			mappings[orgCfg.GetIndexName()] = mapping.V2ESMapping
		case "fragments":
			mappings[orgCfg.GetFragmentsIndexName()] = mapping.FragmentESMapping
		default:
			c.log.Warn().Msgf("ignoring unknown indexType %s during mapping creation", indexType)
		}
	}

	if orgCfg.ElasticSearch.DigitalObjectSuffix != "" {
		indexName := orgCfg.GetIndexName() + "-" + orgCfg.ElasticSearch.DigitalObjectSuffix
		mappings[indexName] = mapping.V2ESMapping
	}

	indexNames = []string{}

	indices := c.Indices()

	for indexName, m := range mappings {
		if withReset {
			storedIndexName, aliasErr := indices.alias.Get(indexName)
			if aliasErr != nil && !errors.Is(aliasErr, ErrAliasNotFound) {
				return []string{}, aliasErr
			}

			if storedIndexName != "" {
				if err := indices.alias.Delete(storedIndexName, indexName); err != nil {
					c.log.Error().Err(err).Str("alias", indexName).
						Str("index", storedIndexName).Msg("unable to delete alias")

					return []string{}, err
				}

				deleteErr := indices.Delete(storedIndexName)
				if deleteErr != nil {
					c.log.Error().Err(deleteErr).Str("alias", indexName).
						Str("index", storedIndexName).Msg("unable to delete index")
					return []string{}, deleteErr
				}
			}
		}

		createName, err := indices.Create(
			indexName,
			m(orgCfg.ElasticSearch.Shards, orgCfg.ElasticSearch.Replicas),
			withAlias,
		)

		if err != nil && !errors.Is(err, ErrIndexAlreadyCreated) {
			return []string{}, err
		}

		if errors.Is(err, ErrIndexAlreadyCreated) {
			if strings.HasSuffix(indexName, "v2") {
				valid, err := c.isMappingValid(createName)
				if err != nil {
					return []string{}, err
				}

				if !valid {
					if err := c.mappingUpdate(createName, mapping.V2MappingUpdate()); err != nil {
						c.log.Error().Err(err).Msg("unable to apply v2 mapping update")
						return []string{}, err
					}

					c.log.Warn().Str("index", createName).Msg("applying elasticsearch mapping update")
				}
			}
		}

		indexNames = append(indexNames, createName)
	}

	return indexNames, nil
}

func (c *Client) isMappingValid(indexName string) (bool, error) {
	res, err := c.index.Indices.GetMapping(c.index.Indices.GetMapping.WithIndex(indexName))
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

func (c *Client) mappingUpdate(indexName, esMapping string) error {
	resp, err := c.index.Indices.PutMapping([]string{indexName}, strings.NewReader(esMapping))
	if err != nil {
		return fmt.Errorf("unable to update mapping; %w", err)
	}

	if resp.HasWarnings() {
		c.log.Warn().Msgf("mapping update warnings: %#v", resp.Warnings())
	}

	return nil
}
