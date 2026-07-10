package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olivere/elastic/v7"

	"github.com/delving/hub3/ikuzo/service/x/diwfragments"
)

// diwFragmentMapping is the Elasticsearch mapping for the per-org DIW
// fragment index. The store is a keyed blob cache, not a search surface:
// html and headTags are marked "index": false so the (potentially large)
// rendered markup is stored and retrievable by id but never analyzed or
// scored, and meta is "enabled": false so its free-form debug fields never
// need a mapping update. Every identifying field is a keyword because the
// only lookup this index ever serves is an exact get-by-DocID.
const diwFragmentMapping = `{
  "mappings": {
    "properties": {
      "orgID":      {"type": "keyword"},
      "collection": {"type": "keyword"},
      "kind":       {"type": "keyword"},
      "recordID":   {"type": "keyword"},
      "lang":       {"type": "keyword"},
      "html":       {"type": "text", "index": false},
      "headTags":   {"type": "text", "index": false},
      "meta":       {"type": "object", "enabled": false}
    }
  }
}`

// DiwFragmentStore persists DIW fragments in one Elasticsearch index per
// organization, keyed by Fragment.DocID so repeated renders of the same
// item/listing upsert in place instead of accumulating duplicates.
//
// VERIFY note (resolved): the task brief that shaped this file assumed the
// typed go-elasticsearch client lived at Client.search. In this package
// Client.search is actually the olivere/elastic v7 client (see sitemap.go,
// oaipmh_store.go) used for query/Get calls, while Client.es is the typed
// github.com/elastic/go-elasticsearch/v8 client (see indices.go,
// Client.Bulk in client.go) used for index administration and bulk writes.
// DiwFragmentStore follows that existing split rather than the brief's
// single-field guess: ensureIndex and Put use client.es (index
// create/exists and bulk, matching indices.go's Indices.Create/Exists and
// client.go's Bulk idiom), and Get uses client.search (matching
// oaipmh_store.go's GetRecord and semantic_store.go's GetByID, including
// the elastic.IsNotFound(err) check for a missing document).
type DiwFragmentStore struct {
	client *Client
}

// NewDiwFragmentStore returns the Elasticsearch-backed diwfragments.Store.
// It requires no setup beyond a live *Client: the destination index is
// created lazily on first write via ensureIndex, so a fresh org needs no
// out-of-band provisioning step before fragments can be rendered into it.
func (c *Client) NewDiwFragmentStore() *DiwFragmentStore {
	return &DiwFragmentStore{client: c}
}

// var _ diwfragments.Store ensures DiwFragmentStore keeps satisfying the
// service-layer interface at compile time; a signature drift on either
// side fails the build here instead of surfacing at runtime.
var _ diwfragments.Store = (*DiwFragmentStore)(nil)

// diwFragmentIndexName scopes fragments per organization and pins a mapping
// version so a future breaking mapping change can roll to a new index name
// (…_v2) without touching or reindexing unrelated record indices.
func diwFragmentIndexName(orgID string) string {
	return fmt.Sprintf("%s_diwfragments_v1", orgID)
}

