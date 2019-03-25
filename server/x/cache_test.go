package server_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/delving/rapid-saas/config"
	. "github.com/delving/rapid-saas/server"
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
