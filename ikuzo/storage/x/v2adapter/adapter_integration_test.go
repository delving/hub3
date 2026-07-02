package v2adapter

import (
	"context"
	"encoding/json"
	"testing"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/domain/semantic"
	es "github.com/delving/hub3/ikuzo/driver/elasticsearch"
	elastic "github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"
)

// TestMain initializes the global config the fragments query builder
// depends on (e.g. ElasticSearch.OrgIDKey); without it ElasticQuery()
// emits a term query with an empty field name and ES returns 400.
func TestMain(m *testing.M) {
	c.InitConfig()
	m.Run()
}

// seedTestIndex recreates the index the adapter derives from the org in
// context ({orgID}v2 = "testorgv2") with the real v2 mapping, then indexes
// one document. The v2 mapping is required: the fragments query builder
// issues nested queries against resources.entries, which fail with
// "all shards failed" on a dynamically-mapped index.
func seedTestIndex(ctx context.Context, t *testing.T, client *elastic.Client) {
	t.Helper()

	const indexName = "testorgv2"

	// Drop stale test indices (wildcard also removes the timestamped
	// indices CreateDefaultMappings creates behind the alias).
	_, _ = client.DeleteIndex(indexName + "*").Do(ctx)

	logger := zerolog.Nop()
	driver, err := es.NewClient(&es.Config{
		Urls:   []string{"http://localhost:9200"},
		Logger: &logger,
		// expvar.Publish panics on duplicate registration when multiple
		// tests create a driver client in one process.
		DisableMetrics: true,
	})
	if err != nil {
		t.Skipf("Failed to create es driver client: %v", err)
	}

	orgCfg := domain.OrganizationConfig{}
	orgCfg.SetOrgID("testorg")
	orgCfg.ElasticSearch.IndexTypes = []string{"v2"}

	// withAlias=true: Indices.Create always creates a timestamped index;
	// the "testorgv2" name the adapter queries only exists as its alias.
	if _, err := driver.CreateDefaultMappings([]domain.OrganizationConfig{orgCfg}, true, false); err != nil {
		t.Skipf("Failed to create v2 mapping: %v", err)
	}

	const entryURI = "https://example.org/resource/test-doc-123"

	testDoc := fragments.FragmentGraph{
		Meta: &fragments.Header{
			HubID:    "test-doc-123",
			OrgID:    "testorg",
			DocType:  "graph",
			EntryURI: entryURI,
		},
		Resources: []*fragments.FragmentResource{
			{
				ID:    entryURI,
				Types: []string{"http://www.europeana.eu/schemas/edm/ProvidedCHO"},
			},
		},
	}
	testDoc.Semantic = testDoc.GenerateSemantic()

	docJSON, _ := json.Marshal(&testDoc)
	_, err = client.Index().
		Index(indexName).
		Id("test-doc-123").
		BodyJson(string(docJSON)).
		Refresh("true").
		Do(ctx)
	if err != nil {
		t.Skipf("Failed to index test document: %v", err)
	}
}

