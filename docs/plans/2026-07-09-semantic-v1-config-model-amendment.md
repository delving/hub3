# Semantic V1 Configuration Model — Proposed Amendment (DRAFT, not yet ratified)

> Produced by the Ultraplan cloud session 2026-07-09 (fresh-context design pass on the three open
> config questions). Salvaged from the session stream after the approval window expired.
> Folds into docs/plans/2026-07-02-semantic-v1-greenfield.md once the four flagged decisions in
> section E are ratified.

---

# Amendment: Configuration model, per-facet v1.1 grammar, introspection principle (2026-07-09)

## A. Global server config — TOML + Go shapes

### A.1 TOML shape (final)

```toml
[semantic]
enabled = true                       # existing; *bool default-on pattern
baseURL = "/api/semantic/v1"         # existing; route mount path
# useV2Adapter / useES8Backend unchanged

[semantic.defaults]
size = 20            # 0..100; absent -> contract default 20
facetLimit = 10      # 1..100; absent -> contract default 10
facetSort = "count"  # "count" | "value"; absent -> "count"
sort = ""            # v1 sort grammar, one entry, e.g. "-dc_date"; "" = relevance

# per-org override — sparse, field-wise overlay
[org.brabantcloud.semantic.defaults]
facetLimit = 20
sort = "-nave_modified"
```

Everything else stays out of `[semantic.defaults]` deliberately:

- **`baseURL` stays at `[semantic]` level** — it is routing, not a query default, and it already exists there.
- **`publicBaseURL` is NOT added in v1.** The handler's `absBase(r)` (Host + `X-Forwarded-Proto`) covers reverse-proxy deployments. Reserve the key name at the `[semantic]` level for a future absolute-link override; adding it later is pure config addition with zero wire impact.
- **`page` default is not configurable.** Page 1 is structural, not policy.
- **Max caps are NOT config** — see A.4.

### A.2 Go struct shapes (final)

One shared config type in `ikuzo/domain` (so the org block and the global block are the same shape, one overlay function, no drift), pointer fields so "absent" is distinguishable from zero — required because `size = 0` and `sort = ""` are both *valid explicit values*:

```go
// ikuzo/domain/organization_config.go
type SemanticDefaults struct {
    Size       *int    `json:"size"`
    FacetLimit *int    `json:"facetLimit"`
    FacetSort  *string `json:"facetSort"`
    Sort       *string `json:"sort"`
}

type SemanticConfig struct {
    Defaults SemanticDefaults `json:"defaults"`
}

// added to OrganizationConfig (alongside OAIPMH, SPARQL, ElasticSearch):
Semantic SemanticConfig `json:"semantic,omitempty"`
```

```go
// ikuzo/ikuzoctl/cmd/config/semantic.go
type Semantic struct {
    Enabled       *bool                   `json:"enabled"`
    BaseURL       string                  `json:"baseURL"`
    UseV2Adapter  bool                    `json:"useV2Adapter"`
    UseES8Backend bool                    `json:"useES8Backend"`
    Defaults      domain.SemanticDefaults `json:"defaults"`
}
```

```go
// ikuzo/service/x/semanticv1 — the runtime struct from the ratified amendment
type Defaults struct {
    Size       int
    FacetLimit int
    FacetSort  string
    Sort       string // "" = relevance (no sort applied)
}

// ContractDefaults returns the compiled contract baseline: {20, 10, "count", ""}.
func ContractDefaults() Defaults

// Overlay returns d with every non-nil field of o applied on top.
func (d Defaults) Overlay(o domain.SemanticDefaults) Defaults

// Validate enforces: 0 <= Size <= MaxSize; 1 <= FacetLimit <= MaxFacetLimit;
// FacetSort in {count,value}; Sort is "" or one entry of the v1 sort grammar
// (optional "-", fieldNamePattern; a comma -> error "v1 supports exactly one sort entry").
func (d Defaults) Validate() error
```

`semanticv1` already imports `ikuzo/domain` (for `GetOrganizationFromCtx`), so `Overlay` taking `domain.SemanticDefaults` creates no new dependency.

### A.3 Resolution / merge design (per-org: ship it in v1)

**Precedence, per field, most-specific wins:**

