# Bulk API: `drop_records` Action — Design

Status: implemented
Date: 2026-04-23
Owner: hub3 (coordinated with Narthex `feat/record-registry`)

## Context

Narthex has a new per-dataset record registry (`records.db`, see narthex
`feat/record-registry`) that tracks each record's content hash and tombstone
status across harvests. The registry already emits explicit per-record
delete intents to Hub3 via a new `drop_records` bulk action, but the
server side of that action does not yet exist. This spec defines that
server side.

The absence of a per-record delete action has a concrete business impact:
Narthex incremental harvests never depublish records that are deleted at
the source. Today the only cleanup path is a full-harvest
`increment_revision` + `clear_orphans` cycle, which re-sends every record
on every full harvest — wasteful for large datasets and inapplicable to
incremental-only flows. A list-based `drop_records` action lets Narthex
depublish exactly the records that need depublishing, on every save
regardless of harvest kind.

## Goals

- Accept a batched list of `hubIds` in a single bulk-action line and
  delete each corresponding record from both the triple store and every
  configured ES index.
- Idempotent: repeated delete calls for already-gone records succeed.
- Scales: batches larger than 10k internally chunk so neither ES nor the
  triple store sees an oversized request.
- Safe: reject cross-dataset / cross-org ids before any delete fires.
- Minimal surface: one parser case, one `DataSet` method, two helpers.

## Non-goals

- Per-id failure reporting. First error aborts the whole call with a 500.
  Narthex keeps the registry rows as pending-drop so the next save
  retries cleanly.
- Posthook emission per record.
- Touching `IndexingBatch`, `ExpectedRecords`, or `Revision` bookkeeping.
  `drop_records` is a side-channel, not a batch participant.
- EAD cache cleanup.
- Async execution via `IndexMessage` / `ActionType_DROP_*`. Synchronous
  in-parser execution matches `disable_index` and `drop_dataset`.
- Hub3 kill switch. Narthex owns the emission toggle
  (`narthex.registry.emitDropRecords`).

## Wire format

`POST /api/index/bulk/` — one JSON object per line, as today. New
action shape:

```json
{
  "dataset": "myspec",
  "orgId": "myorg",
  "action": "drop_records",
  "hubIds": ["myorg_myspec_id1", "myorg_myspec_id2"]
}
```

Narthex computes `hubId = "${orgId}_${spec}_${localId}"` client-side and
sends the list. Hub3 uses `hubId` both as the ES `_id` and as the source
for the triple-store graph URI.

## Request struct

Extend `ikuzo/service/x/bulk/request.go`:

```go
type Request struct {
    // ... existing fields ...
    HubIDs []string `json:"hubIds,omitempty"`
}
```

`HubIDs` is populated only for `drop_records`; other action paths ignore
it.

## Parser

`ikuzo/service/x/bulk/parser.go` gets one new case in the `switch
req.Action` block:

```go
case "drop_records":
    if err := p.dropRecords(ctx, req); err != nil {
        subLogger.Error().Err(err).Str("datasetID", req.DatasetID).
            Msg("Unable to drop records")
        return err
    }
    subLogger.Info().
        Str("datasetID", req.DatasetID).
        Int("count", len(req.HubIDs)).
        Msg("dropped records")
```

and one new helper:

```go
func (p *Parser) dropRecords(ctx context.Context, req *Request) error {
    if len(req.HubIDs) == 0 {
        return nil
    }
    prefix := req.OrgID + "_" + req.DatasetID + "_"
    for _, hid := range req.HubIDs {
        if !strings.HasPrefix(hid, prefix) {
            return fmt.Errorf(
                "hubId %q does not belong to dataset %s/%s",
                hid, req.OrgID, req.DatasetID,
            )
        }
    }
    _, err := p.ds.DropRecordsByHubIDs(ctx, req.HubIDs)
    return err
}
```

Validation is atomic: any single mismatched id aborts the whole request
before any delete fires.

## Model

Add to `hub3/models/dataset.go`:

```go
// DropRecordsByHubIDs deletes specific records from this dataset by
// hubID. Removes the matching graph from the triple store (if enabled)
// and matching documents from every configured ES index. Idempotent:
// a missing hubID is a no-op at both layers. Internally chunked at 10k
// so a single large Narthex drop batch never produces a single
// oversized ES or SPARQL request.
func (ds DataSet) DropRecordsByHubIDs(
    ctx context.Context,
    hubIDs []string,
) (int, error) {
    if len(hubIDs) == 0 {
        return 0, nil
    }
    const chunkSize = 10000
    total := 0
    for i := 0; i < len(hubIDs); i += chunkSize {
        end := i + chunkSize
        if end > len(hubIDs) {
            end = len(hubIDs)
        }
        batch := hubIDs[i:end]

        if c.Config.RDF.RDFStoreEnabled {
            if err := ds.dropGraphsByHubIDs(batch); err != nil {
                return total, fmt.Errorf("triple store drop: %w", err)
            }
        }
        if c.Config.ElasticSearch.Enabled {
            deleted, err := ds.deleteIndexRecordsByHubIDs(ctx, batch)
            if err != nil {
                return total, fmt.Errorf("index drop: %w", err)
            }
            total += deleted
        }
    }
    return total, nil
}
```