// TestV2SearchAdapter_Search_Integration tests the full search flow
// from semantic query → v2 execution → semantic results.
//
// This is an integration test that requires a running Elasticsearch instance
// on localhost:9200. It is skipped with -short or when ES is unreachable.
func TestV2SearchAdapter_Search_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup: Create ES client (assumes ES running on localhost:9200)
	client, err := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		t.Skipf("Elasticsearch not available: %v", err)
	}

	// Create adapter
	logger := zerolog.New(zerolog.NewTestWriter(t))
	adapter := NewV2SearchAdapter(client, "hub3-test", logger)

	// Create context with organization
	org := domain.Organization{
		ID: "testorg",
	}
	ctx := domain.SetOrganizationInContext(context.Background(), org)

	seedTestIndex(ctx, t, client)

	t.Run("simple text search", func(t *testing.T) {
		opts := &semantic.QueryOptions{
			Query: &semantic.TextQuery{
				Value: "test",
			},
			Pagination: &semantic.Pagination{
				Page: 1,
				Size: 10,
			},
		}

		result, err := adapter.Search(ctx, opts, nil)
		if err != nil {
			t.Errorf("Search() error = %v", err)
			return
		}

		if result == nil {
			t.Error("Search() returned nil result")
			return
		}

		// Verify result structure
		if result.Results == nil {
			t.Error("result.Results is nil")
		}
		if result.ResultIDs == nil {
			t.Error("result.ResultIDs is nil")
		}
		if result.Facets == nil {
			t.Error("result.Facets is nil")
		}
	})

	t.Run("search with filters", func(t *testing.T) {
		opts := &semantic.QueryOptions{
			Filters: []semantic.Filter{
				&semantic.PropertyFilter{
					FieldName:    "type",
					OperatorType: semantic.OpEqual,
					Value:        "image",
				},
			},
			Pagination: &semantic.Pagination{
				Page: 1,
				Size: 10,
			},
		}

		result, err := adapter.Search(ctx, opts, nil)
		if err != nil {
			t.Errorf("Search() with filters error = %v", err)
			return
		}

		if result == nil {
			t.Error("Search() returned nil result")
		}
	})

	t.Run("search with facets", func(t *testing.T) {
		opts := &semantic.QueryOptions{
			Facets: []semantic.FacetRequest{
				{Field: "creator", Limit: 10},
				{Field: "year", Limit: 20},
			},
			Pagination: &semantic.Pagination{
				Page: 1,
				Size: 10,
			},
		}

		result, err := adapter.Search(ctx, opts, nil)
		if err != nil {
			t.Errorf("Search() with facets error = %v", err)
			return
		}

		if result == nil {
			t.Error("Search() returned nil result")
		}
	})

	t.Run("missing orgID in context", func(t *testing.T) {
		emptyCtx := context.Background()
		opts := &semantic.QueryOptions{
			Pagination: &semantic.Pagination{
				Page: 1,
				Size: 10,
			},
		}

		_, err := adapter.Search(emptyCtx, opts, nil)
		if err == nil {
			t.Error("Search() should error when orgID missing from context")
		}
	})
}

// TestV2SearchAdapter_GetByID_Integration tests document retrieval by ID.
func TestV2SearchAdapter_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup: Create ES client
	client, err := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		t.Skipf("Elasticsearch not available: %v", err)
	}

	logger := zerolog.New(zerolog.NewTestWriter(t))
	adapter := NewV2SearchAdapter(client, "hub3-test", logger)

	// Create context with organization
	org := domain.Organization{
		ID: "testorg",
	}
	ctx := domain.SetOrganizationInContext(context.Background(), org)

	t.Run("get existing document", func(t *testing.T) {
		// Index into the index the adapter reads: {orgID}v2
		seedTestIndex(ctx, t, client)

		// Now retrieve it
		item, err := adapter.GetByID(ctx, "test-doc-123", nil)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
			return
		}

		if item == nil {
			t.Error("GetByID() returned nil item")
			return
		}

		// Verify semantic structure
		if _, ok := item["@id"]; !ok {
			t.Error("item missing @id field")
		}
	})

	t.Run("get non-existent document", func(t *testing.T) {
		_, err := adapter.GetByID(ctx, "non-existent-id", nil)
		if err == nil {
			t.Error("GetByID() should error for non-existent document")
		}
	})
}

