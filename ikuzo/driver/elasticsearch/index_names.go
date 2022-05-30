package elasticsearch

import (
	"fmt"
	"strings"
)

type IndexNames struct{}

func (in IndexNames) FragmentIndexName(orgID string) string {
	return fmt.Sprintf("%s_frag", in.GetIndexName(orgID))
}

func (in IndexNames) GetSuggestIndexName(orgID string) string {
	return fmt.Sprintf("%s_suggest", in.GetIndexName(orgID))
}

// GetIndexName returns the lowercased indexname.
// This inforced correct behavior when creating an index in ElasticSearch.
func (in IndexNames) GetIndexName(orgID string) string {
	return strings.ToLower(orgID) + "v2"
}

func (in IndexNames) GetV1IndexName(orgID string) string {
	return strings.ToLower(orgID) + "v1"
}

func (in IndexNames) GetDigitalObjectIndexName(orgID, suffix string) string {
	if suffix == "" {
		return in.GetIndexName(orgID)
	}

	return strings.ToLower(orgID) + "v2-" + strings.ToLower(suffix)
}
