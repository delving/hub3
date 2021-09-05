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
	"fmt"
	"io"
	"strings"

	"github.com/tidwall/gjson"
)

type Alias struct {
	client *Client
}

func (c *Client) Alias() Alias {
	return Alias{client: c}
}

// Create creates an alias for the given indexName.
//
// When the index does not exist ErrIndexNotExists is returned
// When the alias is already defined, ErrAliasExist is returned
func (a *Alias) Create(alias, indexName string) error {
	res, conErr := a.client.index.Indices.PutAlias(
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

// Get returns the indexName for the given alias.
//
// When the alias is not found an ErrAliasNotFound error is returned.
func (a *Alias) Get(name string) (indexName string, err error) {
	res, conErr := a.client.index.Indices.GetAlias(
		a.client.index.Indices.GetAlias.WithName(name),
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

// Update removes the alias if it exists from another index and creates a new one linked to indexName.
//
// It returns an error when the alias cannot be updated. It returns the old indexName that is removed
func (a *Alias) Update(name, indexName string) (oldIndexName string, err error) {
	oldIndexName, err = a.Get(name)
	if err != nil {
		return oldIndexName, err
	}

	res, conErr := a.client.index.Indices.UpdateAliases(
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
				name,
				indexName,
				name,
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

// Delete removes the alias from the index it is linked to.
//
// When the alias does not it exist it will return a ErrAliasNotExist error.
//
// When the indexName is empty it will search for the indexName using GetAlias().
func (a *Alias) Delete(name, indexName string) error {
	res, conErr := a.client.index.Indices.DeleteAlias(
		[]string{indexName},
		[]string{name},
	)

	if conErr != nil {
		return conErr
	}

	defer res.Body.Close()

	if res.IsError() {
		et := GetErrorType(res.Body)
		a.client.log.Error().Err(et.Error()).Msgf("%#v", et)
		return et.Error()
	}

	return nil
}