// TestV2SearchAdapter_Aggregate_Integration tests facet-only queries.
func TestV2SearchAdapter_Aggregate_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup: Create ES client
	client, err := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		t.Skipf("Elasticsearch not available: %v", err)
	}

	logger := zerolog.New(zerolog.NewTestWriter(t))
	adapter := NewV2SearchAdapter(client, "hub3-test", logger)

	// Create context with organization
	org := domain.Organization{
		ID: "testorg",
	}
	ctx := domain.SetOrganizationInContext(context.Background(), org)

	seedTestIndex(ctx, t, client)

	t.Run("aggregate with facets", func(t *testing.T) {
		facets := []semantic.FacetRequest{
			{Field: "creator", Limit: 10},
			{Field: "type", Limit: 5},
		}

		results, err := adapter.Aggregate(ctx, facets, nil, nil)
		if err != nil {
			t.Errorf("Aggregate() error = %v", err)
			return
		}

		if results == nil {
			t.Error("Aggregate() returned nil results")
		}
	})

	t.Run("aggregate with filters", func(t *testing.T) {
		facets := []semantic.FacetRequest{
			{Field: "creator", Limit: 10},
		}
		filters := []semantic.Filter{
			&semantic.PropertyFilter{
				FieldName:    "type",
				OperatorType: semantic.OpEqual,
				Value:        "image",
			},
		}

		results, err := adapter.Aggregate(ctx, facets, filters, nil)
		if err != nil {
			t.Errorf("Aggregate() with filters error = %v", err)
			return
		}

		if results == nil {
			t.Error("Aggregate() returned nil results")
		}
	})

	t.Run("missing orgID in context", func(t *testing.T) {
		emptyCtx := context.Background()
		facets := []semantic.FacetRequest{
			{Field: "creator", Limit: 10},
		}

		_, err := adapter.Aggregate(emptyCtx, facets, nil, nil)
		if err == nil {
			t.Error("Aggregate() should error when orgID missing from context")
		}
	})
}

// TestV2SearchAdapter_SearchContext_Integration tests search context management.
func TestV2SearchAdapter_SearchContext_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup: Create ES client
	client, err := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		t.Skipf("Elasticsearch not available: %v", err)
	}

	logger := zerolog.New(zerolog.NewTestWriter(t))
	adapter := NewV2SearchAdapter(client, "hub3-test", logger)

	ctx := context.Background()

	t.Run("save and retrieve search context", func(t *testing.T) {
		searchCtx := &semantic.SearchContext{
			Token:     "test-token-123",
			ResultIDs: []string{"id1", "id2", "id3"},
		}

		// Save context
		err := adapter.SaveSearchContext(ctx, searchCtx)
		if err != nil {
			t.Errorf("SaveSearchContext() error = %v", err)
			return
		}

		// Retrieve context
		retrieved, err := adapter.GetSearchContext(ctx, "test-token-123")
		if err != nil {
			t.Errorf("GetSearchContext() error = %v", err)
			return
		}

		if retrieved.Token != searchCtx.Token {
			t.Errorf("Token = %s, want %s", retrieved.Token, searchCtx.Token)
		}

		if len(retrieved.ResultIDs) != len(searchCtx.ResultIDs) {
			t.Errorf("len(ResultIDs) = %d, want %d", len(retrieved.ResultIDs), len(searchCtx.ResultIDs))
		}
	})

	t.Run("get non-existent context", func(t *testing.T) {
		_, err := adapter.GetSearchContext(ctx, "non-existent-token")
		if err != semantic.ErrContextNotFound {
			t.Errorf("GetSearchContext() error = %v, want ErrContextNotFound", err)
		}
	})

	t.Run("delete search context", func(t *testing.T) {
		searchCtx := &semantic.SearchContext{
			Token:     "delete-test-token",
			ResultIDs: []string{"id1"},
		}

		// Save and verify
		adapter.SaveSearchContext(ctx, searchCtx)
		_, err := adapter.GetSearchContext(ctx, "delete-test-token")
		if err != nil {
			t.Errorf("Context should exist before delete")
		}

		// Delete
		err = adapter.DeleteSearchContext(ctx, "delete-test-token")
		if err != nil {
			t.Errorf("DeleteSearchContext() error = %v", err)
		}

		// Verify deleted
		_, err = adapter.GetSearchContext(ctx, "delete-test-token")
		if err != semantic.ErrContextNotFound {
			t.Error("Context should not exist after delete")
		}
	})
}

// TestV2SearchAdapter_Health_Integration tests health check.
func TestV2SearchAdapter_Health_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup: Create ES client
	client, err := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		t.Skipf("Elasticsearch not available: %v", err)
	}

	logger := zerolog.New(zerolog.NewTestWriter(t))
	adapter := NewV2SearchAdapter(client, "hub3-test", logger)

	ctx := context.Background()

	t.Run("health check", func(t *testing.T) {
		err := adapter.Health(ctx)
		if err != nil {
			t.Errorf("Health() error = %v", err)
		}
	})
}
