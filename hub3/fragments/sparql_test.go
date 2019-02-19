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
package fragments

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sparql", func() {

	Describe("When reading a io.Reader", func() {

		Context("and lines are counted", func() {

			It("it should count all lines", func() {
				fourLines := "1\n2\n3\n4\n"
				Expect(lineCounter(strings.NewReader(fourLines))).To(Equal(4))
			})

			It("it should count last line with no linefeed", func() {
				fourLines := "1\n2\n3\n4"
				Expect(lineCounter(strings.NewReader(fourLines))).To(Equal(4))
			})

		})
	})

})
