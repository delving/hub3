package sparql

import (
	"context"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/hashicorp/go-retryablehttp"
)

// TODO(kiivihal): finish implementation
func Harvest(ctx context.Context, repo *Repo, query string) (responses []*responseWithContext, err error) {
	// repo.queryRaw(query)

	// uri := fmt.Sprintf("%s?format=json&query=%s", endpoint, query)
	// log.Printf("uri: %s", uri)
	// resp, err := client.Get(uri)
	// if err != nil {
	// return responses, err
	// }
	// defer resp.Body.Close()

	// bindings, err := NewResponse(resp.Body)
	// if err != nil {
	// return responses, err
	// }

	// for _, b := range bindings.Results.Bindings {
	// url, err := b.S.asSubject()
	// if err != nil {
	// return responses, err
	// }
	// log.Printf("subject: %s", url)
	// }
	// g, ctx := errgroup.WithContext(ctx)
	// subjects := make(chan string)

	// Produce
	// g.Go(func() error {
	// defer close(subjects)

	// for _, urn := range getUrns {
	// select {
	// case <-ctx.Done():
	// return ctx.Err()
	// case urns <- urn:
	// }
	// }

	// return nil
	// })

	// responses := make(chan io.ReadCloser)

	// // Map
	// nWorkers := 4
	// workers := int32(nWorkers)
	// for i := 0; i < nWorkers; i++ {
	// g.Go(func() error {
	// defer func() {
	// // Last one out closes shop
	// if atomic.AddInt32(&workers, -1) == 0 {
	// close(graphs)
	// }
	// }()

	// for urn := range urns {
	// rdf, err := fb.GetRemoteWebResource(urn, "")
	// if err != nil {
	// return fmt.Errorf("unable to retrieve urn; %w", err)
	// }

	// select {
	// case <-ctx.Done():
	// return ctx.Err()
	// case graphs <- rdf:
	// }
	// }
	// return nil
	// })
	// }

	// // Reduce
	// g.Go(func() error {
	// for graph := range graphs {
	// if graph != nil {

	// defer graph.Close()
	// if err := fb.Graph.Parse(graph, "text/turtle"); err != nil {
	// return fmt.Errorf("unable to parse urn RDF; %w", err)
	// }
	// }
	// }

	// return nil
	// })

	// if err := g.Wait(); err != nil {
	// return responses, err
	// }

	return responses, nil
}

func getRecord(client *retryablehttp.Client, endpoint string, subject rdf.Subject) (*responseWithContext, error) {
	return nil, nil
}
