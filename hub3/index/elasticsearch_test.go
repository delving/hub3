// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package index_test

import (
	"fmt"

	"github.com/delving/rapid-saas/hub3/index"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Elasticsearch", func() {

	Describe("CreateClient", func() {

		Context("When initialised", func() {

			It("Should return an elastic client", func() {
				client := index.ESClient()
				Expect(client).ToNot(BeNil())
				Expect(fmt.Sprintf("%T", client)).To(Equal("*elastic.Client"))
			})
		})
	})

	Describe("CustomRetrier", func() {

		Context("When initialised", func() {

			It("should return a Retrier", func() {
				Expect(index.NewCustomRetrier()).ToNot(BeNil())
			})
		})
	})
})
