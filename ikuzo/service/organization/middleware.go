package organization

import (
	"context"
	"log"
	"net/http"

	d "github.com/delving/hub3/ikuzo/domain"
)

// ResolveOrgByDomain adds a domain.OrganizationID to the context by domain.
// If the domain cannot be resolved it will return a http.StatusBadRequest.
//
// GetOrganizationID can be used to retrieve the domain.OrganizationID from
// the request context.
func (s *Service) ResolveOrgByDomain(next http.Handler) http.Handler {
	domains := map[string]*d.Organization{}

	orgs, err := s.Filter(context.Background())
	if err != nil {
		log.Println("unable to get organizations")
	}

	var defaultOrgID *d.Organization

	for _, org := range orgs {
		for _, domain := range org.Config.Domains {
			domains[domain] = org
			if org.Config.Default {
				defaultOrgID = org
			}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Hostname()
		if domain == "" {
			domain = r.Host
		}

		org, ok := domains[domain]
		if !ok {
			if defaultOrgID != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("no organization available for this domain"))
				return
			}
		}
		// w.Header().Set("ORG-ID", org.RawID())
		r = d.SetOrganization(r, org)
		next.ServeHTTP(w, r)
	})
}
