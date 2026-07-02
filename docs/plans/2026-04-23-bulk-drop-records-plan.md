# Bulk API `drop_records` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `drop_records` bulk action that deletes records by `hubId`
from both the triple store and every configured ES index, chunked in
10k batches, with atomic prefix validation.

**Architecture:** One new parser case that validates + delegates to one
new `DataSet` method. The method owns chunking and fans out to two
private helpers: one SPARQL (`DROP SILENT GRAPH`), one ES
(`DeleteByQuery` on `meta.hubID` terms). Synchronous, fail-fast, outside
the batch/revision machinery. Design: `docs/plans/2026-04-23-bulk-drop-records-design.md`.

**Tech Stack:** Go 1.21+, `olivere/elastic/v7`, `gorequest` for SPARQL
updates, `matryer/is` for tests.

---

## File structure

| File | Role |
|---|---|
| `ikuzo/service/x/bulk/request.go` | Extend `Request` struct with `HubIDs []string`. |
| `ikuzo/service/x/bulk/parser.go` | New switch case `"drop_records"` + helper `dropRecords`. |
| `ikuzo/service/x/bulk/parser_test.go` | Unit tests for the new parser path. |
| `hub3/models/dataset.go` | New public `DropRecordsByHubIDs` + two private helpers `dropGraphsByHubIDs` and `deleteIndexRecordsByHubIDs`. |
| `hub3/models/dataset_drop_records_test.go` | New integration-style test file (keeps the existing nearly-empty `dataset_test.go` out of the way). |
| `docs/plans/2026-04-23-bulk-drop-records-design.md` | Reference — do not modify. |

---

## Task 1: Request struct gains `HubIDs` field

**Files:**
- Modify: `ikuzo/service/x/bulk/request.go:32-50`

- [ ] **Step 1: Write the failing test**

Add to `ikuzo/service/x/bulk/request_test.go`:

```go
func TestRequestDecodesHubIDsField(t *testing.T) {
    is := is.New(t)
    body := `{"action":"drop_records","orgID":"o","dataset":"d","hubIds":["o_d_a","o_d_b"]}`
    var req Request
    err := json.Unmarshal([]byte(body), &req)
    is.NoErr(err)
    is.Equal(req.Action, "drop_records")
    is.Equal(len(req.HubIDs), 2)
    is.Equal(req.HubIDs[0], "o_d_a")
    is.Equal(req.HubIDs[1], "o_d_b")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/service/x/bulk/ -run TestRequestDecodesHubIDsField -v`
Expected: fails because `Request` has no `HubIDs` field (`json: unknown field` in strict mode, or zero-valued `HubIDs` in loose mode — both fail assertions).

- [ ] **Step 3: Add the field**

Edit `ikuzo/service/x/bulk/request.go` Request struct, appending the new field:

```go
type Request struct {
    HubID         string   `json:"hubId"`
    OrgID         string   `json:"orgID"`
    DatasetID     string   `json:"dataset"`
    LocalID       string   `json:"localId"`
    NamedGraphURI string   `json:"graphUri"`
    RecordType    string   `json:"type"`
    Action        string   `json:"action"`
    ContentHash   string   `json:"contentHash"`
    Graph         string   `json:"graph"`
    GraphMimeType string   `json:"graphMimeType"`
    SubjectType   string   `json:"subjectType"`
    Revision      int      `json:"revision"`
    Tags          string   `json:"tags,omitempty"`
    RecDefID      string   `json:"recDefId,omitempty"`
    HubIDs        []string `json:"hubIds,omitempty"`
    aboutTypeURI  []string
    resolvedTags  []string
    indexTypes    []string
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/service/x/bulk/ -run TestRequestDecodesHubIDsField -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add ikuzo/service/x/bulk/request.go ikuzo/service/x/bulk/request_test.go
git commit -m "feat(bulk): add HubIDs field to Request for drop_records"
```

---

## Task 2: Parser validates hubIds

**Files:**
- Modify: `ikuzo/service/x/bulk/parser.go` (add `dropRecords` helper)
- Test: `ikuzo/service/x/bulk/parser_test.go`

This task wires validation without the model call. That keeps this
commit small; Task 4 plugs the model in.

- [ ] **Step 1: Write the failing test**

