package sitemap

type Config struct {
	OrgID   string `json:"-"`
	ID      string `json:"id"`
	BaseURL string `json:"baseURL"`
	Filters string `json:"filters"` // qf and q URL params
}