// ensureIndex creates the fragment index for orgID the first time it is
// needed. It tolerates two forms of races that are normal for a lazily
// created index under concurrent writers: an Exists check that errors out
// (fall through and try to create; Create is itself idempotent) and a
// Create that reports the index already exists (400
// resource_already_exists_exception), which is treated as success.
func (s *DiwFragmentStore) ensureIndex(ctx context.Context, orgID string) error {
	name := diwFragmentIndexName(orgID)

	existsRes, err := s.client.es.Indices.Exists(
		[]string{name},
		s.client.es.Indices.Exists.WithContext(ctx),
	)
	if err == nil && existsRes != nil {
		defer existsRes.Body.Close()
		if existsRes.StatusCode == 200 {
			return nil
		}
	}

	createRes, err := s.client.es.Indices.Create(
		name,
		s.client.es.Indices.Create.WithBody(strings.NewReader(diwFragmentMapping)),
		s.client.es.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("create index %s: %w", name, err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() && createRes.StatusCode != 400 {
		return fmt.Errorf("create index %s: %s", name, createRes.String())
	}

	return nil
}

// Put upserts fragments by DocID with refresh=true so a render worker's
// write is immediately visible to the serving GET routes; fragment batches
// are small (one item/listing render at a time) and infrequent, so paying
// for a synchronous refresh on every write is cheap relative to the
// alternative of a stale cache window right after (re)indexing.
//
// All fragments in a batch are assumed to belong to the same organization
// (the render worker groups its batches that way); ensureIndex only runs
// once, keyed off the first fragment's OrgID.
func (s *DiwFragmentStore) Put(ctx context.Context, fragments []diwfragments.Fragment) error {
	if len(fragments) == 0 {
		return nil
	}

	if err := s.ensureIndex(ctx, fragments[0].OrgID); err != nil {
		return err
	}

	var buf strings.Builder

	for i := range fragments {
		f := fragments[i]

		action, err := json.Marshal(map[string]any{
			"index": map[string]string{
				"_index": diwFragmentIndexName(f.OrgID),
				"_id":    f.DocID(),
			},
		})
		if err != nil {
			return fmt.Errorf("encode bulk action for %s: %w", f.DocID(), err)
		}

		doc, err := json.Marshal(f)
		if err != nil {
			return fmt.Errorf("encode fragment %s: %w", f.DocID(), err)
		}

		buf.Write(action)
		buf.WriteByte('\n')
		buf.Write(doc)
		buf.WriteByte('\n')
	}

	res, err := s.client.es.Bulk(
		strings.NewReader(buf.String()),
		s.client.es.Bulk.WithContext(ctx),
		s.client.es.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("bulk fragment write: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk fragment write: %s", res.String())
	}

	// A bulk request can succeed at the transport level (HTTP 200, so
	// res.IsError() is false) while individual items were rejected —
	// mapping conflicts, shard failures, and the like are reported only
	// per item, under a body-level "errors": true flag. Decode the body
	// and fail loudly on any item error, because the render worker's
	// "stored N" accounting — and therefore the serving cache's
	// completeness — must only be trusted when every fragment landed.
	var bulkResp struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			ID    string          `json:"_id"`
			Error json.RawMessage `json:"error"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&bulkResp); err != nil {
		return fmt.Errorf("decode bulk fragment response: %w", err)
	}

	if bulkResp.Errors {
		var failed []string

		for _, item := range bulkResp.Items {
			for _, op := range item {
				if len(op.Error) > 0 && string(op.Error) != "null" {
					failed = append(failed, op.ID)
				}
			}
		}

		// Name the first few failed DocIDs so the log line alone is
		// actionable without replaying the bulk request.
		const maxNamed = 3

		named := failed
		if len(named) > maxNamed {
			named = named[:maxNamed]
		}

		return fmt.Errorf(
			"bulk fragment write: %d of %d fragments failed (e.g. %s)",
			len(failed), len(fragments), strings.Join(named, ", "),
		)
	}

	return nil
}

// Get fetches one fragment by its deterministic DocID, computed from the
// same fields (orgID, collection, kind, recordID, lang) the caller passes
// in — this mirrors the write path's Fragment.DocID() so a fragment
// written by Put is always found by the equivalent Get. Per the
// diwfragments.Store contract, an absent fragment is not an error: a
// missing document (detected via elastic.IsNotFound) returns (nil, nil) so
// callers can treat "not yet rendered" as a normal, cheap case.
func (s *DiwFragmentStore) Get(
	ctx context.Context,
	orgID, collection string,
	kind diwfragments.Kind,
	recordID, lang string,
) (*diwfragments.Fragment, error) {
	f := diwfragments.Fragment{
		OrgID:      orgID,
		Collection: collection,
		Kind:       kind,
		RecordID:   recordID,
		Lang:       lang,
	}

	res, err := s.client.search.Get().
		Index(diwFragmentIndexName(orgID)).
		Id(f.DocID()).
		Do(ctx)
	if err != nil {
		if elastic.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("get fragment %s: %w", f.DocID(), err)
	}

	var found diwfragments.Fragment
	if err := json.Unmarshal(res.Source, &found); err != nil {
		return nil, fmt.Errorf("decode fragment %s: %w", f.DocID(), err)
	}

	return &found, nil
}