Append to `ikuzo/service/x/bulk/parser_test.go`:

```go
func TestDropRecordsValidatesHubIDPrefix(t *testing.T) {
    is := is.New(t)
    p := &Parser{
        ds: &models.DataSet{OrgID: "org1", Spec: "ds1"},
    }
    req := &Request{
        Action:    "drop_records",
        OrgID:     "org1",
        DatasetID: "ds1",
        HubIDs:    []string{"org1_ds1_a", "org2_ds1_b"}, // second id cross-org
    }
    err := p.dropRecords(context.Background(), req)
    is.True(err != nil)
    is.True(strings.Contains(err.Error(), "does not belong to dataset"))
}

func TestDropRecordsEmptyListIsNoOp(t *testing.T) {
    is := is.New(t)
    p := &Parser{
        ds: &models.DataSet{OrgID: "org1", Spec: "ds1"},
    }
    req := &Request{
        Action:    "drop_records",
        OrgID:     "org1",
        DatasetID: "ds1",
        HubIDs:    nil,
    }
    err := p.dropRecords(context.Background(), req)
    is.NoErr(err)
}
```

Note: imports needed in the test file — `context`, `strings`, and the
`models` package. Top of `parser_test.go` already imports `bytes`,
`encoding/json`, `strings`, `testing`, `rdf2go`, `matryer/is`. Add
`context` and `github.com/delving/hub3/hub3/models` if not already
present.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./ikuzo/service/x/bulk/ -run TestDropRecords -v`
Expected: both fail — `dropRecords` method does not exist on `*Parser`.

- [ ] **Step 3: Add the helper**

Edit `ikuzo/service/x/bulk/parser.go`. Append after the existing
`dropOrphans` method (around line 244):

```go
// dropRecords validates the hubIds belong to this dataset, then asks
// the DataSet to delete them. Idempotent at the storage layer; see
// DataSet.DropRecordsByHubIDs.
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

At this point `models.DataSet` has no `DropRecordsByHubIDs` method yet,
so the package will not compile. The validation test still passes
because the validation error returns before the `DropRecordsByHubIDs`
call, but we need the method to exist for compilation.

- [ ] **Step 4: Stub the model method so the package compiles**

Edit `hub3/models/dataset.go`. Insert after `DropAll` (around line
862):

```go
// DropRecordsByHubIDs is a placeholder so the bulk parser can compile;
// real implementation lands in Task 4.
func (ds DataSet) DropRecordsByHubIDs(
    ctx context.Context,
    hubIDs []string,
) (int, error) {
    return 0, fmt.Errorf("DropRecordsByHubIDs not implemented")
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./ikuzo/service/x/bulk/ -run TestDropRecords -v`
Expected: PASS on both. `TestDropRecordsEmptyListIsNoOp` exits early,
`TestDropRecordsValidatesHubIDPrefix` returns the validation error
before reaching the not-implemented stub.

- [ ] **Step 6: Commit**

```bash
git add ikuzo/service/x/bulk/parser.go ikuzo/service/x/bulk/parser_test.go hub3/models/dataset.go
git commit -m "feat(bulk): parser validates drop_records hubIds against dataset"
```

---

## Task 3: Parser switch case

**Files:**
- Modify: `ikuzo/service/x/bulk/parser.go:315-375`
- Test: `ikuzo/service/x/bulk/parser_test.go`

Wire the switch case so a full `process` call routes `drop_records` to
our validated helper. The test stubs the model method to capture the
hubIds it received.

- [ ] **Step 1: Write the failing test**

Add to `ikuzo/service/x/bulk/parser_test.go`:

```go
func TestProcessRoutesDropRecords(t *testing.T) {
    is := is.New(t)
    // Capture via a package-level hook injected by the test. The real
    // Parser calls p.ds.DropRecordsByHubIDs directly; for this test we
    // rely on the not-yet-implemented stub returning its known error
    // string. A clean "test passed" means: validation passed, the
    // switch dispatched to dropRecords, and DropRecordsByHubIDs
    // returned the stub error.
    p := &Parser{
        ds: &models.DataSet{OrgID: "org1", Spec: "ds1", Revision: 1},
    }
    req := &Request{
        Action:    "drop_records",
        OrgID:     "org1",
        DatasetID: "ds1",
        HubIDs:    []string{"org1_ds1_a"},
    }
    err := p.process(context.Background(), req)
    is.True(err != nil)
    is.True(strings.Contains(err.Error(), "not implemented"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/service/x/bulk/ -run TestProcessRoutesDropRecords -v`
