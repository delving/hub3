package pid

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/delving/hub3/ikuzo/domain"
)

type Type int

const (
	Undefined Type = iota
	Ark
	DOI
	Handle
)

// Meta is the additional metadata that is stored with the PID
type Meta struct {
	OrgID   domain.OrganizationID `gorm:"not null"`
	DataSet string                `gorm:"not null"`
	HubID   string                `gorm:"not null"`
}

type PID struct {
	ID         string    `gorm:"primaryKey;not null"` // the rdf subject
	ExternalID string    `gorm:"not null;index"`      // the external PID ID or URI
	Type       Type      `gorm:"not null"`            // the PID type
	ReplacedBy string    `gorm:"foreignKey:ID"`       // the subject of the PID that is being replaced by
	Tombstone  bool      `gorm:"not null"`            // where pid is now deleted
	ModifiedAt time.Time `gorm:"not null"`            // last modified time
	IsShownAt  string    // the url where PID can be view as HTML
	Meta
}

func (p *PID) inferType() {
	if p.Type != Undefined {
		return
	}

	switch {
	case strings.HasPrefix(p.ExternalID, "ark:"):
		p.Type = Ark
	case isDOI(p.ExternalID):
		p.Type = DOI
	case isHandle(p.ExternalID):
		p.Type = Handle
	default:
		slog.Warn("pid: unable to infer PID type", "pid", p.ExternalID)
	}
}

func (p *PID) cleanID() error {
	pid, err := extractRelativePath(p.ExternalID)
	if err != nil {
		return fmt.Errorf("unable to extract relative path: %w", err)
	}

	p.ExternalID = pid
	return nil
}

func (p *PID) asPayload() SavePayload {
	return SavePayload{
		RDFSubject: p.ID,
		PID:        p.ExternalID,
		IsShownAt:  p.IsShownAt,
		Type:       p.Type,
		Meta:       p.Meta,
		Deleted:    p.Tombstone,
	}
}

func (p *PID) IsDifferent(other *PID) bool {
	if p.ID != other.ID {
		return true
	}
	if p.ExternalID != other.ExternalID {
		return true
	}
	if p.IsShownAt != other.IsShownAt {
		return true
	}
	if p.Tombstone != other.Tombstone {
		return true
	}
	if p.ReplacedBy != other.ReplacedBy {
		return true
	}
	if p.Type != other.Type {
		return true
	}

	return false
}

func isDOI(pid string) bool {
	pid = strings.TrimPrefix(pid, "doi:")

	doiRegex := regexp.MustCompile(`^10\.\d{4,9}/[-._;()/:\w]+$`)
	return doiRegex.MatchString(pid)
}

func isHandle(pid string) bool {
	handleRegex := regexp.MustCompile(`^hdl:.*$`)
	return handleRegex.MatchString(pid)
}

/*
* The flow for processing is as follows:
* - extract info from the graph and request
* - submit it to the task queue
* - process the task
* - if unknown save
* - if known and the same, do nothing
* - if known but subject different then create new pid and update old pid with ReplacedBy
* - if no externalID but known mark as tombstone
* - if no IsShownAt update known
* - function to save multiple pids at the same time
 */
func (s *Service) process(p SavePayload) (pids []*PID, err error) {
	store, err := s.GetStore(p.OrgID.String())
	if err != nil {
		return pids, fmt.Errorf("unable to get store: %w", err)
	}

	resp, err := store.Get(p.PID, p.RDFSubject)
	if err != nil && !errors.Is(err, ErrTombstone) {
		return pids, err
	}

	pid := p.asPID()

	// pid is unknown so return immediately
	if resp.Empty() {
		return append(pids, &pid), nil
	}

	// pid is identical to know pid so do nothing
	if !resp.CurrentPID.IsDifferent(&pid) {
		return pids, nil
	}

	// if the pid has a different subject then create a new pid and update the old pid
	if resp.CurrentPID.ID != pid.ID {
		resp.CurrentPID.ReplacedBy = pid.ID
		pids = append(pids, resp.CurrentPID)
		resp.CurrentPID = &pid
	}

	pids = append(pids, &pid)

	return pids, nil
}

// extractRelativePath extracts the relative path from a URI without the leading slash.
// For example, from "https://n2t.net/ark:/63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea"
// it will return "ark:/63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea"
func extractRelativePath(uri string) (string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	path := parsedURL.Path
	// Remove leading slash if it exists
	path = strings.TrimPrefix(path, "/")

	return path, nil
}
