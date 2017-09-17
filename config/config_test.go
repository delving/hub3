package config_test

import (
	. "bitbucket.org/delving/rapid/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	Describe("initialisation", func() {

		Context("without calling initConfig", func() {

			It("should not be initialised", func() {
				Expect(Config.OrgID).To(BeEmpty())
				Expect(Config.HTTP.Port).To(Equal(0))
			})

		})

		Context("when calling initConfig", func() {

			It("should be initialised with defaults", func() {
				InitConfig()
				Expect(Config.HTTP.Port).To(Equal(3001))
				Expect(Config.OrgID).ToNot(BeEmpty())
			})

		})

		Context("when setting a config value", func() {

			It("should be available in the global scope", func() {
				Expect(Config.Logging.SentryDSN).To(BeEmpty())
				Config.Logging.SentryDSN = "test"
				Expect(Config.Logging.SentryDSN).ToNot(BeEmpty())
			})
		})
	})
})
