package pid

type MockStore struct {
	DataByID         map[string][]*PID
	DataByRdfSubject map[string]*PID
}

func NewMockStore() *MockStore {
	return &MockStore{
		DataByID:         make(map[string][]*PID),
		DataByRdfSubject: make(map[string]*PID),
	}
}

func (m *MockStore) Get(id, subject string) (StoreResponse, error) {
	known := []*PID{}

	if resp, exists := m.DataByID[id]; exists {
		known = append(known, resp...)
	}
	if resp, exists := m.DataByRdfSubject[subject]; exists {
		known = append(known, resp)
	}

	return NewStoreResponse(known), nil
}

func (m *MockStore) Save(pids ...*PID) error {
	for _, pid := range pids {
		m.DataByRdfSubject[pid.ID] = pid

		var found bool
		known, ok := m.DataByID[pid.ExternalID]
		if ok {
			for idx, p := range known {
				if p.ID == pid.ID && p.ExternalID == pid.ExternalID {
					known[idx] = pid
					found = true
				}
			}
		}

		if !found {
			known = append(known, pid)
		}
		m.DataByID[pid.ID] = known

	}
	return nil
}
