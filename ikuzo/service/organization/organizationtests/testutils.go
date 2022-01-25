package organizationtests

import (
	"bytes"
	"fmt"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
	"github.com/spf13/viper"
)

var tomlExample = []byte(`
[org.playground]
domains = ["localhost:3000"]
default = false

[org.hub3]
domains = ["localhost:3001"]
customID = "hub3"
default = true

[org.hub3.elasticsearch]
indexTypes = ["v2"]
minimumShouldMatch = "2<70%"
# indexName = "hub3"
# digitalObjectSuffix = "scans"

[org.hub3.elasticsearch.defaults]
maxTreeSize = 251
facetSize = 50
limit = 20
maxLimit = 1000

[[org.hub3.sitemaps]]
id = "all"
baseURL = "http://localhost:3001"
filters = "meta.tags:narthex"
`)

type testConfig struct {
	Org map[string]domain.OrganizationConfig `json:"org"`
}

// NewTestOrganizationService should only be used in tests.
// When it is called from a compiled binary outside the project
// it will panic.
func NewTestOrganizationService() *organization.Service {
	viper.SetConfigType("toml")

	err := viper.ReadConfig(bytes.NewBuffer(tomlExample))
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var cfg testConfig
	if err = viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("fatal error viper: %w", err))
	}

	store := memory.NewOrganizationStore()

	svc, err := organization.NewService(store)
	if err != nil {
		panic(fmt.Errorf("fatal error organization service: %w", err))
	}

	if err := svc.AddOrgs(cfg.Org); err != nil {
		panic(fmt.Errorf("fatal error organization service: %w", err))
	}

	return svc
}