Expected: FAIL with `unknown bulk action: drop_records` (hitting the
default case of the switch).

- [ ] **Step 3: Add the switch case**

Edit `ikuzo/service/x/bulk/parser.go` inside the `switch req.Action`
block in `process` (around line 315). Add a new case after
`"drop_dataset"`:

```go
case "drop_records":
    if err := p.dropRecords(ctx, req); err != nil {
        subLogger.Error().
            Err(err).
            Str("datasetID", req.DatasetID).
            Msg("Unable to drop records")
        return err
    }
    subLogger.Info().
        Str("datasetID", req.DatasetID).
        Int("count", len(req.HubIDs)).
        Msg("dropped records")
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/service/x/bulk/ -run TestProcessRoutesDropRecords -v`
Expected: PASS (the stub's "not implemented" error propagates out, the
test asserts on exactly that).

- [ ] **Step 5: Run the full bulk package tests to be sure**

Run: `go test ./ikuzo/service/x/bulk/ -v`
Expected: all existing tests pass unchanged.

- [ ] **Step 6: Commit**

```bash
git add ikuzo/service/x/bulk/parser.go ikuzo/service/x/bulk/parser_test.go
git commit -m "feat(bulk): wire drop_records case into parser switch"
```

---

## Task 4: DataSet.DropRecordsByHubIDs real implementation

**Files:**
- Modify: `hub3/models/dataset.go` (replace stub)
- Test: `hub3/models/dataset_drop_records_test.go` (new)

Replace the Task-2 stub with the real method. Chunking at 10k is
enforced here so the two storage helpers always see bounded input.

Build tag note: `hub3/models` existing tests (e.g. `store_test.go`)
run without external services because they do not cover the storage
layer. These new tests must be able to run **without** ES or triple
store — we exercise the chunking + empty-list branches with the
storage config flags flipped off.

- [ ] **Step 1: Write the failing test**

Create `hub3/models/dataset_drop_records_test.go`:

```go
// Copyright © 2026 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package models

import (
    "context"
    "testing"

    c "github.com/delving/hub3/config"
    "github.com/matryer/is"
)

func withStoragesDisabled(t *testing.T) func() {
    t.Helper()
    c.InitConfig()
    prevRDF := c.Config.RDF.RDFStoreEnabled
    prevES := c.Config.ElasticSearch.Enabled
    c.Config.RDF.RDFStoreEnabled = false
    c.Config.ElasticSearch.Enabled = false
    return func() {
        c.Config.RDF.RDFStoreEnabled = prevRDF
        c.Config.ElasticSearch.Enabled = prevES
    }
}

func TestDropRecordsByHubIDsEmpty(t *testing.T) {
    is := is.New(t)
    restore := withStoragesDisabled(t)
    defer restore()

    ds := DataSet{OrgID: "org1", Spec: "ds1"}
    count, err := ds.DropRecordsByHubIDs(context.Background(), nil)
    is.NoErr(err)
    is.Equal(count, 0)

    count, err = ds.DropRecordsByHubIDs(context.Background(), []string{})
    is.NoErr(err)
    is.Equal(count, 0)
}

func TestDropRecordsByHubIDsNoOpWhenBothStoragesOff(t *testing.T) {
    is := is.New(t)
    restore := withStoragesDisabled(t)
    defer restore()

    ds := DataSet{OrgID: "org1", Spec: "ds1"}
    count, err := ds.DropRecordsByHubIDs(
        context.Background(),
        []string{"org1_ds1_a", "org1_ds1_b"},
    )
    is.NoErr(err)
    is.Equal(count, 0) // nothing deleted because nothing enabled
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./hub3/models/ -run TestDropRecordsByHubIDs -v`
Expected: both fail — the stub returns `"DropRecordsByHubIDs not implemented"`.

- [ ] **Step 3: Replace the stub with the real method**

Edit `hub3/models/dataset.go`. Replace the stub
`func (ds DataSet) DropRecordsByHubIDs(...)` added in Task 2 with:

