package elasticsearch

import (
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/tidwall/gjson"
)

// AliasCreate creates an alias for the given indexName.
//
// When the index does not exist ErrIndexNotExists is returned
// When the alias is already defined, ErrAliasExist is returned
func AliasCreate(es *elasticsearch.Client, alias, indexName string) error {
	res, conErr := es.Indices.PutAlias(
		[]string{indexName},
		alias,
	)

	if conErr != nil {
		return fmt.Errorf("unable to connect: %w", conErr)
	}

	defer res.Body.Close()

	if res.IsError() {
		return GetErrorType(res.Body).Error()
	}

	return nil
}

// AliasGet returns the indexName for the given alias.
//
// When the alias is not found an ErrAliasNotFound error is returned.
func AliasGet(es *elasticsearch.Client, alias string) (indexName string, err error) {
	res, conErr := es.Indices.GetAlias(
		es.Indices.GetAlias.WithName(alias),
	)
	if conErr != nil {
		return "", conErr
	}

	defer res.Body.Close()

	if res.IsError() {
		return "", GetErrorType(res.Body).Error()
	}

	indexName = getIndexNameFromAlias(res.Body)

	return indexName, nil
}

func getIndexNameFromAlias(r io.Reader) string {
	json := read(r)

	// take the first index name for now
	for k := range gjson.Parse(json).Map() {
		return k
	}

	return ""
}

// AliasUpdate removes the alias if it exists from another index and creates a new one linked to indexName.
//
// It returns an error when the alias cannot be updated. It returns the old indexName that is removed
func AliasUpdate(es *elasticsearch.Client, alias, indexName string) (oldIndexName string, err error) {
	oldIndexName, err = AliasGet(es, alias)
	if err != nil {
		return oldIndexName, err
	}

	res, conErr := es.Indices.UpdateAliases(
		strings.NewReader(
			fmt.Sprintf(
				`
				{
					"actions" : [
						{ "remove" : { "index" : "%s", "alias" : "%s" } },
						{ "add" : { "index" : "%s", "alias" : "%s" } }
					]
				}
				`,
				oldIndexName,
				alias,
				indexName,
				alias,
			),
		),
	)

	if conErr != nil {
		return "", conErr
	}

	defer res.Body.Close()

	if res.IsError() {
		return "", GetErrorType(res.Body).Error()
	}

	return oldIndexName, nil
}

// AliasDelete removes the alias from the index it is linked to.
//
// When the alias does not it exist it will return a ErrAliasNotExist error.
//
// When the indexName is empty it will search for the indexName using GetAlias().
func AliasDelete(es *elasticsearch.Client, indexName, alias string) error {
	res, conErr := es.Indices.DeleteAlias(
		[]string{indexName},
		[]string{alias},
	)

	if conErr != nil {
		return conErr
	}

	defer res.Body.Close()

	if res.IsError() {
		return GetErrorType(res.Body).Error()
	}

	return nil
}
