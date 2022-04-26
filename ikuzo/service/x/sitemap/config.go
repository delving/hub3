package sitemap

type Config struct {
	ID      string `json:"id"`
	BaseURL string `json:"baseURL"`
	Filters string `json:"filters"` // qf and q URL params
	OrgID   string `json:"-"`
}
