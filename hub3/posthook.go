package hub3

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	r "github.com/deiu/rdf2go"
	c "github.com/delving/rapid-saas/config"
	"github.com/parnurzeal/gorequest"
)

// PostHook  holds the info for building a crea
type PostHook struct {
	Graph   *r.Graph
	Spec    string
	Deleted bool
	Subject string
}

// NewPostHook creates a new posthook and populates the rdf2go Graph
func NewPostHook(g *r.Graph, spec string, delete bool, subject string) *PostHook {
	return &PostHook{g, spec, delete, subject}
}

// Valid determines if the posthok is valid to apply.
func (ph PostHook) Valid() bool {
	for _, e := range c.Config.PostHook.ExcludeSpec {
		if e == ph.Spec {
			return false
		}
	}
	return true
}

// ApplyPostHook applies the posthook to all the configured URLs
func ApplyPostHook(ph *PostHook) {
	//time.Sleep(100 * time.Millisecond)
	for _, u := range c.Config.PostHook.URLs {
		err := ph.Post(u)
		if err != nil {
			log.Printf("Unable to store %s to %s", ph.Subject, u)
		}
	}
}

// Post sends json-ld to the specified endpointt
func (ph PostHook) Post(url string) error {
	json, err := ph.String()
	if err != nil {
		return err
	}
	request := gorequest.New()
	if ph.Deleted {
		deleteURL := fmt.Sprintf("%/delete", url)
		_, _, errs := request.Delete(deleteURL).
			Query(fmt.Sprintf("id=%s", ph.Subject)).
			Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError, http.StatusRequestTimeout).
			End()
		if errs != nil {
			log.Printf("Unable to delete: %#v", errs)
			return errs[0]
		}
		return nil
	}
	//log.Printf("Storing %s", ph.Subject)
	_, _, errs := request.Post(url).
		Type("text").
		Send(json).
		//Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError, http.StatusRequestTimeout).
		End()
	if errs != nil {
		log.Printf("Unable to store: %#v", errs)
		return errs[0]
	}
	//log.Printf("Stored %s", ph.Subject)
	return nil
}

// Bytes returns the posthook as an JSON-LD bytes.Buffer
func (ph PostHook) Bytes() (bytes.Buffer, error) {
	var b bytes.Buffer
	err := ph.Graph.Serialize(&b, "application/json-ld")
	if err != nil {
		return b, err
	}
	return b, nil
}

// Bytes returns the posthook as an JSON-LD string
func (ph PostHook) String() (string, error) {
	b, err := ph.Bytes()
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
