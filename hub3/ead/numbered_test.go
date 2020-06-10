// Copyright 2017 Delving B.V.
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

package ead_test

import (
	"context"
	"testing"

	"github.com/delving/hub3/hub3/ead"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestNumbered(t *testing.T) {
	is := is.New(t)

	dsc := new(ead.Cdsc)
	err := parseUtil(dsc, "ead.0x.xml")
	is.NoErr(err)

	cfg := ead.NewNodeConfig(context.Background())
	_, nodeCount, err := dsc.NewNodeList(cfg)
	is.NoErr(err)

	is.Equal(int(nodeCount), 7)
}
