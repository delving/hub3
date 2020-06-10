// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearch

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// IndexCreate creates a new index with the supplied mapping.
//
// The internal index created from the alias and a timestamp. If 'force'
// is false, no index is created when the alias exists. When it does not exist
// a new index is created and the alias is set.
//
// When force is true, there is no check for the alias and no alias is set for
// created index.
func IndexCreate(es *elasticsearch.Client, alias, mapping string, withAlias bool) (indexName string, err error) {
	if withAlias {
		storedIndexName, aliasErr := AliasGet(es, alias)
		if aliasErr != nil && !errors.Is(aliasErr, ErrAliasNotFound) {
			return "", aliasErr
		}

		if storedIndexName != "" {
			return storedIndexName, ErrIndexAlreadyCreated
		}
	}

	indexName = fmt.Sprintf("%s_%s", alias, time.Now().Format("20060102150405.999"))

	res, err := es.Indices.Create(
		indexName,
		es.Indices.Create.WithBody(strings.NewReader(mapping)),
		es.Indices.Create.WithWaitForActiveShards("1"),
	)
	if err != nil {
		return "", fmt.Errorf("elastic Indices.Create: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		return "", GetErrorType(res.Body).Error()
	}

	if withAlias {
		err = AliasCreate(es, alias, indexName)
		if err != nil {
			return indexName, err
		}
	}

	return indexName, nil
}

// IndexDelete delete the index from ElasticSearch.
//
// If the error does not exist an ErrIndexNotExist error is returned.
func IndexDelete(es *elasticsearch.Client, indexName string) error {
	res, err := es.Indices.Delete(
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

func IndexExists(es *elasticsearch.Client, indexName string) error {
	res, err := es.Indices.Exists(
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

// IndexSwitch updates the alias with the new index, and deletes the old index.
//
// When you only want to swith the alias use UpdateAlias().
func IndexSwitch(es *elasticsearch.Client, alias, newIndexName string, deleteOldIndex bool) (oldIndexName string, err error) {
	oldIndexName, err = AliasUpdate(es, alias, newIndexName)
	if err != nil {
		return "", err
	}

	if deleteOldIndex {
		if deleteErr := IndexDelete(es, oldIndexName); deleteErr != nil {
			return oldIndexName, deleteErr
		}
	}

	return oldIndexName, nil
}
