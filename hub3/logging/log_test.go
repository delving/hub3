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

package logging_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bitbucket.org/delving/rapid/config"
	"bitbucket.org/delving/rapid/hub3/logging"
)

var _ = Describe("Log", func() {

	Describe("When initialised", func() {

		Context("and no sentry dsn is present", func() {

			Config.Logging.SentryDSN = ""
			log := logging.NewLogger()

			It("should return a logrus logger", func() {
				Expect(log).ToNot(BeNil())
				Expect(fmt.Sprintf("%T", log)).To(Equal("*logrus.Logger"))

			})

			It("should not have started raven", func() {
				Expect(log.Hooks).To(BeEmpty())
			})
		})

		Context("when a Sentry DSN is present", func() {

			Config.Logging.SentryDSN = "https://0a833ad240ba4aea847d70f07a0babbd:5a2feb29b4c441a5bcd7f182e0579600@sentry.io/218042"
			l := logging.NewLogger()

			It("logrus should have a Senty hook", func() {
				Expect(l.Hooks).ToNot(BeEmpty())

			})
		})
	})

})
