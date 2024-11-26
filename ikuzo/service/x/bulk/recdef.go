package bulk

import "fmt"

type RecDefResolver struct {
	m               map[string]string
	defaultRecDefID string
}

func (rdr *RecDefResolver) Resolve(recDefID string) (string, error) {
	uri, ok := rdr.m[recDefID]
	if !ok {
		return "", fmt.Errorf("no rec-def configured for %q", recDefID)
	}

	return uri, nil
}