```go
// DropRecordsByHubIDs deletes specific records from this dataset by
// hubID. Removes the matching graph from the triple store (if enabled)
// and matching documents from every configured ES index (if enabled).
// Idempotent at both layers: missing hubIDs are silently accepted.
// Internally chunked at 10_000 so a single oversized call from a
// client never produces a single oversized ES or SPARQL request.
// Returns the cumulative number of ES documents deleted; first error
// aborts remaining chunks (fail-fast).
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

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./hub3/models/ -run TestDropRecordsByHubIDs -v`
Expected: both PASS.

- [ ] **Step 5: Also run the bulk parser tests**

Run: `go test ./ikuzo/service/x/bulk/ -run TestDropRecords -v`
Expected: same three tests still pass (`process` test now gets an
empty-chunks no-op result `(0, nil)` instead of the stub error).

The `TestProcessRoutesDropRecords` test will need updating — it
expected `"not implemented"`. Change the expectation:

```go
func TestProcessRoutesDropRecords(t *testing.T) {
    is := is.New(t)
    _ = c.InitConfig
    prevRDF := c.Config.RDF.RDFStoreEnabled
    prevES := c.Config.ElasticSearch.Enabled
    c.Config.RDF.RDFStoreEnabled = false
    c.Config.ElasticSearch.Enabled = false
    defer func() {
        c.Config.RDF.RDFStoreEnabled = prevRDF
        c.Config.ElasticSearch.Enabled = prevES
    }()

    p := &Parser{
        ds: &models.DataSet{OrgID: "org1", Spec: "ds1", Revision: 1},
    }
    req := &Request{
        Action:    "drop_records",
        OrgID:     "org1",
        DatasetID: "ds1",
        HubIDs:    []string{"org1_ds1_a"},
    }
    err := p.process(context.Background(), req)
    is.NoErr(err) // validation + real method with storages off = no error
}
```

Add the `c` import (`c "github.com/delving/hub3/config"`) in
`parser_test.go` if not already present.

- [ ] **Step 6: Re-run the full bulk and models package tests**

Run:
```
go test ./ikuzo/service/x/bulk/ -v
go test ./hub3/models/ -v
```
Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add hub3/models/dataset.go hub3/models/dataset_drop_records_test.go ikuzo/service/x/bulk/parser_test.go
git commit -m "feat(models): DropRecordsByHubIDs with 10k chunking"
```

---

## Task 5: Triple-store helper

**Files:**
- Modify: `hub3/models/dataset.go` (add `dropGraphsByHubIDs`)
- Test: `hub3/models/dataset_drop_records_test.go`

- [ ] **Step 1: Write the failing test**

Append to `hub3/models/dataset_drop_records_test.go`:

```go
func TestDropGraphsByHubIDsBuildsDropSilentQuery(t *testing.T) {
    is := is.New(t)
    // Override the SPARQL sender with a fake that records its input.
    var captured string
    prev := sparqlUpdateSender
    sparqlUpdateSender = func(orgID, update string) []error {
        captured = update
        return nil
    }
    defer func() { sparqlUpdateSender = prev }()

    ds := DataSet{OrgID: "org1", Spec: "ds1"}
    err := ds.dropGraphsByHubIDs([]string{"org1_ds1_a", "org1_ds1_b"})
    is.NoErr(err)

    is.True(strings.Contains(captured, "DROP SILENT GRAPH <urn:org1_ds1_a/graph>"))
    is.True(strings.Contains(captured, "DROP SILENT GRAPH <urn:org1_ds1_b/graph>"))
}

func TestDropGraphsByHubIDsPropagatesError(t *testing.T) {
    is := is.New(t)
    prev := sparqlUpdateSender
    sparqlUpdateSender = func(orgID, update string) []error {
        return []error{fmt.Errorf("boom")}
    }
    defer func() { sparqlUpdateSender = prev }()

    ds := DataSet{OrgID: "org1", Spec: "ds1"}
    err := ds.dropGraphsByHubIDs([]string{"org1_ds1_a"})
    is.True(err != nil)
    is.True(strings.Contains(err.Error(), "boom"))
}
```

Add the imports `fmt`, `strings` to the test file if not already there.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hub3/models/ -run TestDropGraphs -v`
Expected: FAIL — neither `dropGraphsByHubIDs` nor `sparqlUpdateSender`
exists.