### Triple-store helper

```go
func (ds DataSet) dropGraphsByHubIDs(hubIDs []string) error {
    var sb strings.Builder
    for _, hid := range hubIDs {
        sb.WriteString("DROP SILENT GRAPH <urn:")
        sb.WriteString(hid)
        sb.WriteString("/graph>;\n")
    }
    _, err := runSparqlUpdateQuery(sb.String())
    return err
}
```

`DROP SILENT` makes already-missing graphs a no-op.

### ES helper

```go
func (ds DataSet) deleteIndexRecordsByHubIDs(
    ctx context.Context,
    hubIDs []string,
) (int, error) {
    asIface := make([]interface{}, len(hubIDs))
    for i, h := range hubIDs {
        asIface[i] = h
    }
    q := elastic.NewBoolQuery().
        Must(elastic.NewTermQuery(c.Config.ElasticSearch.SpecKey, ds.Spec)).
        Must(elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, ds.OrgID)).
        Must(elastic.NewTermsQuery("meta.hubID", asIface...))

    indices := resolveIndicesFor(ds.OrgID)
    total := 0
    for _, idx := range indices {
        res, err := index.ESClient().DeleteByQuery().
            Index(idx).
            Query(q).
            Conflicts("proceed").
            Do(ctx)
        if err != nil {
            return total, err
        }
        if res != nil {
            total += int(res.Deleted)
        }
    }
    return total, nil
}
```

Spec+orgID constraint on the query is redundant with the parser prefix
check but defends in depth against misconfigured callers.

`resolveIndicesFor` mirrors the index-resolution logic already present
in `deleteAllIndexRecords` (v1, v2, fragment, suggest, as configured).

## Failure semantics

- Validation failure → error returned from `dropRecords`, 500 propagates
  to Narthex. Narthex's registry row stays pending-drop, retried next
  save. No partial deletes because validation runs before all deletes.
- Triple-store error → error returned. Some chunks may have already
  committed; the failed chunk is retried as-is on Narthex's next save
  (SPARQL `DROP SILENT` is idempotent).
- ES error → same: total counter reflects what succeeded; Narthex does
  not confirm the drop, retries next save.

Fail-fast. No partial-success 2xx path.

## Testing

Unit tests in `ikuzo/service/x/bulk/parser_test.go`:

1. `drop_records` with empty `hubIds` — no error, no delete side effect.
2. `drop_records` with one mismatched-prefix id — returns validation
   error, asserts no delete called.
3. `drop_records` with valid ids — asserts `DropRecordsByHubIDs` called
   once with the exact id list. Use a fake `DataSet` wrapper or inject a
   test double.
4. `drop_records` with 25_000 valid ids — asserts internal chunking
   yields 3 ES DeleteByQuery calls per configured index and 3 SPARQL
   requests (use counting fakes).

Model-level tests in `hub3/models/dataset_test.go`:

5. Seed ES + triple store with 5 records. Call
   `DropRecordsByHubIDs` with 3 of the 5 ids. Assert those 3 are gone
   in both stores and the other 2 remain. Use the same test scaffolding
   other dataset tests use (build tag + config.InitConfig + fixture
   dataset).
6. Re-call with the same 3 ids — no error, no effect (idempotency).

## Rollout coordination

1. Merge this PR to `dev-v0.5`. Tag + deploy to dcn-acpt Hub3.
2. In Narthex `feat/record-registry`, change `DsInfo.dropRecordsByIds`
   so it:
   - Accepts `localIds` (unchanged signature).
   - Computes `hubIds = s"${orgId}_${spec}_$localId"` at emit time.
   - Sends `{"action":"drop_records","hubIds":[...]}` instead of
     `{"action":"drop_records","ids":[...]}`.
3. Add `narthex.registry.emitDropRecords` config (default false) so
   Narthex can ship with registry writes enabled but drop POSTs muted
   until Hub3 dcn-acpt is confirmed green.
4. Deploy Narthex to dcn-acpt with `emitDropRecords=true`.
5. Run Task 5 verification matrix (narthex plan).
6. Promote Hub3 + Narthex to production dcn.

## Open items (resolved during brainstorm)

- Payload shape: `hubIds` array.
- Execution: synchronous in-parser.
- Delete path: `DeleteByQuery` on `meta.hubID` terms + one SPARQL
  request per chunk.
- Validation: prefix check atomically before any delete.
- Partial failure: fail-fast, 500.
- Batch/revision accounting: outside batches, no revision bump.
- Chunking threshold: 10_000 hubIds.