```
explicit request param  >  [org.<id>.semantic.defaults]  >  [semantic.defaults]  >  compiled ContractDefaults()
```

**Where each step happens:**

1. **Startup (Task 10, `Semantic.AddOptions`):** `global := ContractDefaults().Overlay(cfg.Semantic.Defaults)`; validate; pass via `semanticv1.WithDefaults(global)`. Also validate every `cfg.Org[id].Semantic.Defaults` block by computing `global.Overlay(orgBlock)` and calling `Validate()` — fail startup naming the org and key.
2. **Per request (handler, Task 8):** after `domain.GetOrganizationFromCtx(r)`, compute `eff := s.defaults.Overlay(org.Config.Semantic.Defaults)`. This is four pointer checks — no new infra, matching the OAI-PMH per-org pattern. As defense-in-depth for organizations sourced outside the TOML file, the handler calls `eff.Validate()` and returns a 500 `hydra:Error` on violation (unreachable for TOML-only deployments because startup already validated).
3. `eff` is then threaded explicitly: `ParseQuery(values, eff)`, `ParseSearchBody(r, eff)`, `EncodeQuery(opts, eff)`, `BuildCollection(base, opts, res, eff)`. Explicit parameters, not fields smuggled inside `QueryOptions` — parse/encode stay pure functions and `reflect.DeepEqual`-based parity tests keep working.

**Recommendation: ship per-org in v1**, not shape-only. Rationale: (a) the marginal cost is one struct field, one 10-line overlay, one per-request call — the org is already resolved in every handler; (b) `OrganizationConfig.ElasticSearch.Defaults{FacetSize, Limit, MaxLimit}` proves multi-tenant deployments already demand per-org search defaults; shipping global-only guarantees a follow-up that touches the same handler signature chain; (c) retrofitting later would change `WithDefaults` semantics mid-flight; doing it now keeps `WithDefaults` meaning "the service-level base the org overlay applies to" from day one. Per-org **enablement** (`[org.x.semantic] enabled = false`, OAI-PMH style) is *not* shipped — global `enabled` only; adding a per-org gate later is additive.

### A.4 Max caps stay compiled contract constants

`MaxSize = 100`, `MaxFacetLimit = 100` remain constants, not config. Rationale:

- `/docs` is generated from the same tables the parser validates against ("`/docs` cannot lie"). Configurable caps would make the frozen contract text and error messages (`"size must be 0..100"`) deployment-relative, and the contract test suite — which *is* the contract (D9) — could no longer assert them.
- Config fills absent values; caps define the *surface of valid values*. Keeping the value-space compiled is the same honesty rule as keeping the parameter list compiled.
- Escape valve: a deployment that genuinely needs bigger pages is asking for a contract revision (raise the constant for everyone in v1.1), not a config knob.

Startup validation enforces `configured defaults ⊆ caps`, so config can never produce a default the parser would reject as an explicit value.

### A.5 Config × closed surface — the invariant

**Config may only fill absent parameters; it can never add, remove, rename, or re-range a parameter.** Concretely:

- The parser's known-key tables (`scalarParams`, `allOperators`, filter/facet key grammars) are compiled. No config input reaches them.
- With *any* `Defaults` value: unknown params still 400; explicit out-of-range values still 400 (`facetLimit=200` fails even if it "just wants more than the default").
- Contract test: run the same rejection suite under two different injected `Defaults` and assert identical 400 behavior.

### A.6 The `sort=` escape hatch (new grammar item, required by configurable default sort)

A configured default sort creates an expressiveness gap: with `sort = "-dc_date"` configured, the v1 grammar as ratified gives a client **no way to request relevance ordering**. Fix, shipped in v1 (Task 2):

- **`sort=` with an empty value parses to `Sort = nil` (relevance), explicitly overriding any configured default.** With no configured default it is an accepted no-op.
- Canonical encoding: when `opts.Sort == nil` **and** the effective default sort is non-empty, `EncodeQuery` emits `sort=` (empty value); when the default is `""`, it omits `sort` entirely; when `opts.Sort` equals the effective default it omits; otherwise it emits the value. This keeps `ParseQuery(EncodeQuery(o, d), d) ≡ o` mechanically for every `d`.
- `/docs` documents the empty value as "relevance (overrides the server's default sort)".

