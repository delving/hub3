package config

type Bulk struct {
	DBPath        string `json:"dbPath,omitempty"`
	StoreRequests bool
	Minio         struct {
		Endpoint        string `json:"endpoint,omitempty"`
		AccessKeyID     string `json:"accessKeyID,omitempty"`
		SecretAccessKey string `json:"secretAccessKey,omitempty"`
		UseSSL          bool   `json:"useSSL,omitempty"`
		BucketName      string `json:"bucketName,omitempty"`
	} `json:"minio,omitempty"`
}
