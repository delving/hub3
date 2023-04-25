package config

type Bulk struct {
	DBPath string `json:"dbPath,omitempty"`
	Minio  struct {
		Host     string `json:"host,omitempty"`
		Bucket   string `json:"bucket,omitempty"`
		UserName string `json:"userName,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"minio,omitempty"`
}
