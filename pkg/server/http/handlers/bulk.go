package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/delving/hub3/hub3"
	"github.com/delving/hub3/hub3/index"
	"github.com/gammazero/workerpool"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/olivere/elastic/v7"
)

var bp *elastic.BulkProcessor
var bps *elastic.BulkProcessorService
var wp *workerpool.WorkerPool
var ctx context.Context

func init() {
	wp = workerpool.New(10)
}

// RegisterBulkIndexer registers routes for the BulkIndexer API.
func RegisterBulkIndexer(r chi.Router) {
	// Narthex endpoint
	r.Post("/api/rdf/bulk", bulkAPI)
	// TODO remove later
	r.Post("/api/index/bulk", bulkAPI)
}

// BulkProcessor creates a new BulkProcessor.
func BulkProcessor() *elastic.BulkProcessor {
	if bp != nil {
		return bp
	}
	var err error
	ctx = context.Background()
	bps := index.CreateBulkProcessorService()
	bp, err = bps.Do(ctx)
	if err != nil {
		log.Fatalf("Unable to start BulkProcessor: %#v", err)
	}
	return bp
}

// bulkApi receives bulkActions in JSON form (1 per line) and processes them in
// ingestion pipeline.
func bulkAPI(w http.ResponseWriter, r *http.Request) {
	response, err := hub3.ReadActions(ctx, r.Body, BulkProcessor(), wp)
	if err != nil {
		log.Println("Unable to read actions")
		errR := ErrRender(err)
		// todo fix errr renderer for better narthex consumption.
		_ = errR.Render(w, r)
		render.Render(w, r, errR)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
	return
}
