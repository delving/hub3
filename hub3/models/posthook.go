package models

import (
	"bytes"
	"fmt"
	"log"

	r "github.com/deiu/rdf2go"
	c "github.com/delving/rapid-saas/config"
	"github.com/gammazero/workerpool"
	"github.com/parnurzeal/gorequest"
)

// PostHookJob  holds the info for building a crea
type PostHookJob struct {
	Graph   *r.Graph
	Spec    string
	Deleted bool
	Subject string
}

// PostHookJobFactory can be used to fire off PostHookJob jobs
type PostHookJobFactory struct {
	Spec string
	wp   *workerpool.WorkerPool
}

// NewPostHookJob creates a new PostHookJob and populates the rdf2go Graph
func NewPostHookJob(g *r.Graph, spec string, delete bool, subject string) *PostHookJob {
	return &PostHookJob{g, spec, delete, subject}
}

// Valid determines if the posthok is valid to apply.
func (ph PostHookJob) Valid() bool {
	return ProcessSpec(ph.Spec)
}

// ProcessSpec determines if a PostHookJob should be applied for a specific spec
func ProcessSpec(spec string) bool {
	for _, e := range c.Config.PostHook.ExcludeSpec {
		if e == spec {
			return false
		}
	}
	return true
}

// ApplyPostHookJob applies the PostHookJob to all the configured URLs
func ApplyPostHookJob(ph *PostHookJob) {
	//time.Sleep(100 * time.Millisecond)
	for _, u := range c.Config.PostHook.URLs {
		err := ph.Post(u)
		if err != nil {
			log.Printf("Unable to send %s to %s", ph.Subject, u)
		}
	}
}

// Post sends json-ld to the specified endpointt
func (ph PostHookJob) Post(url string) error {
	request := gorequest.New()
	if ph.Deleted {
		log.Printf("Deleting via posthook: %s", ph.Subject)
		deleteURL := fmt.Sprintf("%s/delete", url)
		req := request.Post(deleteURL).
			Query(fmt.Sprintf("id=%s", ph.Subject))
			//Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError, http.StatusRequestTimeout).
		log.Printf("%v", req)
		log.Printf("%v", req.RawString)
		rsp, body, errs := req.End()
		log.Printf("%#v -> %#v", rsp, body)
		if errs != nil {
			log.Printf("Unable to delete: %#v", errs)
			return errs[0]
		}
		return nil
	}
	//log.Printf("Storing %s", ph.Subject)
	json, err := ph.String()
	if err != nil {
		return err
	}
	_, _, errs := request.Post(url).
		Type("text").
		Send(json).
		//Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError, http.StatusRequestTimeout).
		End()
	if errs != nil {
		log.Printf("Unable to store: %#v", errs)
		return errs[0]
	}
	log.Printf("Stored %s", ph.Subject)
	return nil
}

// Bytes returns the PostHookJob as an JSON-LD bytes.Buffer
func (ph PostHookJob) Bytes() (bytes.Buffer, error) {
	var b bytes.Buffer
	err := ph.Graph.Serialize(&b, "application/json-ld")
	if err != nil {
		return b, err
	}
	return b, nil
}

// Bytes returns the PostHookJob as an JSON-LD string
func (ph PostHookJob) String() (string, error) {
	b, err := ph.Bytes()
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
