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

package handlers_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/delving/hub3/config"
	. "github.com/delving/hub3/hub3/server/http/handlers"
)

var _ = Describe("Cache", func() {

	Context("when preparing a request", func() {
		config.InitConfig()
		config.Config.Cache.Enabled = true
		config.Config.Cache.CacheDomain = "acpt.nationaalarchief.nl"
		config.Config.Cache.StripPrefix = true
		config.Config.Cache.APIPrefix = "/api/cache/http"
		domain := config.Config.Cache.CacheDomain

		It("should strip the APIPrefix from the request url", func() {
			url := "http://localhost:3000/api/cache/http/gaf/search/ead/F1270773"
			req, err := http.NewRequest("POST", url, nil)
			Expect(err).ToNot(HaveOccurred())
			cacheKey, err := PrepareCacheRequest(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheKey).ToNot(BeEmpty())
			Expect(domain).ToNot(BeEmpty())
			Expect(req.URL.Hostname()).ToNot(ContainSubstring("localhost"))
			Expect(req.URL.Scheme).To(Equal("https"))
			Expect(req.URL.Path).ToNot(HavePrefix(config.Config.Cache.APIPrefix))
			Expect(req.RequestURI).To(BeEmpty())
			fmt.Println(req.URL.Hostname(), domain)

		})

		It("should override the configured domain", func() {
			url := "http://localhost:3000/api/cache/http/gaf/search/ead/F1270773?domain=https://custom.io:9002"
			req, err := http.NewRequest("POST", url, nil)
			Expect(req.URL.Query().Get("domain")).To(Equal("https://custom.io:9002"))
			Expect(err).ToNot(HaveOccurred())
			_, err = PrepareCacheRequest(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(req.URL.Query().Get("domain")).To(Equal(""))
			Expect(req.URL.Host).To(ContainSubstring("custom.io:9002"))
		})

	})

})
