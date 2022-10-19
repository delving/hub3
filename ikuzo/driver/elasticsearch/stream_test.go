package elasticsearch

import (
	"testing"
)

func TestStreamFragments(t *testing.T) {
	// is := is.New(t)

	// testSets := []string{
	// "mauritshuis",
	// // "museum-de-fundatie",
	// // "rijksmuseum",
	// // "catharijneconvent",
	// // "stedelijk-museum-schiedam",
	// // "van-abbe-museum",
	// // "museum-belvedere",
	// // "rijksakademie",
	// // "moderne-kunst-museum-deventer",
	// }

	// query := elastic.NewBoolQuery()
	// for _, setSpec := range testSets {
	// query.Should(elastic.NewTermQuery("meta.spec", setSpec))
	// }

	// cfg := Config{Urls: []string{"http://localhost:9200"}}
	// client, err := NewClient(&cfg)
	// is.NoErr(err)

	// streamConfig := &StreamConfig{
	// Query:      query,
	// IndexNames: []string{"dcnv2_20200526062245.616"},
	// }

	// printFunc := func(hit *elastic.SearchHit) error {
	// log.Printf("hit: %s", hit.Id)
	// return nil
	// }

	// seen, streamErr := client.Stream(context.Background(), streamConfig, printFunc)
	// is.NoErr(streamErr)
	// is.Equal(seen, 856)
}
