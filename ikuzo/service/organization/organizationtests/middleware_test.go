package organizationtests_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/organization/organizationtests"
	"github.com/go-chi/chi"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestService_ResolveOrgByDomain(t *testing.T) {
	is := is.New(t)

	svc := organizationtests.NewTestOrganizationService()

	r := chi.NewRouter()
	r.Use(svc.ResolveOrgByDomain)
	r.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
		// GetOrganizationID(r)
		w.Header().Set("X-Test", "yes")
		w.Write([]byte("bye"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	demoOrgID, err := domain.NewOrganizationID("demo")
	is.NoErr(err)

	// todo parse basename from request
	org := domain.Organization{
		ID: demoOrgID,
		Config: domain.OrganizationConfig{
			Domains: []string{ts.URL},
		},
	}

	err = svc.Put(context.TODO(), &org)
	is.NoErr(err)

	req, err := http.NewRequest("GET", ts.URL+"/hi", nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("should give badrequest, got; %d", resp.StatusCode)
	}

	// orgID, err := organization.GetOrganizationID(resp.Request)
	// is.NoErr(err)
	// is.Equal(orgID, "")
}
