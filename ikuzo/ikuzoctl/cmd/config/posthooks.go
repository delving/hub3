package config

import (
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/storage/x/ginger"
)

type PostHook struct {
	Name        string   `json:"name"`
	ExcludeSpec []string `json:"excludeSpec"`
	URL         string   `json:"url"`
	OrgID       string   `json:"orgID"`
	APIKey      string   `json:"apiKey"`
	UserName    string   `json:"userName"`
	Password    string   `json:"password"`
	CustomWait  int      `json:"customWait"`
}

// nolint:unparam // in the future other posthook services can return errors
func (cfg *Config) getPostHookServices() ([]domain.PostHookService, error) {
	svc := []domain.PostHookService{}

	for _, ph := range cfg.PostHooks {
		if ph.Name == "ginger" && ph.URL != "" {
			svc = append(
				svc,
				ginger.NewPostHook(
					ph.OrgID,
					ph.URL,
					ph.APIKey,
					ph.CustomWait,
					ph.ExcludeSpec...,
				),
			)
		}
	}

	return svc, nil
}