No analogous gap exists for `size`/`facetLimit`/`facetSort`: every point of their value spaces is explicitly expressible regardless of config.

### A.7 Per-deployment canonical URLs — soundness argument

`EncodeQuery` omitting *configured* defaults means canonical URLs differ across deployments and orgs. This is sound because:

- **Hypermedia links live and die inside one (deployment, org) pair.** Links embed the request host; org resolution is by domain; a link generated under org A's defaults is parsed under org A's defaults. Bijectivity is only ever exercised within a single `Defaults` value, and it holds for each one (A.6 covers the one degenerate case).
- **Default drift across restarts is the intended semantics of a default.** A URL that omits `facetLimit` *asks for the server default*; if operators change the default, the URL means the new default — exactly like page-size defaults everywhere. Clients that need pinning have it: the envelope echoes `hub3:facetLimit`/`hub3:facetSort`/`hub3:sort`, and any explicitly-sent non-default value survives in every generated link.
- **Within one response all links are generated under one `eff`** — no mixed-defaults inconsistency is possible.
- **Contract tests are deterministic by construction:** they build the service with explicit `WithDefaults` (and explicit org overlays where tested) and assert encodings against those known values; no test asserts a canonical string without owning the `Defaults` that produced it.

The envelope echo condition is refined one notch: `hub3:sort` is echoed whenever sorting is *applied*, **including when it was applied from config** — so a client can always distinguish "server sorted by -dc_date for me" from relevance without knowing the deployment's config.

### A.8 Startup validation rules (fail-fast, complete list)

In `Semantic.AddOptions`, before service construction; any violation returns an error (startup aborts) naming the TOML key and, for org blocks, the org id:

1. `size` ∈ [0, 100]; `facetLimit` ∈ [1, 100] (constants `MaxSize`/`MaxFacetLimit`).
2. `facetSort` ∈ {`count`, `value`}.
3. `sort` = `""` or a single v1 sort entry: optional `-` + `fieldNamePattern`; any comma → `"semantic.defaults.sort: v1 supports exactly one sort entry"`.
4. Rules 1–3 applied to `global.Overlay(orgBlock)` for every `[org.<id>.semantic.defaults]`.
5. `baseURL` must begin with `/` (route mount sanity; existing empty→default behavior kept).