- [ ] **Step 3: Introduce the sender seam + helper**

Edit `hub3/models/dataset.go`. Near the top of the file (after the
package imports, before the type declarations), add:

```go
// sparqlUpdateSender is an indirection used by helpers that issue
// SPARQL Update requests so tests can intercept the payload without
// needing a real triple store. Production callers should not replace
// this.
var sparqlUpdateSender = fragments.UpdateViaSparql
```

Add the `fragments` import if not present:
`"github.com/delving/hub3/hub3/fragments"`.

Then add the helper method (next to `DropRecordsByHubIDs`):

```go
// dropGraphsByHubIDs issues a single SPARQL Update request containing
// one DROP SILENT GRAPH statement per hubID. SILENT means the operation
// succeeds even when a graph is already absent, giving the caller
// idempotency for free.
func (ds DataSet) dropGraphsByHubIDs(hubIDs []string) error {
    var sb strings.Builder
    for _, hid := range hubIDs {
        sb.WriteString("DROP SILENT GRAPH <urn:")
        sb.WriteString(hid)
        sb.WriteString("/graph>;\n")
    }
    errs := sparqlUpdateSender(ds.OrgID, sb.String())
    if len(errs) > 0 {
        return errs[0]
    }
    return nil
}
```

Add `"strings"` to the dataset.go imports if not already present.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hub3/models/ -run TestDropGraphs -v`
Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add hub3/models/dataset.go hub3/models/dataset_drop_records_test.go
git commit -m "feat(models): dropGraphsByHubIDs emits DROP SILENT GRAPH"
```

---

## Task 6: ES helper

**Files:**
- Modify: `hub3/models/dataset.go` (add `deleteIndexRecordsByHubIDs`)
- Test: `hub3/models/dataset_drop_records_test.go`

The query-building is what we can unit-test cheaply; the actual ES
round trip is integration territory and is covered on dcn-acpt in
Task 5 of the Narthex plan. Here we assert the query shape via a
seam.

- [ ] **Step 1: Write the failing test**

Append to `hub3/models/dataset_drop_records_test.go`:

```go
func TestDeleteIndexRecordsByHubIDsBuildsTermsQuery(t *testing.T) {
    is := is.New(t)

    var capturedIndices []string
    var capturedHubIDs [][]interface{}

    prev := esDeleteByQuerySender
    esDeleteByQuerySender = func(
        ctx context.Context,
        index string,
        q elastic.Query,
    ) (int, error) {
        capturedIndices = append(capturedIndices, index)
        // Extract the terms clause so we can assert hubIDs shipped through
        src, err := q.Source()
        if err != nil {
            return 0, err
        }
        b, _ := json.Marshal(src)
        var shaped struct {
            Bool struct {
                Must []map[string]interface{} `json:"must"`
            } `json:"bool"`
        }
        _ = json.Unmarshal(b, &shaped)
        for _, clause := range shaped.Bool.Must {
            if terms, ok := clause["terms"].(map[string]interface{}); ok {
                if ids, ok := terms["meta.hubID"].([]interface{}); ok {
                    capturedHubIDs = append(capturedHubIDs, ids)
                }
            }
        }
        return 2, nil
    }
    defer func() { esDeleteByQuerySender = prev }()

    // Ensure config.InitConfig ran once
    c.InitConfig()
    prevTypes := c.Config.ElasticSearch.IndexTypes
    c.Config.ElasticSearch.IndexTypes = []string{"v2"}
    defer func() { c.Config.ElasticSearch.IndexTypes = prevTypes }()

    ds := DataSet{OrgID: "org1", Spec: "ds1"}
    count, err := ds.deleteIndexRecordsByHubIDs(
        context.Background(),
        []string{"org1_ds1_a", "org1_ds1_b"},
    )
    is.NoErr(err)
    is.Equal(count, 2)
    is.Equal(len(capturedIndices), 1) // v2 only

    is.Equal(len(capturedHubIDs), 1)
    is.Equal(len(capturedHubIDs[0]), 2)
    is.Equal(capturedHubIDs[0][0].(string), "org1_ds1_a")
    is.Equal(capturedHubIDs[0][1].(string), "org1_ds1_b")
}
```

