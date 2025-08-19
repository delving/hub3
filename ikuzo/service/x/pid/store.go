package pid

import "sort"

type Store interface {
	Get(id, subject string) (resp StoreResponse, err error)
	Save(pids ...*PID) (err error)
}

type StoreResponse struct {
	CurrentPID *PID
	OtherIDs   []*PID
}

func (sr StoreResponse) GetSupportedProfiles() (profiles []*ProfileInfo) {
	if sr.CurrentPID.DataSet == "brabantse-gebouwen" {
		profiles = append(profiles, &ProfileInfo{
			URI:   "urn:profile/schema",
			Token: "schema",
		})
	}

	profiles = append(profiles, &ProfileInfo{
		URI:   "urn:profile/edm",
		Token: "edm",
	})

	return profiles
}

func NewStoreResponse(pids []*PID) (resp StoreResponse) {
	if len(pids) == 0 {
		return resp
	}
	if len(pids) == 1 {
		resp.CurrentPID = pids[0]
		return resp
	}

	sort.Slice(pids, func(i, j int) bool {
		return pids[i].ModifiedAt.Before(pids[j].ModifiedAt)
	})

	// find empty replaced by
	for _, pid := range pids {
		if pid.ReplacedBy == "" {
			resp.CurrentPID = pid
		}
	}

	// otherwise default to the oldest entry
	if resp.CurrentPID == nil {
		resp.CurrentPID = pids[0]
	}

	for _, pid := range pids {
		if pid == resp.CurrentPID {
			continue
		}
		resp.OtherIDs = append(resp.OtherIDs, pid)
	}

	return resp
}

func (resp StoreResponse) Empty() bool {
	return resp.CurrentPID == nil && len(resp.OtherIDs) == 0
}

func (resp StoreResponse) Tombstone() bool {
	if resp.CurrentPID == nil {
		return false
	}

	return resp.CurrentPID.Tombstone
}
