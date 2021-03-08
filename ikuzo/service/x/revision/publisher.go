package revision

type PublisherStats struct {
	DatasetHash    string
	RepoHash       string
	TotalPublished int
	New            int
	Updated        int
	Deleted        int
}