Add imports `encoding/json` and `olivere/elastic/v7` (use the same
import path the rest of the file uses — likely
`"github.com/olivere/elastic/v7"` aliased as `elastic`). Peek at the
top of `hub3/models/dataset.go` for the canonical alias.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hub3/models/ -run TestDeleteIndexRecordsByHubIDs -v`
Expected: FAIL — `deleteIndexRecordsByHubIDs` and
`esDeleteByQuerySender` do not exist.

- [ ] **Step 3: Introduce the sender seam + helper**

Edit `hub3/models/dataset.go`. Near the existing `sparqlUpdateSender`
seam, add:

```go
// esDeleteByQuerySender is an indirection used by helpers that issue
// ES DeleteByQuery requests so tests can assert on the final query
// without needing a real ES cluster.
var esDeleteByQuerySender = func(
    ctx context.Context,
    indexName string,
    q elastic.Query,
) (int, error) {
    res, err := index.ESClient().DeleteByQuery().
        Index(indexName).
        Query(q).
        Conflicts("proceed").
        Do(ctx)
    if err != nil {
        return 0, err
    }
    if res == nil {
        return 0, fmt.Errorf(unexpectedResponseMsg, res)
    }
    return int(res.Deleted), nil
}
```

Add the helper method:

```go
// deleteIndexRecordsByHubIDs issues one DeleteByQuery per configured
// index type, constrained to this dataset and the given hubIDs. Defence
// in depth: even though the parser already validates the hubId prefix,
// the spec+orgID filters here prevent a misrouted request from deleting
// outside its dataset.
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

    indices := ds.resolveIndicesForHubIDDelete()
    total := 0
    for _, idx := range indices {
        deleted, err := esDeleteByQuerySender(ctx, idx, q)
        if err != nil {
            return total, err
        }
        total += deleted
    }
    return total, nil
}

// resolveIndicesForHubIDDelete mirrors the index-type resolution logic
// used by deleteAllIndexRecords/deleteIndexOrphans, scoped to the
// indices that actually hold per-hubID documents.
func (ds DataSet) resolveIndicesForHubIDDelete() []string {
    var indices []string
    for _, indexType := range c.Config.ElasticSearch.IndexTypes {
        switch indexType {
        case v1Type:
            indices = append(indices, c.Config.ElasticSearch.GetV1IndexName(ds.OrgID))
        case v2Type:
            indices = append(indices, c.Config.ElasticSearch.GetIndexName(ds.OrgID))
        case fragmentType:
            indices = append(indices, c.Config.ElasticSearch.FragmentIndexName(ds.OrgID))
        }
    }
    return indices
}
```

Note: we intentionally do not target the `suggest` index because those
documents are not keyed by `meta.hubID` in the same shape — they are
a derived index that the normal indexing pipeline rebuilds. Skipping
keeps the delete simple and avoids partial failures on indices that
might not carry the field.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hub3/models/ -run TestDeleteIndexRecordsByHubIDs -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add hub3/models/dataset.go hub3/models/dataset_drop_records_test.go
git commit -m "feat(models): deleteIndexRecordsByHubIDs via terms query"
```

---

## Task 7: Chunking behaviour test

**Files:**
- Test: `hub3/models/dataset_drop_records_test.go`

Verify the 10k chunking in `DropRecordsByHubIDs` produces the expected
number of downstream calls.

- [ ] **Step 1: Write the failing test**

Append to `hub3/models/dataset_drop_records_test.go`:

