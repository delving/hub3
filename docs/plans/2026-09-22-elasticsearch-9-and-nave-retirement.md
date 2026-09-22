# Elasticsearch 9 Migration and Nave Retirement

**Date:** 2026-09-22
**Status:** Findings and proposed sequence — not yet ratified
**Related:** [ADR 0001](../adr/0001-semantic-v1-wraps-v2.md), [V2 Feature Freeze](../v2-feature-freeze.md), [V2-to-Semantic Migration Design](2026-02-20-v2-semantic-migration-design.md)

## Goal

Retire Nave (the Python/Django application) by reimplementing its v1 search API in
Go, so that the entire search stack runs on hub3 and can be moved to
Elasticsearch 9. Along the way, reduce `hub3/fragments` from its original
maximally-extensible design to what production actually uses, and use the
measurements below to make the remaining queries faster.

Three things make this the right moment:

- Elasticsearch 7.17 reached **end of support on 2026-01-15**. The cluster no
  longer receives security patches.
- `olivere/elastic`, which builds the query DSL across 34 files, is dead and
  **cannot be bridged** to ES 8+ (see [Client situation](#client-situation-olivere-is-a-dead-end)).
- Indexing is fast enough — proven on 2026-09-21, when 351 datasets were
  re-indexed end-to-end in a few hours — that an offline rebuild followed by a
  switch is a practical migration strategy rather than a risky one.

## Terminology

"v1" is badly overloaded in this codebase. Throughout this document:

| Term | Meaning |
|---|---|
| **Nave v1 API** | The public search API served by the Python/Django application |
| **`brabantcloudv1`** | The Elasticsearch index Nave queries |
| **`/api/search/v1`** | A hub3 endpoint, effectively unused (7 requests in a week, all monitoring) |
| **`/api/search/v2`** | The hub3 search API, feature-frozen, used by the DIW sites |
| **Semantic v1** | `/api/semantic/v1`, the JSON-LD/Hydra API that is the designated successor |

## What production actually does

All figures measured on 2026-09-21 from `ikuzoctl` structured logs on
`eb-acpt-ingestion` (the live serving stack), over a 7-day window, plus
Elasticsearch's own counters.

### Traffic distribution

| Path | Requests / 7 days | Consumer | p50 | p95 |
|---|---|---|---|---|
| `/{index}/_search` (ES proxy) → `brabantcloudv1` | **1,574,573** | Nave, via `elasticsearch-py/7.8.0` on Python 3.9.2 | 3.5 ms | 31 ms |
| `/api/search/v2` | 42,641 | WordPress DIW sites | 15 ms | 53 ms |
| `/api/search/v1` | 7 | monitoring only | — | — |
| `/api/es/*` | 0 | — | — | — |

**The ES proxy carries 37× more traffic than the entire v2 API.** It is mounted
at `ikuzo/service/x/esproxy/routes.go` as `/{index}/_search` (plus a
`/{index}/{documentType}/_search` variant) and hands Nave's request body to
Elasticsearch unmodified, with a groupcache layer in front. Nave reaches it
because its settings point at the public host:

```python
ES_URLS = ['https://api.brabantcloud.nl']
INDEX_NAME = 'brabantcloudv1'
```

### Index sizes

| Index | Documents | Store | Queries/min (live) |
|---|---|---|---|
| `brabantcloudv1` | 2,046,703 | 13.2 GB | 384 |
| `brabantcloudv2` | 107,975,890 | 15.2 GB | 15 |

The v2 index holds ~53× more documents in comparable space: each record explodes
into many nested documents.

### v2 API parameter usage

Of the 50+ parameters the v2 API accepts, 23 appeared in a week and the top nine
account for 99% of use:

| Parameter | Count |
|---|---|
| `itemFormat` | 37,343 |
| `qf.-meta.spec` | 31,571 |
| `qf.id` | 31,571 |
| `rows` | 6,623 |
| `format` | 5,784 |
| `facet.field` | 1,541 |
| `page` | 1,335 |
| `lang` | 973 |
| `itemTypePath` | 971 |

Then a long tail: `qf[]` (940), `facet.limit` (700), `q` (518), `sortBy` (479),
`facetBoolType` (479), `moreLikeThis` (288), `mlt.count` (288), `hqf[]` (206),
`qf.meta.spec` (27), `facet.expand` (16), `scrollID` (14), `mlt.filterkey` (4),
`facet.filter` (2), `facet` (1).

**Zero usage in the whole window:** every `collapse*` variant (`collapseOn`,
`collapseSize`, `collapseSort`, `collapseFormat`, `collapseCount`), `byQuery`,
`byDepth`, `byMimeType`, `byParent`, `byLeaf`, `byChildCount`, `byUnitID`,
`allParents`, `facet.cursor`, `facet.full`, `cursorHint`, `contextIndex`,
`batchIndex`, `batchSize`.

> **Caveat before deleting any of these:** this is 7 days of one deployment. The
> `by*` and tree parameters look like EAD/archive browsing, which may be
> seasonal or site-specific. Confirm against a longer window, or against the
> knowledge that EAD is not in play for this customer, before removing them.

Notably, `q` appears 518 times against 31,571 `qf.id` lookups: **the DIW uses the
v2 API mostly as a document store keyed by identifier, not as a search engine.**

### Performance baseline

Aggregate v2: p50 15.1 ms, p90 42.5 ms, p95 53.5 ms, p99 258.7 ms, max 5,095 ms.

Per parameter combination:

| Combination | n | p50 | p95 |
|---|---|---|---|
| `itemFormat,qf.-meta.spec,qf.id` | 31,572 | 14.8 ms | 22.8 ms |
| `rows` | 5,033 | 43.3 ms | 62.7 ms |
| `format,itemFormat` | 3,923 | 1.8 ms | 2.4 ms |
| `facet.field,format,itemFormat,page,qf[],rows` | 657 | 91.5 ms | 152.9 ms |
| **`facet.field,facet.limit,facetBoolType,format,itemFormat,itemTypePath,lang`** | **389** | **793.0 ms** | **1032.6 ms** |
| `format,itemFormat,itemTypePath,lang,mlt.count,moreLikeThis` | 284 | 22.7 ms | 59.2 ms |

One combination is ~50× slower than everything else and accounts for essentially
the entire p99. Separately, 318 requests exceeded 500 ms, almost all
`nave_buildingDepicted` lookups against `brabantse-gebouwen`, peaking near 5
seconds. **These two are the concrete optimisation targets** — not a vague
"make search faster".

## Client situation: olivere is a dead end

`github.com/olivere/elastic/v7 v7.0.32` builds the query DSL in 34 files
(~290 constructor calls: 90 `NewTermQuery`, 68 `NewBoolQuery`, 27
`NewTermsAggregation`, 11 `NewNestedAggregation`, 10 `NewNestedQuery`).

- **Last release: v7.0.32, 2022-03-19** — the version we pin. The author has
  stated publicly that a v8 will not happen and points users at the official
  client.
- **No compatibility bridge exists.** Elasticsearch's REST API compatibility
  mode is the normal way to let a 7.x client talk to an 8.x server, but it
  requires a custom `Content-Type` header, and olivere hardcodes it —
  verified in the module cache at `v7@v7.0.32/request.go:27`:
  ```go
  req.Header.Set("Content-Type", "application/json")
  ```
  So "upgrade the server, keep olivere" is not an option. **The client migration
  must land before or with the server upgrade.**

### Migrate to v9, not v8

The official client's `esdsl` package is the only ergonomic analogue to
olivere's builder API. Verified by downloading both modules:

| Module | `typedapi/esdsl` |
|---|---|
| `go-elasticsearch/v8@v8.19.7` (final v8) | **absent** |
| `go-elasticsearch/v9@v9.5.2` | **present** (1,117 files) |

Migrating to v8 means doing all the painful work and still not getting the good
API. v8 and v9 can coexist in one `go.mod`, so the migration can proceed file by
file.

`ikuzo/storage/x/elasticsearch8` (~2,500 lines of implementation, ~2,000 lines of
tests: query builder, aggregation builder, result parser, introspection, store,
client) is the existing foundation for this — 12 commits between February and
April 2026, the last one specifically about building nested filters and
aggregations against `resources.entries`. **It is paused work, not dead code,
and must not be removed in any cleanup.**

## Nested types survive the upgrade

Researched against Elastic's breaking-changes and release-notes sources.
Current GA at time of writing: 9.5.4; final 8.x minor: 8.19.

**Nothing about the `nested` field type, `nested` queries, `nested`/
`reverse_nested` aggregations, or `inner_hits` was removed, deprecated, or
semantically redefined between 7.17 and 9.5.** The double-nested mapping
(`resources` → `resources.entries`) remains fully supported and no more
expensive than it is today.

New limits introduced in 9.3/9.5 are far above our usage: we have **1 nested
parent of an allowed 50**, and 2 nested field mappings of an allowed 100
(`index.mapping.nested_objects.limit` is unchanged at 10,000).

### Breaking changes checked against this codebase

| Change (ES 8.0 unless noted) | Present here? |
|---|---|
| Sort `nested_path` / `nested_filter` removed | **Yes — one site:** `hub3/fragments/api.go:1685` |
| Joda date formats must become java-time | **Yes:** `"format": "dateOptionalTime"` in `internal/mapping/v1.go` (2×) |
| `reverse_nested` validation tightened (9.2.1, backported to 8.19) | Not used |
| `common` / `cutoff_frequency` / `moving_avg` removed | Not used |
| Aggregation order keys `_term`/`_time` → `_key` | Not used |
| `boost` on field mappings removed | Not used |
| Multi-field inside multi-field removed | Not used |
| `_type` in queries no longer matches | Not used |
| `sparse_vector`, `geo_shape` params removed | Not used |
| Mapping types removed | **Yes — the proxy route** `/{index}/{documentType}/_search` |

The `api.go:1685` fix is small and self-evident: the `_int` sort branch still
uses `NestedPath()`/`NestedFilter()`, while the `default` branch immediately
below it already uses the modern `NestedSort` API. It can be fixed today, on
7.17, independently of everything else.

### Reindex rules

| Upgrade | Rule |
|---|---|
| 7.17 → 8.x | Indices created in 7.x remain fully read/write. **No reindex required.** |
| 8.x → 9.x | Indices still carrying a 7.x creation version **must be reindexed, deleted, or marked read-only**, or nodes refuse to start. |

Our indices were created in 7.x, and upgrading to 8.x does not relabel them.
Reaching 9.x therefore requires a rebuild — which is exactly the offline-rebuild
strategy already chosen, so this is a constraint we were going to satisfy anyway.

## The decisive finding

`hub3/fragments` serves 42,641 requests a week. The ES proxy serves 1,574,573 —
and passes Nave's raw 7.x query DSL straight through to Elasticsearch without
translation.

**This means the ES upgrade is gated on Nave's query syntax, which lives in
Django code, not in this repository.** Any 7.x construct Nave uses that 8.0
removed will break the moment the server moves. We do not currently know what
those constructs are.

This is what makes the "reimplement Nave v1 in Go" goal the load-bearing piece
of work rather than a nice-to-have: it converts an unknown, unowned query
surface into one we control, test, and can migrate deliberately.

## Proposed sequence

### 0. Free wins, doable now (no dependencies)

- Fix `hub3/fragments/api.go:1685` to use `NestedSort`. Works on 7.17.
- Replace `dateOptionalTime` with an explicit java-time format in the v1 mapping.
- Investigate the 793 ms facet combination and the ~5 s `nave_buildingDepicted`
  lookups. Both are single code paths and neither needs an ES upgrade to fix.

### 1. Inventory what Nave actually sends

Before anything else. The proxy sees every query; temporary logging of request
bodies plus an analysis script yields, within a day, the exact set of DSL
constructs in use and which of them do not survive 8.0.

Without this, both the Go reimplementation and the ES upgrade are guesswork.
This also produces the acceptance criteria for step 2: a Go v1 implementation is
correct when it answers the recorded corpus identically.

### 2. Reimplement the Nave v1 API in Go

Target: `brabantcloudv1` served from hub3, so Nave can be switched off. Use the
recorded corpus from step 1 as both specification and regression suite. Build
against `go-elasticsearch/v9` + `esdsl` from the start — there is no reason to
write new code against a dead library.

### 3. Narrow `hub3/fragments` to measured usage

8,793 lines of hand-written code (plus 5,220 generated and 5,317 test). The
measurements above say which parameters earn their keep. This also retires the
`v2adapter` bridge that exists only to let the semantic API borrow v2's
olivere-based backend.

### 4. Migrate the remaining olivere call sites to v9 + esdsl

34 files, incrementally, with v8 and v9 side by side in `go.mod`.

### 5. Server upgrade with offline rebuild

7.17.28 → 7.17.29 → 8.19.x → 9.5.x, rolling, running the Upgrade Assistant
before each major hop. Build the new index offline, verify, then switch — the
approach already chosen, and the only one that satisfies the 9.x reindex rule
without downtime.

## Open questions

1. **EAD/archives** — do the unused `by*`/tree parameters still serve anyone, or
   can they go? Determines how much of step 3 is deletion versus preservation.
2. **`/api/search/v1`** (the hub3 endpoint, 7 requests/week) — retire outright?
3. **v9 client against a 7.17 server** — Elastic guarantees forward
   compatibility only (client ≥ server). If steps 2 and 4 are to land before
   step 5, this needs empirical verification; otherwise client and server must
   move in one window.
4. **`elasticsearch-py 7.8.0` on Python 3.9.2** — irrelevant if Nave is retired
   in step 2, but a blocker if Nave outlives the server upgrade.

## Appendix: unshipped packages

Separate from this plan, an audit found 41 of 100 packages are not part of the
`ikuzoctl` binary. Most are legitimate (separate CLI tools under `tools/cmd/`,
test helpers, `internal` packages). Twenty have **zero importers**:

| Package | Lines | Last touched |
|---|---|---|
| `ikuzo/service/x/{mapping,rdf,record}` | 15 each | 2020-06-10 |
| `tmp` | 21 | untracked scratch |
| `ikuzo/service/x/oaipmh/pmh` | 76 | 2022-01-24 |
| `hub3/ead/alto` | 183 | 2022-05-09 |
| `ikuzo/storage/x/file` | 699 | 2022-05-30 |
| `ikuzo/service/x/harvest` | 736 | 2022-10-14 |
| `ikuzo/rdf/schema{,/recdef}` | 260 | 2022-11-01 |
| `ikuzo/service/x/search/es` | 100 | 2023-05-07 |
| `ikuzo/rdf/api{,/v1,/v2}` | 8 | 2024-05-06 |
| `ikuzo/rdf/formats/hextuples` | 451 | 2024-11-29 |
| `ikuzo/rdf/formats/turtle` | 671 | 2026-03-19 |
| `ikuzo/storage/x/elasticsearch8` | 4,485 | 2026-04-15 |
| `ikuzo/storage/x/elasticsearch` | 1,039 | 2026-07-02 |
| `ikuzo/service/x/debug` | 152 | 2026-07-02 |

The 2020–2022 entries are safe removals. **`elasticsearch8` is explicitly not** —
it is the foundation for step 4. The two RDF format parsers and the recently
touched packages need a decision from someone who knows their intent.
