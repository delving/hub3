package elasticsearch

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Indices struct {
	client *Client
	alias  Alias
}

func (c *Client) Indices() Indices {
	return Indices{
		client: c,
		alias:  c.Alias(),
	}
}

// IndexCreate creates a new index with the supplied mapping.
//
// The internal index created from the alias and a timestamp. If 'force'
// is false, no index is created when the alias exists. When it does not exist
// a new index is created and the alias is set.
//
// When force is true, there is no check for the alias and no alias is set for
// created index.
func (idx *Indices) Create(alias, mapping string, withAlias bool) (indexName string, err error) {
	if withAlias {
		storedIndexName, aliasErr := idx.alias.Get(alias)
		if aliasErr != nil && !errors.Is(aliasErr, ErrAliasNotFound) {
			return "", aliasErr
		}

		if storedIndexName != "" {
			return storedIndexName, ErrIndexAlreadyCreated
		}
	}

	indexName = fmt.Sprintf("%s_%s", alias, time.Now().Format("20060102150405.999"))

	res, err := idx.client.index.Indices.Create(
		indexName,
		idx.client.index.Indices.Create.WithBody(strings.NewReader(mapping)),
		idx.client.index.Indices.Create.WithWaitForActiveShards("1"),
	)
	if err != nil {
		return "", fmt.Errorf("elastic Indices.Create: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		return "", GetErrorType(res.Body).Error()
	}

	if withAlias {
		err = idx.alias.Create(alias, indexName)
		if err != nil {
			return indexName, err
		}
	}

	return indexName, nil
}

// Delete delete the index from ElasticSearch.
//
// If the error does not exist an ErrIndexNotExist error is returned.
func (idx *Indices) Delete(indexName string) error {
	res, err := idx.client.index.Indices.Delete(
		[]string{indexName},
	)
	if err != nil {
		return err
	}

	if res.IsError() {
		return ErrIndexNotFound
	}

	return nil
}

func (idx Indices) Exists(indexName string) error {
	res, err := idx.client.index.Indices.Exists(
		[]string{indexName},
	)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return ErrIndexNotFound
	}

	return nil
}

// Switch updates the alias with the new index, and deletes the old index.
//
// When you only want to swith the alias use UpdateAlias().
func (idx *Indices) Switch(alias, newIndexName string, deleteOldIndex bool) (oldIndexName string, err error) {
	oldIndexName, err = idx.alias.Update(alias, newIndexName)
	if err != nil {
		return "", err
	}

	if deleteOldIndex {
		if deleteErr := idx.Delete(oldIndexName); deleteErr != nil {
			return oldIndexName, deleteErr
		}
	}

	return oldIndexName, nil
}
