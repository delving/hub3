package ead

// import (
// "bytes"
// "context"
// "log"
// "net/http"
// "os"
// "strconv"
// "sync"

// c "github.com/delving/hub3/config"
// "github.com/delving/hub3/hub3/server/http/handlers"
// "github.com/go-chi/render"
// "github.com/kiivihal/goharvest/oai"
// "github.com/pkg/errors"
// "golang.org/x/sync/errgroup"
// "golang.org/x/sync/semaphore"
// )

// type Sync struct {
// URL           string   `json:"url"`
// Spec          string   `json:"spec"`
// Prefix        string   `json:"prefix"`
// From          string   `json:"from"`
// Until         string   `json:"until"`
// InProgress    bool     `json:"inProgress"`
// Scheduled     int      `json:"scheduled"`
// Processed     int      `json:"processed"`
// ClevelsStored int      `json:"clevelsStored"`
// Errors        []string `json:"errors"`
// ErrorCount    int      `json:"errorCount"`
// Workers       int      `json:"workers"`
// sync.Mutex
// }

// var sem = semaphore.NewWeighted(int64(1))
// var eadSync *Sync

// func syncEAD(w http.ResponseWriter, r *http.Request) {
// if sem.TryAcquire(int64(1)) {
// if eadSync == nil {
// eadSync = &Sync{
// Workers: 4,
// }
// }

// params := r.URL.Query()

// url := params.Get("url")
// prefix := params.Get("prefix")

// if url == "" || prefix == "" {
// render.JSON(w, r, eadSync)
// sem.Release(int64(1))

// return
// }

// es := Sync{Workers: eadSync.Workers}

// for k := range params {
// switch k {
// case "from":
// es.From = params.Get(k)
// case "until":
// es.Until = params.Get(k)
// case "url":
// es.URL = params.Get(k)
// case "prefix":
// es.Prefix = params.Get(k)
// case "workers":
// wrks, err := strconv.Atoi(params.Get(k))
// if err != nil {
// http.Error(w, err.Error(), http.StatusBadRequest)
// return
// }

// if wrks > 8 {
// log.Printf("enforcing max workers limit of 8 instead of: %d", wrks)
// wrks = 8
// }

// es.Workers = wrks
// }
// }

// eadSync = &es
// eadSync.InProgress = true

// go func() {
// defer sem.Release(int64(1))
// log.Printf("aqcuiring semaphore lock")

// ctx := context.Background()

// err := harvestEAD(ctx, eadSync)
// if err != nil {
// eadSync.Errors = append(eadSync.Errors, err.Error())
// log.Printf("harvesting error: %s", err)
// }

// eadSync.InProgress = false

// log.Printf("releasing semaphore lock")
// }()
// }

// render.JSON(w, r, eadSync)
// }

// func harvestEAD(ctx context.Context, s *Sync) error {
// log.Printf("start ead sync harvest")

// s.InProgress = true

// g, ctx := errgroup.WithContext(ctx)
// ids := make(chan string, 1000)

// g.Go(func() error {
// defer close(ids)

// req := (&oai.Request{
// BaseURL:        s.URL,
// Verb:           "ListIdentifiers",
// Set:            s.Spec,
// MetadataPrefix: s.Prefix,
// })
// if s.From != "" {
// req.From = s.From
// }
// if s.Until != "" {
// req.Until = s.Until
// }

// req.HarvestIdentifiers(func(header *oai.Header) {
// select {
// case ids <- header.Identifier:
// s.Lock()
// s.Scheduled++
// s.Unlock()
// case <-ctx.Done():
// // ctx.Err()
// }
// })
// return nil
// })

// err := os.MkdirAll(c.Config.EAD.CacheDir, os.ModePerm)
// if err != nil {
// log.Printf("unable to create EAD cache dir; %#v", err)
// return err
// }

// for i := 0; i < s.Workers; i++ {
// g.Go(func() error {
// for id := range ids {
// id := id
// err := storeRecord(id, s.Prefix, s.URL, s)
// if err != nil {
// s.Lock()
// s.Errors = append(s.Errors, err.Error())
// log.Printf("ead %s processing error: %#v", id, err)
// s.ErrorCount++
// s.Unlock()
// }
// log.Printf("EAD OAI-PMH: %d/%d (errors: %d)", s.Processed, s.Scheduled, len(s.Errors))
// }
// return nil
// })
// }

// go func() {
// err := g.Wait()
// if err != nil {
// log.Printf("unable to wait for all goroutines to finish; %#v", err)
// }
// }()

// // Check whether any of the goroutines failed. Since g is accumulating the
// // errors, we don't need to send them (or check for them) in the individual
// // results sent on the channel.
// if err := g.Wait(); err != nil {
// log.Println(err)
// return err
// }

// return nil
// }

// func storeRecord(identifier, prefix, url string, s *Sync) error {
// req := (&oai.Request{
// BaseURL:        url,
// Verb:           "GetRecord",
// MetadataPrefix: prefix,
// Identifier:     identifier,
// })

// var err error

// req.Harvest(func(r *oai.Response) {
// rawBody := r.GetRecord.Record.Metadata.Body
// headerSize := int64(len(rawBody))
// b := bytes.NewReader(rawBody)
// cfg, localErr := processEAD(b, headerSize, "", handlers.BulkProcessor(), false)
// if localErr != nil {
// log.Printf("problem procesing oai.Response: %#v", r)
// err = errors.Wrapf(localErr, "unable to process EAD for %s", req.GetFullURL())
// s.Lock()
// s.Processed++
// s.Unlock()
// return
// }
// s.Lock()
// s.Processed++
// s.ClevelsStored += int(cfg.Counter.GetCount())
// s.Unlock()
// })

// return err
// }