Validation logic lives in `semanticv1.Defaults.Validate()` (single source of truth with the parser's range checks — same constants, same field-name regex); the config package only calls it.

## B. Per-field facet settings — v1.1 grammar, designed now

### B.1 GET grammar (v1.1)

`facet` remains the sole membership parameter. Options are bracketed keys mirroring the `filter[field][op]` precedent:

```
facet=dc_creator&facet=dc_date
facet[dc_creator][limit]=5
facet[dc_creator][sort]=value
```

Rules:

- Key grammar: `^facet\[([^\]\[]+)\]\[([^\]\[]+)\]$`; field must match `fieldNamePattern`; option name must be in the compiled v1.1 option set `{limit, sort}`. Unknown option → 400 `"unknown facet option %q for field %q"`.
- An option key whose field has no `facet=` membership entry → 400 `"facet[%s][%s] requires facet=%s"` (mirrors the `collapse.size requires collapse` rule).
- Duplicate option key (multiple values) → 400.
- Value ranges: `limit` ∈ [1, MaxFacetLimit]; `sort` ∈ {`count`, `value`}. Same constants and error wording as the global params.
- **Reserved, documented as future, still 400 in v1.1:** `facet[f][type]` (`terms` | `histogram` | `minmax`), `facet[f][interval]` (e.g. `1y`), `facet[f][ranges]` (e.g. `1600..1650,1650..1700`). Ranges and configurable-interval histograms are **native-backend only** (v2 has no user-bounded range aggs; its histogram is hardwired to isoDate + fixed interval), so these ship with the native backend, not with v1.1-on-bridge.

### B.2 POST grammar (v1.1)

`facets` entries become a string-OR-object union:

```json
"facets": [
  "dc_date",
  {"field": "dc_creator", "limit": 5, "sort": "value"}
]
```

- Object keys: `field` (required, `fieldNamePattern`), `limit`, `sort` — same ranges as GET; unknown keys in the object → 400 (per-entry `DisallowUnknownFields`).
- Decoded as `[]json.RawMessage`, each entry dispatched on its first byte (`"` → string, `{` → object, anything else → 400).

### B.3 Internal model and canonical encoding (v1.1)

Internal model: `Facets []FacetSelection` with `FacetSelection{Field string; Limit int; Sort string}` where zero values mean *inherit the global effective value*. Canonicalization rules (bijectivity-preserving):

- Membership entries sorted by field and **deduplicated** (see B.5 — v1 already dedupes).
- An explicit per-facet option equal to the *effective global* value (request value or configured default) normalizes to "inherit" at parse time and is therefore omitted by `EncodeQuery` — the same omit-the-default rule as top-level params, applied one level down. One `QueryOptions`, one canonical string, under every `Defaults`.
- `url.Values.Encode()`'s lexical key sort already makes `facet=…&facet[a][limit]=…&facet[b][sort]=…` deterministic; no extra ordering logic.

### B.4 Envelope + bridge (v1.1)

**Envelope:** each per-facet block (already an object: `{"hub3:field": …, "hub3:values": […]}`) additively gains `"hub3:limit"` and `"hub3:sort"` when that facet's effective value differs from the global effective value. Top-level `hub3:facetLimit`/`hub3:facetSort` keep meaning "the global effective default for facets without overrides". Context 1.1 adds `hub3:limit` (term `limit`); `hub3:sort` already exists from the v1 amendment. Purely additive: v1 clients ignore unknown keys inside facet objects.

**v2 bridge mapping (real capability today):** per-facet `limit` and `sort` are executable on v2 *now*, but only via JSON-object `facet.field` entries, and only if the global clobber is avoided:

- Emit **every** facet as a JSON-object `facet.field` entry with an **explicit `size`** (the facet's effective limit — override or global effective), e.g. `facet.field={"field":"dc_creator","size":5,"byName":true,"asc":true}`.
- **Omit `facet.limit` entirely** — `SearchRequest.Aggregations` (hub3/fragments/api.go:1179) overwrites every per-field `Size` whenever `FacetLimit != 0`. Emitting object sizes for *all* facets (not just overridden ones) is also what keeps non-overridden facets on the v1 effective default rather than falling back to v2's own `ElasticSearch.FacetSize` config (api.go:104).
- `sort=value` → `byName: true, asc: true`; `sort=count` → omit both (v2's count-desc default).

**Native-only (flagged):** `type=histogram` with configurable field/interval, `ranges`, `minmax`. Do not bridge these; the reserved grammar exists so the native backend ships them without new syntax.

### B.5 What v1 must do NOW (the additivity checklist, with task placement)

| # | Requirement | Why | Lands in |
|---|---|---|---|
| 1 | `facet[...]` GET keys → 400. Already guaranteed by the closed surface (they fall to the `unknown parameter` branch). Add an explicit contract-test case `facet%5Bdc_creator%5D%5Blimit%5D=5` → 400 so the guarantee is pinned, not incidental. | v1.1 flips a documented 400 into a feature — unambiguous | Task 8 (contract test) |
| 2 | POST `facets` parsed as `[]json.RawMessage`; non-string entry → 400 with the *specific* message `"facet entries must be strings in v1; per-facet options are not yet supported"` — not Go's generic `cannot unmarshal object into … string`. | Clear upgrade signal; the generic error names internal types | Task 3 |
| 3 | Facet membership values deduplicated (sorted-unique) at parse in both parsers. | v1.1 options are keyed by field; duplicate membership would make option attachment ambiguous | Tasks 2 + 3 |
| 4 | Envelope facet blocks are JSON objects (they are, per Task 7's `buildFacetBlocks`); add a contract-test assertion that each `hub3:facets` entry is an object with `hub3:field`/`hub3:values`. | Per-facet echo in v1.1 must be additive key insertion | Tasks 7 + 8 |
| 5 | Keep `QueryOptions.Facets []string` in v1 (decision — see flagged decision 4). | — | Task 2 (no change) |
| 6 | **Bridge correctness fix:** Task 5's table maps `FacetSort=value` → `facet.sort=value`, but v2's global `facet.sort` is only honored in a separate handler path, not in the `NewSearchRequest`/`ExecuteWithParallelAggregations` path the bridge uses. The v1 translator must instead emit JSON-object `facet.field` entries with `byName:true, asc:true` when `FacetSort == "value"` (plain names when `count`). This is a v1 bug fix that *also* pre-builds the exact object-emission machinery v1.1 needs. | v1 `facetSort=value` must actually work; contract honesty | Task 5 (amend mapping table + tests) |
| 7 | `/docs` documents `facetLimit`/`facetSort`/`size`/`sort` defaults from the resolved `Defaults` struct (not literals), and documents the `sort=` empty-value relevance override. | single source of truth; A.6 | Task 9 |
| 8 | All defaults plumbing from section A: `Defaults` type + `ContractDefaults` + `Validate` (Task 2); `ParseQuery(values, d)` / `ParseSearchBody(r, d)` / `EncodeQuery(opts, d)` (Tasks 2–3); `BuildCollection(base, opts, res, d)` + echo fields (Task 7); `WithDefaults` + per-request org overlay in handlers (Task 8); `domain.SemanticDefaults` + `OrganizationConfig.Semantic` + `AddOptions` validation/wiring (Task 10). | ratified amendments + this design | Tasks 2, 3, 7, 8, 10 |
| 9 | Context file: `facetLimit`, `facetSort`, `sort` terms (already amended into Task 1). v1.1 will add `limit` in context **1.1**, never by editing 1.0. | frozen-context rule | Task 1 |

## C. Phase-2 introspection principle

**Introspection describes; it never gates (D8 stands).** The parser consults nothing but its compiled tables and the field-name regex — that stays true forever. Introspection is a read-only *description* of two distinct sources, and the design principle is to never blur them:

- **From data (the backend), always live:** which fields actually exist (ES mappings / field-caps, later triple-store predicates), their types, value samples, cardinalities. This is per-org index truth and can change with every ingest. It is discovered through the same `SearchStore` seam (a future `Introspect` method or sibling interface), so a native backend replaces it the same way it replaces search. A field that introspection doesn't list still filters fine — it matches nothing.
- **From config, curated annotations only:** human labels, curated/recommended facet lists, display ordering, and the *effective defaults* — the same resolved `Defaults` value the parser and `/docs` already use. Config describes *presentation intent*, never existence, and a config entry for a nonexistent field is harmless (it annotates nothing).

The single-source-of-truth rule from `/docs` extends verbatim: whatever endpoint introspection gets, its "defaults" section is rendered from the same per-request `eff` the parser would use — including org overlays — so it cannot disagree with behavior.

**What v1 config reserves for it: only naming space.** `[semantic.defaults]` stays deliberately narrow so a future curated block has an obvious, non-colliding home:

```toml
[[semantic.facets]]          # phase 2 — NOT shipped, NOT parsed in v1
field = "dc_creator"
label = "Creator"
# [[org.<id>.semantic.facets]] mirrors it per org
```

Nothing in v1 parses or validates such a block (unknown TOML keys are ignored by the section decoding, so its later arrival is config-additive). The legacy `V1Config.DefaultFacets []FacetConfig{Field, Name, MaxSize}` is its conceptual ancestor and is explicitly **not** reused — it dies with the Task 11 cutover scope. One consistent-but-deferred extension noted for the record: a curated facet list *could* someday also serve as `defaults.facets` (facets applied when the `facet` param is absent) — that would still obey the fill-absent rule, since `facet` is an existing parameter — but it materially changes response payloads and is deferred until introspection itself lands.

## D. Contract-test additions implied by this design

1. **Defaults injection:** service built with a non-contract `Defaults` (e.g. `{Size: 5, FacetLimit: 3, FacetSort: "value", Sort: "-dc_date"}`); empty-param request must reach the store with those values and echo them (`hub3:facetLimit`, `hub3:facetSort` when facets requested, `hub3:sort` — including when the sort came from config).
2. **Canonical omission is defaults-relative:** encode the same `QueryOptions` under two different `Defaults`; assert the omitted keys differ and `ParseQuery(EncodeQuery(o,d),d) ≡ o` under each.
3. **Closed surface is config-independent:** the dead-param rejection suite plus `facetLimit=200` / `size=101` run under two `Defaults` values with identical 400 results.
4. **`sort=` relevance override:** with default sort configured, `sort=` yields `Sort == nil`; canonical encoding of nil-sort emits `sort=` iff the effective default sort is non-empty; round-trips.
5. **v1.1 pre-pinning:** `facet[dc_creator][limit]=5` → 400 `hydra:Error`; POST object facet entry → 400 containing `"facet entries must be strings in v1"`; duplicate `facet=a&facet=a` normalizes to one; every `hub3:facets` entry is a JSON object.
6. **Per-org overlay:** two orgs via the test org-middleware stand-in, one with `Semantic.Defaults` overrides; assert parse results, echoes, and link encodings differ accordingly, and that link URLs generated under an org re-parse identically under that org.
7. **Config-package startup tests** (in `ikuzo/ikuzoctl/cmd/config`): `facetSort = "alpha"`, `size = 200`, `sort = "-a,b"`, bad org block → `AddOptions` error naming key (and org); valid config wires `WithDefaults` with the overlay result.
8. **Bridge facetSort test (Task 5):** `FacetSort: "value"` → object-form `facet.field` with `byName:true, asc:true`; `count` → plain names; and (v1.1, recorded now for the bridge's v1.1 test) per-facet limits → object entries with explicit `size` and **no** `facet.limit` key.

## E. Flagged user-facing decisions (reasonable people could differ)

1. **Per-org defaults: ship in v1 vs shape-only.** *Recommend ship in v1.* Cost is one struct field + a four-pointer overlay per request; hub3 is org-first everywhere else (`ElasticSearch.Defaults`, OAI-PMH), and deferring guarantees a follow-up through the same signatures. The counterargument (smaller v1 test surface) buys little because the overlay is where all the risk isn't — the parser is unchanged.
2. **Caps (size ≤ 100, facetLimit ≤ 100): compiled constants vs config.** *Recommend compiled.* `/docs`-cannot-lie and tests-are-the-contract both break if the value space is deployment-relative; a deployment needing more is a contract revision, not a knob. This is the decision most likely to get pushback from operators — decide it explicitly.
3. **`sort=` empty value as the explicit-relevance escape hatch.** *Recommend accept.* Without it, any deployment that configures a default sort silently removes relevance ranking from the API — dishonest by omission. The alternative (a reserved keyword like `sort=relevance`) collides with the legal field-name space; the empty value cannot collide and round-trips cleanly.
4. **`QueryOptions.Facets`: keep `[]string` in v1 vs introduce `[]FacetSelection` now.** *Recommend keep `[]string`.* It is an internal type behind a compiled seam — the v1.1 refactor is compiler-checked and confined to five known touchpoints (`query.go`, `parse_get.go`, `parse_post.go`, `v2bridge/translate.go`, contract-test fakes); the wire grammar is what must be future-proof, and B.1–B.2 make it so. A struct with dead fields today invites the bridge to half-implement them. (If the team prefers zero-churn later, switching now is safe because no task has been executed yet — but the ratified plan snippets would all need mechanical edits.)

---

### Critical Files for Implementation

- /home/user/repo/docs/plans/2026-07-02-semantic-v1-greenfield.md — the plan this amendment folds into (Tasks 1–10 placements in B.5)
- /home/user/repo/ikuzo/ikuzoctl/cmd/config/semantic.go — gains `Defaults domain.SemanticDefaults`, startup validation, `WithDefaults` wiring (Task 10)
- /home/user/repo/ikuzo/domain/organization_config.go — gains `SemanticDefaults` + `OrganizationConfig.Semantic` (per-org overlay source)
- /home/user/repo/hub3/fragments/api.go — v2 seam the bridge targets: `NewFacetField` object form (line 103), the `facet.limit` clobber (line 1179), `FacetField.byName/asc/size`
- /home/user/repo/ikuzo/ikuzoctl/cmd/config/config.go — section registration pattern (`Semantic` already in the `Config` struct and options list)