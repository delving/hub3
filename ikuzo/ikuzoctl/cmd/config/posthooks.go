package config

import (
	"github.com/delving/hub3/ikuzo/service/x/bulk"
	"github.com/delving/hub3/ikuzo/storage/x/ginger"
)

type PostHook struct {
	Name        string   `json:"name"`
	ExcludeSpec []string `json:"excludeSpec"`
	URL         string   `json:"url"`
	OrgID       string   `json:"orgID"`
	APIKey      string   `json:"apiKey"`
}

// nolint:unparam // in the future other posthook services can return errors
func (cfg *Config) getPostHookServices() ([]bulk.PostHookService, error) {
	svc := []bulk.PostHookService{}

	for _, ph := range cfg.PostHooks {
		if ph.Name == "ginger" && ph.URL != "" {
			svc = append(
				svc,
				ginger.NewPostHook(
					ph.OrgID,
					ph.URL,
					ph.APIKey,
					ph.ExcludeSpec...,
				),
			)
		}
	}

	return svc, nil
}