```go
func TestDropRecordsByHubIDsChunksAt10k(t *testing.T) {
    is := is.New(t)
    c.InitConfig()
    prevRDF := c.Config.RDF.RDFStoreEnabled
    prevES := c.Config.ElasticSearch.Enabled
    prevTypes := c.Config.ElasticSearch.IndexTypes
    c.Config.RDF.RDFStoreEnabled = true
    c.Config.ElasticSearch.Enabled = true
    c.Config.ElasticSearch.IndexTypes = []string{"v2"}
    defer func() {
        c.Config.RDF.RDFStoreEnabled = prevRDF
        c.Config.ElasticSearch.Enabled = prevES
        c.Config.ElasticSearch.IndexTypes = prevTypes
    }()

    sparqlCalls := 0
    prevSparql := sparqlUpdateSender
    sparqlUpdateSender = func(orgID, update string) []error {
        sparqlCalls++
        return nil
    }
    defer func() { sparqlUpdateSender = prevSparql }()

    esCalls := 0
    prevES2 := esDeleteByQuerySender
    esDeleteByQuerySender = func(ctx context.Context, index string, q elastic.Query) (int, error) {
        esCalls++
        return 0, nil
    }
    defer func() { esDeleteByQuerySender = prevES2 }()

    // 25k hubIds → 3 chunks (10k, 10k, 5k)
    hubIDs := make([]string, 25_000)
    for i := range hubIDs {
        hubIDs[i] = fmt.Sprintf("org1_ds1_%d", i)
    }

    ds := DataSet{OrgID: "org1", Spec: "ds1"}
    _, err := ds.DropRecordsByHubIDs(context.Background(), hubIDs)
    is.NoErr(err)

    is.Equal(sparqlCalls, 3)
    is.Equal(esCalls, 3) // 1 index type × 3 chunks
}
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./hub3/models/ -run TestDropRecordsByHubIDsChunksAt10k -v`
Expected: PASS — the chunking already works; this test just locks the
behaviour in.

- [ ] **Step 3: Commit**

```bash
git add hub3/models/dataset_drop_records_test.go
git commit -m "test(models): lock in 10k chunking for DropRecordsByHubIDs"
```

---

## Task 8: Full regression pass + review

**Files:** none modified.

- [ ] **Step 1: Run the full hub3 + bulk test suites**

```
go test ./hub3/... ./ikuzo/service/x/bulk/ -count=1
```

Expected: all pass. If anything unrelated fails, investigate; flaky
tests are worth a separate fix rather than accepting risk.

- [ ] **Step 2: Static checks**

```
go vet ./...
gofmt -l hub3/models/dataset.go hub3/models/dataset_drop_records_test.go ikuzo/service/x/bulk/parser.go ikuzo/service/x/bulk/parser_test.go ikuzo/service/x/bulk/request.go ikuzo/service/x/bulk/request_test.go
```

`go vet` should print nothing new. `gofmt -l` should print no files.
If either complains, fix and re-run before opening the PR.

- [ ] **Step 3: Request code review**

Reference skill: `superpowers:requesting-code-review`.

Ensure the review includes: matching the design doc, idempotency
semantics, prefix validation atomicity, chunking correctness at the 10k
boundary, no cross-package coupling regressions.

- [ ] **Step 4: Open the PR**

```
gh pr create --base dev-v0.5 --title "feat(bulk): drop_records action" --body "$(cat <<'EOF'
## Summary
- New bulk action `drop_records` takes a list of `hubIds` and deletes matching records from the triple store and every configured ES index.
- Internally chunked at 10k. Atomic prefix validation. Fail-fast on any storage error.

## Design
docs/plans/2026-04-23-bulk-drop-records-design.md

## Coordination
Pairs with Narthex feat/record-registry. Narthex will flip payload key from `ids` to `hubIds` and send `hubId = orgId_spec_localId` in a follow-up commit once this PR is merged to dev-v0.5 and deployed to dcn-acpt.

## Test plan
- [ ] unit tests (this PR): parser validation, model chunking, SPARQL/ES query shape
- [ ] manual: dcn-acpt deploy, call drop_records with a handful of real hubIds, verify ES + triple store state
- [ ] end-to-end: Narthex Task 5 verification matrix
EOF
)"
```

---

## Self-review notes

Spec coverage: every section of the design doc has a corresponding
task. Section 1 wire format → Task 1. Section 2 parser switch → Tasks
2, 3. Section 3 model → Tasks 4, 5, 6. Testing plan → Tasks 5, 6, 7.
Rollout coordination → Task 8 PR body.

Method signatures used consistently:
- `DropRecordsByHubIDs(ctx, hubIDs) (int, error)` — Tasks 2, 4, 7.
- `dropGraphsByHubIDs(hubIDs) error` — Task 5.
- `deleteIndexRecordsByHubIDs(ctx, hubIDs) (int, error)` — Task 6.
- `sparqlUpdateSender(orgID, update) []error` — Tasks 5, 7.
- `esDeleteByQuerySender(ctx, index, query) (int, error)` — Tasks 6, 7.

No placeholders. Every step has concrete code or a concrete command.
