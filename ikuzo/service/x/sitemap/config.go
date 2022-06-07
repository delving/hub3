package sitemap

import "strings"

type Config struct {
	ID            string   `json:"id"`
	BaseURL       string   `json:"baseURL"`
	Query         string   `json:"query"`
	Filters       []string `json:"filters"` // qf and q URL params
	OrgID         string   `json:"-"`
	ExcludedSpecs []string `json:"excludedSpecs"`
}

func (c *Config) IsExcludedSpec(spec string) bool {
	for _, excluded := range c.ExcludedSpecs {
		if strings.EqualFold(spec, excluded) {
			return true
		}
	}

	return false
}
