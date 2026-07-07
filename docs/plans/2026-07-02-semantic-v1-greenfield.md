# Semantic V1 Greenfield Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the 14k-line accreted semantic API surface with a ~1.2k-line greenfield package implementing a frozen, honest v1 contract that wraps v2 search.

**Architecture:** One new package `ikuzo/service/x/semanticv1` owns the contract: types, GET/POST parsing (bijective with a canonical query-string encoding), validation, Hydra/JSON-LD envelope, handlers. The v2 compatibility layer lives in `ikuzo/service/x/semanticv1/internal/v2bridge` — Go's `internal/` rule makes it unimportable outside semanticv1, so retiring it later (when a native backend lands) is a deletion, not a migration. The bridge is two pure mappings around v2's own machinery: `QueryOptions → url.Values` (consumed by v2's public param parser `fragments.NewSearchRequest`) on the way in, and v2's decoded `ScrollResultV4` response `→ semanticv1.SearchResult` on the way out — all query parsing and ES response decoding stays v2's battle-tested code. Item content passes through as-is (the `semantic` view produced by `hub3/fragments`); the API only wraps it in an envelope. Old packages are deleted at cutover, not before.

**Tech Stack:** Go stdlib + chi router + olivere/elastic (via existing v2 infra) + matryer/is for tests. No new dependencies.

## Global Constraints

- Public field names use the v2 underscore searchLabel form (`dc_creator`), passed verbatim to v2. **No colon codec anywhere.**
- GET and POST express the identical `QueryOptions`; every valid request has a canonical GET query-string encoding (used for all response links).
- Unknown query parameters and unknown POST fields are **rejected with a 400 `hydra:Error`** naming the offender. This is the backwards-compatibility mechanic: absent ≠ broken, so params can be added later without ambiguity.
- Item documents are never mutated: no injected `@context`, no `hub3:navigation`, no `_warning` inside members. Envelope data lives beside content, never in it.
- Responses reference the versioned context URL (`{base}/contexts/hub3/1.0/context.jsonld`); the inline context map is not embedded in responses.
- Every response is `application/ld+json`.
- No content/schema validation of field names in phase 1: fields are opaque strings matching `^[A-Za-z][A-Za-z0-9_.]*$`; unknown fields flow to v2 and return 0 results harmlessly.
- Old code (`ikuzo/service/x/semantic`, `ikuzo/domain/semantic`, old adapter files, `ikuzo/storage/x/elasticsearch8`, `ikuzo/storage/x/elasticsearch/semantic_store*`) is untouched until Task 11 cutover.
- Max line length 140, gofmt, table-driven tests, errors wrapped with fmt.Errorf %w.

---

# Part 1 — The Contract (review this before executing)

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/semantic/v1/` | Hydra EntryPoint |
| GET | `/api/semantic/v1/docs` | `hydra:ApiDocumentation` — documents exactly the surface below, nothing else |
| GET | `/api/semantic/v1/search` | Search |
| POST | `/api/semantic/v1/search` | Search, JSON body, same capabilities |
| GET | `/api/semantic/v1/resource/{id}` | Item detail — hubID or URL-escaped entryURI; content as-is |
| GET | `/api/semantic/v1/contexts/hub3/1.0/context.jsonld` | The versioned JSON-LD context (static artifact) |

## GET parameters (complete, closed list)

| Param | Values | Notes |
|---|---|---|
| `query` | free text | v2 `query` |
| `query.fields` | comma list of fields | v2 `searchFields` |
| `query.operator` | `AND` (default) \| `OR` | |
| `query.fuzzy` | `true` | appends `~` per term |
| `filter[{field}][{op}]` | value; repeatable | ops: `eq neq in nin gt gte lt lte contains startswith exists between`; `between` value is `min,max`; `exists` value ignored |
| `hfilter[{field}][{op}]` | same | hidden: applied, excluded from `hub3:activeFilters` |
| `facet` | field name; repeatable | v2 `facet.field` |
| `facetLimit` | int 1..100, default 10 | **global** (matches v2 capability; per-facet limits are phase 2) |
| `facetSort` | `count` (default) \| `value` | global |
| `facetBool` | `and` \| `or` | v2 `facetBoolType` |
| `sort` | `field` or `-field`; **single** | second `sort` param → 400 (v2 supports one) |
| `page` | int ≥1, default 1 | |
| `size` | int 0..100, default 20 | `0` = facets only |
| `collapse` | field | v2 `collapseOn` |
| `collapse.size` | int ≥1 | v2 `collapseSize` |
| `collapse.sort` | field (no direction — v2 limitation) | v2 `collapseSort` |
| `debug` | `true` | response gains `hub3:debug` with the translated v2 params |

Anything else → `400 hydra:Error "unknown parameter: {name}"`.

## POST body (complete)

```json
{
  "@context": "https://{host}/api/semantic/v1/contexts/hub3/1.0/context.jsonld",
  "@type": "hub3:SearchQuery",
  "query": {"value": "rembrandt", "fields": ["dc_title"], "operator": "AND", "fuzzy": false},
  "filters": [
    {"@type": "hub3:Filter", "field": "dc_creator", "operator": "eq", "values": ["Rembrandt"], "hidden": false},
    {"@type": "hub3:Filter", "field": "dc_date", "operator": "between", "values": ["1600", "1700"]}
  ],
  "facets": ["dc_creator", "edm_dataProvider"],
  "facetLimit": 20,
  "facetSort": "count",
  "facetBool": "and",
  "sort": {"field": "dc_date", "desc": true},
  "page": 1,
  "size": 20,
  "collapse": {"field": "edm_isShownBy", "size": 5, "sort": "dc_date"},
  "debug": false
}
```

One filter type. Unknown `@type` on a filter → 400 (no silent default). Unknown top-level keys → 400. `@context`/`@type` on the envelope are accepted and not otherwise processed in v1 (documented as such — no false JSON-LD-processing claims).

## Response envelope

```json
{
  "@context": "https://{host}/api/semantic/v1/contexts/hub3/1.0/context.jsonld",
  "@id": "https://{host}/api/semantic/v1/search?query=rembrandt",
  "@type": ["hydra:Collection", "schema:SearchResultsPage"],
  "hydra:totalItems": 1234,
  "hydra:member": [ { "...item content verbatim (SemanticView)..." } ],
  "hydra:view": {
    "@id": ".../search?page=1&query=rembrandt",
    "@type": "hydra:PartialCollectionView",
    "hydra:first": ".../search?page=1&query=rembrandt",
    "hydra:next": ".../search?page=2&query=rembrandt",
    "hydra:previous": null
  },
  "hub3:facets": [{
    "hub3:field": "dc_creator",
    "hub3:values": [{"hub3:value": "Rembrandt", "hub3:count": 41, "hub3:selected": false,
                     "hub3:applyURL": ".../search?filter%5Bdc_creator%5D%5Beq%5D=Rembrandt&query=rembrandt"}]
  }],
  "hub3:activeFilters": [{"hub3:field": "dc_creator", "hub3:operator": "eq", "hub3:values": ["Rembrandt"],
                          "hub3:removeURL": ".../search?query=rembrandt"}],
  "hub3:timing": {"hub3:took": 12, "hub3:unit": "ms"},
  "hub3:debug": {"hub3:v2Params": {"...": "..."}}
}
```

All `hub3:*` terms and the hydra aliases are defined in the served 1.0 context file. View/apply/remove URLs are built from the **canonical encoding of the parsed QueryOptions** — identical for GET and POST requests (this fixes the old POST-pagination-drops-the-query bug by construction).

Errors: `{"@context": ctxURL, "@type": "hydra:Error", "hydra:title": "...", "hydra:description": "...", "hydra:statusCode": 400}`.

## Decisions to ratify (D1–D9)

- **D1 — underscore field names.** `dc_creator` public, no codec. Context maps them to IRIs later.
- **D2 — global facetLimit/facetSort** instead of per-facet: matches what v2 can execute; per-facet returns with the native backend.
- **D3 — single sort.** Multi-sort → 400 (v2 executes one; silent truncation is dishonest).
- **D4 — cut in v1** (all currently dead or broken; reintroduction is additive): geo filters, `cursor`/facet cursors, `contextIndex`, `languages`, `expand`, `fields` (selection), `detailLevel`, `include=` providers, `peek` (use `size=0`), boost.
- **D5 — cut introspection endpoints** (`/introspect/*`) — content discovery is phase 2 per scope.
- **D6 — cut typed search** (`/type/{t}/search`, `/type/{t}/docs`) — resource types are content coupling; phase 2.
- **D7 — cut search-context navigation** (`?context=` token, `/contexts/query/` CRUD) — in-memory, predictable tokens, breaks behind LB; navigation returns with a real cursor design.
- **D8 — no schema gating.** Field names validated by regex only; the compiled-in EDM/nave registry leaves the request path entirely.
- **D9 — the contract tests are the contract.** `contract_test.go` (Task 8) is the acceptance suite any future backend must pass.
- **D10 — the v2 compatibility layer is internal and retirable.** It lives in `semanticv1/internal/v2bridge`, enters v2 through its public param contract (`url.Values` → `fragments.NewSearchRequest`) and leaves through v2's decoded response (`ScrollResultV4` → `SearchResult`). No other package can import it (compiler-enforced); retirement = delete `internal/v2bridge` + the one re-export file when the native backend passes the contract suite.

---

# Ratification record (2026-07-07)

All decisions D1–D10 ratified by the user (review via the design artifact + question rounds), with these **amendments** that executors must apply on top of the task code below:

- **D2 amendment — configurable facet defaults.** `facetLimit`/`facetSort` stay global params, but their *defaults* come from server config (TOML `[semantic]` section, e.g. `facetLimit = 5`, `facetSort = "count"`), not hardcoded constants. Effective values are echoed in the envelope (see envelope amendment). Per-facet options (`facet[f].limit`) are v1.1 additive; the bridge can construct richer v2 params then.
- **D3 amendment — sort syntax + configurable default.** The `sort` param uses comma-list *syntax* from day one (`sort=-dc_date`); v1 accepts exactly ONE entry — a second entry → 400 ("only one sort field supported in v1") — so multi-sort later needs no new syntax. Server config gains a default sort (`sort = "-dc_date"`) applied when the param is absent; translated to v2 `sortBy`/`sortAsc`. Effective sort echoed in the envelope.
- **Defaults plumbing.** `ParseQuery`/`ParseSearchBody` take a `Defaults` struct (`Size`, `FacetLimit`, `FacetSort`, `Sort`) supplied via a `WithDefaults` service option and wired from config in Task 10; `defaultOptions()` derives from it. `EncodeQuery` omits values equal to the *configured* defaults.
- **Envelope amendment — effective-settings echo.** `BuildCollection` adds `"hub3:facetLimit"`, `"hub3:facetSort"` (when facets requested) and `"hub3:sort"` (when sorting applied) at the top level; the three terms are added to the context file (Task 1 term list + test).
- **Vocabulary IRI ratified:** `"hub3": "http://schemas.delving.eu/hub3/"` (slash namespace, matching the existing `http://schemas.delving.eu/nave/terms/` precedent). Replace the `https://hub3.delving.org/ns/hub3#` placeholder in Task 1's context file.
- Envelope shape otherwise ratified as designed.

# Part 2 — Implementation Tasks

### Task 1: Versioned JSON-LD context artifact

**Files:**
- Create: `ikuzo/service/x/semanticv1/contexts/hub3/1.0/context.jsonld`
- Create: `ikuzo/service/x/semanticv1/context.go`
- Test: `ikuzo/service/x/semanticv1/context_test.go`

**Interfaces:**
- Produces: `ContextJSONLD []byte` (embedded), `ContextURLPath = "/contexts/hub3/1.0/context.jsonld"`, `ContextURL(baseURL string) string`.

- [ ] **Step 1: Write the failing test**

```go
package semanticv1

import (
	"encoding/json"
	"testing"

	"github.com/matryer/is"
)

func TestContextArtifact(t *testing.T) {
	is := is.New(t)

	var doc map[string]any
	is.NoErr(json.Unmarshal(ContextJSONLD, &doc))

	ctx, ok := doc["@context"].(map[string]any)
	is.True(ok)

	// every term the envelope emits must be defined
	for _, term := range []string{
		"hydra", "hub3", "schema",
		"member", "totalItems", "view", "first", "next", "previous",
		"facets", "activeFilters", "timing", "debug",
		"field", "values", "value", "count", "selected", "applyURL", "removeURL",
		"operator", "took", "unit", "v2Params",
	} {
		_, defined := ctx[term]
		is.True(defined) // term must be in context: term
	}

	is.Equal(ContextURL("https://example.org/api/semantic/v1"),
		"https://example.org/api/semantic/v1/contexts/hub3/1.0/context.jsonld")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/service/x/semanticv1/ -run TestContextArtifact -v`
Expected: FAIL (package does not exist yet / undefined: ContextJSONLD)

- [ ] **Step 3: Write the context file and loader**

`contexts/hub3/1.0/context.jsonld`:

```json
{
  "@context": {
    "@version": 1.1,
    "hydra": "http://www.w3.org/ns/hydra/core#",
    "schema": "https://schema.org/",
    "hub3": "https://hub3.delving.org/ns/hub3#",
    "member": "hydra:member",
    "totalItems": "hydra:totalItems",
    "view": {"@id": "hydra:view", "@type": "@id"},
    "first": {"@id": "hydra:first", "@type": "@id"},
    "next": {"@id": "hydra:next", "@type": "@id"},
    "previous": {"@id": "hydra:previous", "@type": "@id"},
    "facets": "hub3:facets",
    "activeFilters": "hub3:activeFilters",
    "timing": "hub3:timing",
    "debug": "hub3:debug",
    "field": "hub3:field",
    "values": "hub3:values",
    "value": "hub3:value",
    "count": "hub3:count",
    "selected": "hub3:selected",
    "applyURL": {"@id": "hub3:applyURL", "@type": "@id"},
    "removeURL": {"@id": "hub3:removeURL", "@type": "@id"},
    "operator": "hub3:operator",
    "took": "hub3:took",
    "unit": "hub3:unit",
    "v2Params": "hub3:v2Params"
  }
}
```

`context.go`:

```go
// Package semanticv1 implements the frozen Semantic API v1 contract:
// a JSON-LD/Hydra wrapper around v2 search. See
// docs/plans/2026-07-02-semantic-v1-greenfield.md for the contract.
package semanticv1

import _ "embed"

//go:embed contexts/hub3/1.0/context.jsonld
var ContextJSONLD []byte

// ContextURLPath is the route (relative to the API base) where the
// versioned context is served.
const ContextURLPath = "/contexts/hub3/1.0/context.jsonld"

// ContextURL returns the absolute context URL for a given API base URL.
func ContextURL(baseURL string) string {
	return baseURL + ContextURLPath
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/service/x/semanticv1/ -run TestContextArtifact -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ikuzo/service/x/semanticv1/
git commit -m "feat(semanticv1): versioned JSON-LD context artifact (1.0)"
```

### Task 2: Contract types + canonical GET encoding (bijective)

**Files:**
- Create: `ikuzo/service/x/semanticv1/query.go` (types + `EncodeQuery`)
- Create: `ikuzo/service/x/semanticv1/parse_get.go` (`ParseQuery`)
- Test: `ikuzo/service/x/semanticv1/query_test.go`

**Interfaces:**
- Produces:
  - `type Operator string` with constants `OpEq OpNeq OpIn OpNin OpGt OpGte OpLt OpLte OpContains OpStartsWith OpExists OpBetween` and `func (o Operator) Valid() bool`
  - `type Filter struct { Field string; Operator Operator; Values []string; Hidden bool }`
  - `type TextQuery struct { Value string; Fields []string; Operator string; Fuzzy bool }`
  - `type Sort struct { Field string; Desc bool }`
  - `type Collapse struct { Field string; Size int; Sort string }`
  - `type QueryOptions struct { Query *TextQuery; Filters []Filter; Facets []string; FacetLimit int; FacetSort string; FacetAnd bool; Sort *Sort; Page, Size int; Collapse *Collapse; Debug bool }` with defaults Page=1, Size=20, FacetLimit=10, FacetSort="count"
  - `func ParseQuery(values url.Values) (*QueryOptions, error)` — rejects unknown params
  - `func EncodeQuery(opts *QueryOptions) url.Values` — canonical (sorted, defaults omitted); `ParseQuery(EncodeQuery(o))` ≡ `o`
  - `var fieldNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.]*$`)`
  - `type ContractError struct { Status int; Title, Description string }` implementing `error`

- [ ] **Step 1: Write the failing tests**

```go
package semanticv1

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/matryer/is"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    *QueryOptions
		wantErr string
	}{
		{
			name: "defaults",
			raw:  "",
			want: &QueryOptions{Page: 1, Size: 20, FacetLimit: 10, FacetSort: "count"},
		},
		{
			name: "full surface",
			raw: "query=rembrandt&query.fields=dc_title,dc_description&query.operator=OR&query.fuzzy=true" +
				"&filter%5Bdc_creator%5D%5Beq%5D=Rembrandt&hfilter%5Bmeta_tags%5D%5Beq%5D=highlight" +
				"&facet=dc_creator&facet=edm_dataProvider&facetLimit=25&facetSort=value&facetBool=or" +
				"&sort=-dc_date&page=2&size=50&collapse=edm_isShownBy&collapse.size=5&collapse.sort=dc_date&debug=true",
			want: &QueryOptions{
				Query:      &TextQuery{Value: "rembrandt", Fields: []string{"dc_title", "dc_description"}, Operator: "OR", Fuzzy: true},
				Filters: []Filter{
					{Field: "dc_creator", Operator: OpEq, Values: []string{"Rembrandt"}},
					{Field: "meta_tags", Operator: OpEq, Values: []string{"highlight"}, Hidden: true},
				},
				Facets: []string{"dc_creator", "edm_dataProvider"}, FacetLimit: 25, FacetSort: "value", FacetAnd: false,
				Sort: &Sort{Field: "dc_date", Desc: true}, Page: 2, Size: 50,
				Collapse: &Collapse{Field: "edm_isShownBy", Size: 5, Sort: "dc_date"}, Debug: true,
			},
		},
		{
			name: "between filter",
			raw:  "filter%5Bdc_date%5D%5Bbetween%5D=1600,1700",
			want: &QueryOptions{Page: 1, Size: 20, FacetLimit: 10, FacetSort: "count",
				Filters: []Filter{{Field: "dc_date", Operator: OpBetween, Values: []string{"1600", "1700"}}}},
		},
		{
			name: "in via repeated values",
			raw:  "filter%5Bdc_type%5D%5Bin%5D=Painting&filter%5Bdc_type%5D%5Bin%5D=Drawing",
			want: &QueryOptions{Page: 1, Size: 20, FacetLimit: 10, FacetSort: "count",
				Filters: []Filter{{Field: "dc_type", Operator: OpIn, Values: []string{"Painting", "Drawing"}}}},
		},
		{name: "unknown param rejected", raw: "detailLevel=full", wantErr: "unknown parameter"},
		{name: "unknown operator rejected", raw: "filter%5Bdc_creator%5D%5Bfuzzy%5D=x", wantErr: "unknown operator"},
		{name: "bad field name rejected", raw: "filter%5B%3Bdrop%5D%5Beq%5D=x", wantErr: "invalid field"},
		{name: "second sort rejected", raw: "sort=a&sort=b", wantErr: "single sort"},
		{name: "size above max rejected", raw: "size=101", wantErr: "size"},
		{name: "negative page rejected", raw: "page=-1", wantErr: "page"},
		{name: "between needs two values", raw: "filter%5Bdc_date%5D%5Bbetween%5D=1600", wantErr: "between"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			vals, err := url.ParseQuery(tt.raw)
			is.NoErr(err)

			got, err := ParseQuery(vals)
			if tt.wantErr != "" {
				is.True(err != nil)
				is.True(contains(err.Error(), tt.wantErr))
				return
			}
			is.NoErr(err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestEncodeQueryRoundTrip(t *testing.T) {
	is := is.New(t)
	opts := &QueryOptions{
		Query:   &TextQuery{Value: "night watch", Fuzzy: true, Operator: "AND"},
		Filters: []Filter{{Field: "dc_creator", Operator: OpIn, Values: []string{"a", "b"}}},
		Facets:  []string{"dc_creator"}, FacetLimit: 10, FacetSort: "count",
		Sort: &Sort{Field: "dc_date", Desc: true}, Page: 3, Size: 20,
	}
	back, err := ParseQuery(EncodeQuery(opts))
	is.NoErr(err)
	if !reflect.DeepEqual(back, opts) {
		t.Errorf("round trip mismatch:\ngot  %+v\nwant %+v", back, opts)
	}
	// canonical: defaults omitted
	is.Equal(EncodeQuery(&QueryOptions{Page: 1, Size: 20, FacetLimit: 10, FacetSort: "count"}).Encode(), "")
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
```

(add `"strings"` to imports)

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./ikuzo/service/x/semanticv1/ -run "TestParseQuery|TestEncodeQueryRoundTrip" -v`
Expected: FAIL (undefined types)

- [ ] **Step 3: Implement types + EncodeQuery (`query.go`)**

```go
package semanticv1

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

type Operator string

const (
	OpEq         Operator = "eq"
	OpNeq        Operator = "neq"
	OpIn         Operator = "in"
	OpNin        Operator = "nin"
	OpGt         Operator = "gt"
	OpGte        Operator = "gte"
	OpLt         Operator = "lt"
	OpLte        Operator = "lte"
	OpContains   Operator = "contains"
	OpStartsWith Operator = "startswith"
	OpExists     Operator = "exists"
	OpBetween    Operator = "between"
)

var allOperators = map[Operator]bool{
	OpEq: true, OpNeq: true, OpIn: true, OpNin: true, OpGt: true, OpGte: true,
	OpLt: true, OpLte: true, OpContains: true, OpStartsWith: true, OpExists: true, OpBetween: true,
}

func (o Operator) Valid() bool { return allOperators[o] }

var fieldNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.]*$`)

type Filter struct {
	Field    string
	Operator Operator
	Values   []string
	Hidden   bool
}

type TextQuery struct {
	Value    string
	Fields   []string
	Operator string // AND (default) | OR
	Fuzzy    bool
}

type Sort struct {
	Field string
	Desc  bool
}

type Collapse struct {
	Field string
	Size  int
	Sort  string
}

type QueryOptions struct {
	Query      *TextQuery
	Filters    []Filter
	Facets     []string
	FacetLimit int
	FacetSort  string // count | value
	FacetAnd   bool
	Sort       *Sort
	Page       int
	Size       int
	Collapse   *Collapse
	Debug      bool
}

const (
	DefaultPage       = 1
	DefaultSize       = 20
	MaxSize           = 100
	DefaultFacetLimit = 10
	MaxFacetLimit     = 100
)

func defaultOptions() *QueryOptions {
	return &QueryOptions{Page: DefaultPage, Size: DefaultSize, FacetLimit: DefaultFacetLimit, FacetSort: "count"}
}

// ContractError is a client-facing request error rendered as hydra:Error.
type ContractError struct {
	Status      int
	Title       string
	Description string
}

func (e *ContractError) Error() string { return fmt.Sprintf("%s: %s", e.Title, e.Description) }

func badRequest(title, format string, args ...any) *ContractError {
	return &ContractError{Status: 400, Title: title, Description: fmt.Sprintf(format, args...)}
}

// EncodeQuery renders opts as the canonical GET encoding: sorted keys via
// url.Values.Encode, defaults omitted. ParseQuery(EncodeQuery(o)) == o.
func EncodeQuery(opts *QueryOptions) url.Values {
	v := url.Values{}

	if q := opts.Query; q != nil {
		if q.Value != "" {
			v.Set("query", q.Value)
		}
		if len(q.Fields) > 0 {
			v.Set("query.fields", strings.Join(q.Fields, ","))
		}
		if q.Operator != "" && q.Operator != "AND" {
			v.Set("query.operator", q.Operator)
		}
		if q.Fuzzy {
			v.Set("query.fuzzy", "true")
		}
	}

	for _, f := range opts.Filters {
		key := fmt.Sprintf("filter[%s][%s]", f.Field, f.Operator)
		if f.Hidden {
			key = fmt.Sprintf("hfilter[%s][%s]", f.Field, f.Operator)
		}
		switch f.Operator {
		case OpExists:
			v.Add(key, "true")
		case OpBetween:
			v.Add(key, f.Values[0]+","+f.Values[1])
		default:
			for _, val := range f.Values {
				v.Add(key, val)
			}
		}
	}

	for _, f := range opts.Facets {
		v.Add("facet", f)
	}
	if opts.FacetLimit != DefaultFacetLimit {
		v.Set("facetLimit", strconv.Itoa(opts.FacetLimit))
	}
	if opts.FacetSort != "count" {
		v.Set("facetSort", opts.FacetSort)
	}
	if opts.FacetAnd {
		v.Set("facetBool", "and")
	}

	if s := opts.Sort; s != nil {
		if s.Desc {
			v.Set("sort", "-"+s.Field)
		} else {
			v.Set("sort", s.Field)
		}
	}

	if opts.Page != DefaultPage {
		v.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != DefaultSize {
		v.Set("size", strconv.Itoa(opts.Size))
	}

	if c := opts.Collapse; c != nil {
		v.Set("collapse", c.Field)
		if c.Size > 0 {
			v.Set("collapse.size", strconv.Itoa(c.Size))
		}
		if c.Sort != "" {
			v.Set("collapse.sort", c.Sort)
		}
	}

	if opts.Debug {
		v.Set("debug", "true")
	}

	return v
}
```

(add `"strings"` to imports)

- [ ] **Step 4: Implement ParseQuery (`parse_get.go`)**

```go
package semanticv1

import (
	"net/url"
	"strconv"
	"strings"
)

var filterKeyPattern = regexp.MustCompile(`^(h?filter)\[([^\]\[]+)\]\[([^\]\[]+)\]$`)

var scalarParams = map[string]bool{
	"query": true, "query.fields": true, "query.operator": true, "query.fuzzy": true,
	"facetLimit": true, "facetSort": true, "facetBool": true, "sort": true,
	"page": true, "size": true, "collapse": true, "collapse.size": true,
	"collapse.sort": true, "debug": true,
}

// ParseQuery parses the GET surface into QueryOptions. Unknown parameters,
// unknown operators, and out-of-range values are rejected with ContractError.
func ParseQuery(values url.Values) (*QueryOptions, error) {
	opts := defaultOptions()

	for key, vals := range values {
		switch {
		case key == "facet":
			for _, f := range vals {
				if !fieldNamePattern.MatchString(f) {
					return nil, badRequest("Invalid request", "invalid field name %q", f)
				}
				opts.Facets = append(opts.Facets, f)
			}
		case scalarParams[key]:
			if len(vals) > 1 && key == "sort" {
				return nil, badRequest("Invalid request", "only a single sort parameter is supported")
			}
			if err := applyScalar(opts, key, vals[0]); err != nil {
				return nil, err
			}
		case strings.HasPrefix(key, "filter[") || strings.HasPrefix(key, "hfilter["):
			f, err := parseFilterParam(key, vals)
			if err != nil {
				return nil, err
			}
			opts.Filters = append(opts.Filters, *f)
		default:
			return nil, badRequest("Invalid request", "unknown parameter: %s", key)
		}
	}

	sortFilters(opts.Filters)
	sort.Strings(opts.Facets)

	return opts, nil
}

func parseFilterParam(key string, vals []string) (*Filter, error) {
	m := filterKeyPattern.FindStringSubmatch(key)
	if m == nil {
		return nil, badRequest("Invalid request", "malformed filter parameter: %s", key)
	}
	field, op := m[2], Operator(m[3])

	if !fieldNamePattern.MatchString(field) {
		return nil, badRequest("Invalid request", "invalid field name %q", field)
	}
	if !op.Valid() {
		return nil, badRequest("Invalid request", "unknown operator %q for field %q", string(op), field)
	}

	f := &Filter{Field: field, Operator: op, Hidden: m[1] == "hfilter"}

	switch op {
	case OpExists:
		// value ignored
	case OpBetween:
		parts := strings.SplitN(vals[0], ",", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, badRequest("Invalid request", "between requires min,max for field %q", field)
		}
		f.Values = parts
	default:
		f.Values = vals
	}

	return f, nil
}

func applyScalar(opts *QueryOptions, key, val string) error {
	switch key {
	case "query":
		ensureQuery(opts).Value = val
	case "query.fields":
		for _, fld := range strings.Split(val, ",") {
			if !fieldNamePattern.MatchString(fld) {
				return badRequest("Invalid request", "invalid field name %q", fld)
			}
			ensureQuery(opts).Fields = append(ensureQuery(opts).Fields, fld)
		}
	case "query.operator":
		if val != "AND" && val != "OR" {
			return badRequest("Invalid request", "query.operator must be AND or OR")
		}
		ensureQuery(opts).Operator = val
	case "query.fuzzy":
		ensureQuery(opts).Fuzzy = val == "true"
	case "facetLimit":
		n, err := strconv.Atoi(val)
		if err != nil || n < 1 || n > MaxFacetLimit {
			return badRequest("Invalid request", "facetLimit must be 1..%d", MaxFacetLimit)
		}
		opts.FacetLimit = n
	case "facetSort":
		if val != "count" && val != "value" {
			return badRequest("Invalid request", "facetSort must be count or value")
		}
		opts.FacetSort = val
	case "facetBool":
		switch val {
		case "and":
			opts.FacetAnd = true
		case "or":
			opts.FacetAnd = false
		default:
			return badRequest("Invalid request", "facetBool must be and or or")
		}
	case "sort":
		field, desc := val, false
		if strings.HasPrefix(val, "-") {
			field, desc = val[1:], true
		}
		if !fieldNamePattern.MatchString(field) {
			return badRequest("Invalid request", "invalid sort field %q", field)
		}
		opts.Sort = &Sort{Field: field, Desc: desc}
	case "page":
		n, err := strconv.Atoi(val)
		if err != nil || n < 1 {
			return badRequest("Invalid request", "page must be a positive integer")
		}
		opts.Page = n
	case "size":
		n, err := strconv.Atoi(val)
		if err != nil || n < 0 || n > MaxSize {
			return badRequest("Invalid request", "size must be 0..%d", MaxSize)
		}
		opts.Size = n
	case "collapse":
		if !fieldNamePattern.MatchString(val) {
			return badRequest("Invalid request", "invalid collapse field %q", val)
		}
		ensureCollapse(opts).Field = val
	case "collapse.size":
		n, err := strconv.Atoi(val)
		if err != nil || n < 1 {
			return badRequest("Invalid request", "collapse.size must be a positive integer")
		}
		ensureCollapse(opts).Size = n
	case "collapse.sort":
		if !fieldNamePattern.MatchString(val) {
			return badRequest("Invalid request", "invalid collapse.sort field %q", val)
		}
		ensureCollapse(opts).Sort = val
	case "debug":
		opts.Debug = val == "true"
	}
	return nil
}

func ensureQuery(opts *QueryOptions) *TextQuery {
	if opts.Query == nil {
		opts.Query = &TextQuery{Operator: "AND"}
	}
	return opts.Query
}

func ensureCollapse(opts *QueryOptions) *Collapse {
	if opts.Collapse == nil {
		opts.Collapse = &Collapse{}
	}
	return opts.Collapse
}

// sortFilters gives Filters a deterministic order (url.Values is a map):
// by field, then operator, then hidden.
func sortFilters(filters []Filter) {
	sort.Slice(filters, func(i, j int) bool {
		a, b := filters[i], filters[j]
		if a.Field != b.Field {
			return a.Field < b.Field
		}
		if a.Operator != b.Operator {
			return a.Operator < b.Operator
		}
		return !a.Hidden && b.Hidden
	})
}
```

(imports also need `"regexp"` and `"sort"`; a `collapse` validation detail: if `collapse.size`/`collapse.sort` arrive without `collapse`, `ensureCollapse` leaves `Field` empty — add at the end of `ParseQuery`, before return: `if opts.Collapse != nil && opts.Collapse.Field == "" { return nil, badRequest("Invalid request", "collapse.size/collapse.sort require collapse") }`. Note the "full surface" test sets `Operator: "OR"` on TextQuery via query.operator; the round-trip test constructs `Operator: "AND"` explicitly to match `ensureQuery`.)

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./ikuzo/service/x/semanticv1/ -v`
Expected: PASS (all)

- [ ] **Step 6: Commit**

```bash
git add ikuzo/service/x/semanticv1/
git commit -m "feat(semanticv1): contract types, GET parser, canonical bijective encoding"
```

### Task 3: POST body parsing (same QueryOptions)

**Files:**
- Create: `ikuzo/service/x/semanticv1/parse_post.go`
- Test: `ikuzo/service/x/semanticv1/parse_post_test.go`

**Interfaces:**
- Consumes: `QueryOptions`, `Filter`, `Operator`, `badRequest`, `defaultOptions`, `fieldNamePattern`, `sortFilters` from Task 2.
- Produces: `func ParseSearchBody(r io.Reader) (*QueryOptions, error)`.

- [ ] **Step 1: Write the failing tests**

```go
package semanticv1

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestParseSearchBody(t *testing.T) {
	is := is.New(t)

	body := `{
	  "@context": "https://example.org/api/semantic/v1/contexts/hub3/1.0/context.jsonld",
	  "@type": "hub3:SearchQuery",
	  "query": {"value": "rembrandt", "fields": ["dc_title"], "operator": "OR", "fuzzy": true},
	  "filters": [
	    {"@type": "hub3:Filter", "field": "dc_creator", "operator": "eq", "values": ["Rembrandt"]},
	    {"@type": "hub3:Filter", "field": "dc_date", "operator": "between", "values": ["1600","1700"], "hidden": true}
	  ],
	  "facets": ["dc_creator"],
	  "facetLimit": 25,
	  "sort": {"field": "dc_date", "desc": true},
	  "page": 2,
	  "size": 50
	}`

	got, err := ParseSearchBody(strings.NewReader(body))
	is.NoErr(err)

	// POST parses to the same QueryOptions the equivalent GET produces
	getRaw := "query=rembrandt&query.fields=dc_title&query.operator=OR&query.fuzzy=true" +
		"&filter%5Bdc_creator%5D%5Beq%5D=Rembrandt&hfilter%5Bdc_date%5D%5Bbetween%5D=1600,1700" +
		"&facet=dc_creator&facetLimit=25&sort=-dc_date&page=2&size=50"
	vals, err := url.ParseQuery(getRaw)
	is.NoErr(err)
	want, err := ParseQuery(vals)
	is.NoErr(err)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("POST/GET divergence:\npost %+v\nget  %+v", got, want)
	}
}

func TestParseSearchBodyRejections(t *testing.T) {
	tests := []struct {
		name, body, wantErr string
	}{
		{"unknown top-level key", `{"detailLevel": "full"}`, "unknown field"},
		{"unknown filter type", `{"filters":[{"@type":"hub3:GeoFilter","field":"f","operator":"eq","values":["x"]}]}`, "unknown filter @type"},
		{"unknown operator", `{"filters":[{"field":"f","operator":"fuzzy","values":["x"]}]}`, "unknown operator"},
		{"invalid json", `{`, "invalid JSON"},
		{"size above max", `{"size": 500}`, "size"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			_, err := ParseSearchBody(strings.NewReader(tt.body))
			is.True(err != nil)
			is.True(strings.Contains(err.Error(), tt.wantErr))
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./ikuzo/service/x/semanticv1/ -run TestParseSearchBody -v`
Expected: FAIL (undefined: ParseSearchBody)

- [ ] **Step 3: Implement (`parse_post.go`)**

```go
package semanticv1

import (
	"bytes"
	"encoding/json"
	"io"
)

type searchBody struct {
	Context    json.RawMessage `json:"@context"`
	TypeValue  string          `json:"@type"`
	Query      *textQueryBody  `json:"query"`
	Filters    []filterBody    `json:"filters"`
	Facets     []string        `json:"facets"`
	FacetLimit *int            `json:"facetLimit"`
	FacetSort  string          `json:"facetSort"`
	FacetBool  string          `json:"facetBool"`
	Sort       *sortBody       `json:"sort"`
	Page       *int            `json:"page"`
	Size       *int            `json:"size"`
	Collapse   *collapseBody   `json:"collapse"`
	Debug      bool            `json:"debug"`
}

type textQueryBody struct {
	TypeValue string   `json:"@type"`
	Value     string   `json:"value"`
	Fields    []string `json:"fields"`
	Operator  string   `json:"operator"`
	Fuzzy     bool     `json:"fuzzy"`
}

type filterBody struct {
	TypeValue string   `json:"@type"`
	Field     string   `json:"field"`
	Operator  string   `json:"operator"`
	Values    []string `json:"values"`
	Hidden    bool     `json:"hidden"`
}

type sortBody struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

type collapseBody struct {
	Field string `json:"field"`
	Size  int    `json:"size"`
	Sort  string `json:"sort"`
}

// ParseSearchBody parses a POST /search JSON body into the same
// QueryOptions the GET surface produces. Unknown top-level fields and
// unknown filter @type values are rejected — this keeps future additions
// unambiguous. The envelope @context/@type are accepted, not processed.
func ParseSearchBody(r io.Reader) (*QueryOptions, error) {
	raw, err := io.ReadAll(io.LimitReader(r, 1<<20))
	if err != nil {
		return nil, badRequest("Invalid request", "unable to read request body")
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()

	var body searchBody
	if err := dec.Decode(&body); err != nil {
		if strings.Contains(err.Error(), "unknown field") {
			return nil, badRequest("Invalid request", "unknown field in request body: %v", err)
		}
		return nil, badRequest("Invalid request", "invalid JSON: %v", err)
	}

	opts := defaultOptions()

	if q := body.Query; q != nil {
		op := q.Operator
		if op == "" {
			op = "AND"
		}
		if op != "AND" && op != "OR" {
			return nil, badRequest("Invalid request", "query.operator must be AND or OR")
		}
		for _, fld := range q.Fields {
			if !fieldNamePattern.MatchString(fld) {
				return nil, badRequest("Invalid request", "invalid field name %q", fld)
			}
		}
		opts.Query = &TextQuery{Value: q.Value, Fields: q.Fields, Operator: op, Fuzzy: q.Fuzzy}
	}

	for _, fb := range body.Filters {
		if fb.TypeValue != "" && fb.TypeValue != "hub3:Filter" && fb.TypeValue != "Filter" {
			return nil, badRequest("Invalid request", "unknown filter @type %q", fb.TypeValue)
		}
		op := Operator(fb.Operator)
		if !op.Valid() {
			return nil, badRequest("Invalid request", "unknown operator %q for field %q", fb.Operator, fb.Field)
		}
		if !fieldNamePattern.MatchString(fb.Field) {
			return nil, badRequest("Invalid request", "invalid field name %q", fb.Field)
		}
		if op == OpBetween && len(fb.Values) != 2 {
			return nil, badRequest("Invalid request", "between requires exactly two values for field %q", fb.Field)
		}
		f := Filter{Field: fb.Field, Operator: op, Values: fb.Values, Hidden: fb.Hidden}
		if op == OpExists {
			f.Values = nil
		}
		opts.Filters = append(opts.Filters, f)
	}

	for _, f := range body.Facets {
		if !fieldNamePattern.MatchString(f) {
			return nil, badRequest("Invalid request", "invalid field name %q", f)
		}
		opts.Facets = append(opts.Facets, f)
	}

	if body.FacetLimit != nil {
		if *body.FacetLimit < 1 || *body.FacetLimit > MaxFacetLimit {
			return nil, badRequest("Invalid request", "facetLimit must be 1..%d", MaxFacetLimit)
		}
		opts.FacetLimit = *body.FacetLimit
	}
	if body.FacetSort != "" {
		if body.FacetSort != "count" && body.FacetSort != "value" {
			return nil, badRequest("Invalid request", "facetSort must be count or value")
		}
		opts.FacetSort = body.FacetSort
	}
	switch body.FacetBool {
	case "":
	case "and":
		opts.FacetAnd = true
	case "or":
	default:
		return nil, badRequest("Invalid request", "facetBool must be and or or")
	}

	if s := body.Sort; s != nil {
		if !fieldNamePattern.MatchString(s.Field) {
			return nil, badRequest("Invalid request", "invalid sort field %q", s.Field)
		}
		opts.Sort = &Sort{Field: s.Field, Desc: s.Desc}
	}

	if body.Page != nil {
		if *body.Page < 1 {
			return nil, badRequest("Invalid request", "page must be a positive integer")
		}
		opts.Page = *body.Page
	}
	if body.Size != nil {
		if *body.Size < 0 || *body.Size > MaxSize {
			return nil, badRequest("Invalid request", "size must be 0..%d", MaxSize)
		}
		opts.Size = *body.Size
	}

	if c := body.Collapse; c != nil {
		if !fieldNamePattern.MatchString(c.Field) {
			return nil, badRequest("Invalid request", "invalid collapse field %q", c.Field)
		}
		opts.Collapse = &Collapse{Field: c.Field, Size: c.Size, Sort: c.Sort}
	}

	opts.Debug = body.Debug

	sortFilters(opts.Filters)
	sort.Strings(opts.Facets)

	return opts, nil
}
```

(imports also need `"sort"` and `"strings"`)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ikuzo/service/x/semanticv1/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ikuzo/service/x/semanticv1/
git commit -m "feat(semanticv1): POST body parser, GET/POST parity by construction"
```

### Task 4: Store interface + result types

**Files:**
- Create: `ikuzo/service/x/semanticv1/store.go`
- Test: none (interface + plain structs; covered by Tasks 5–8)

**Interfaces:**
- Produces:

```go
package semanticv1

import (
	"context"
	"encoding/json"
	"errors"
)

// ErrNotFound is returned by GetByID when no document matches.
var ErrNotFound = errors.New("resource not found")

// SearchStore is the backend seam. The v2 adapter implements it today;
// a native ES or triple-store backend implements it later. Any
// implementation must pass the contract tests in contract_test.go.
type SearchStore interface {
	Search(ctx context.Context, orgID string, opts *QueryOptions) (*SearchResult, error)
	GetByID(ctx context.Context, orgID, id string) (json.RawMessage, error)
}

// SearchResult is backend-neutral. Items are the semantic item documents
// verbatim — the service wraps them, never edits them.
type SearchResult struct {
	Total  int64
	Items  []json.RawMessage
	Facets []Facet
	TookMS int64
	Debug  map[string]any // populated only when opts.Debug
}

type Facet struct {
	Field  string
	Total  int
	Values []FacetValue
}

type FacetValue struct {
	Value    string
	Count    int
	Selected bool
}
```

- [ ] **Step 1: Write the file exactly as above**
- [ ] **Step 2: Verify it compiles**

Run: `go build ./ikuzo/service/x/semanticv1/`
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add ikuzo/service/x/semanticv1/store.go
git commit -m "feat(semanticv1): SearchStore interface and result types"
```

### Task 5: v2 bridge — query translation (internal package)

**Files:**
- Create: `ikuzo/service/x/semanticv1/internal/v2bridge/translate.go`
- Test: `ikuzo/service/x/semanticv1/internal/v2bridge/translate_test.go`

**Interfaces:**
- Consumes: `semanticv1.QueryOptions`, `Filter`, operators, `TextQuery` (Task 2). Note: `internal/v2bridge` importing its parent `semanticv1` is legal and creates no cycle — `semanticv1` itself never imports the bridge directly except via the single re-export file added in Task 6.
- Produces: `func translateToV2(opts *semanticv1.QueryOptions) (url.Values, error)` in package `v2bridge` — emits only params that `fragments.NewSearchRequest` (v2's public param contract) parses.

**The complete mapping (this table IS the translator spec):**

| semanticv1 | v2 param | Encoding |
|---|---|---|
| Query.Value | `query` | terms joined; `~` suffix per term when Fuzzy; terms joined with ` AND ` when Operator=AND and multiple terms, ` OR ` when OR |
| Query.Fields | `searchFields` | comma-joined |
| Filter eq | `qf` | `field:value` (one per value) |
| Filter neq | `qf` | `-field:value` |
| Filter in | `qf` | `field:(v1 OR v2 ...)` |
| Filter nin | `qf` | `-field:(v1 OR v2 ...)` |
| Filter gt / gte / lt / lte | `qf` | `field:{v TO *}` / `field:[v TO *]` / `field:{* TO v}` / `field:[* TO v]` |
| Filter between | `qf` | `field:[min TO max]` |
| Filter contains | `qf` | `field:*value*` |
| Filter startswith | `qf` | `field:value*` |
| Filter exists | `qf` | `_exists_:field` |
| Filter Hidden=true | `hqf` | same encodings as above |
| Facets | `facet.field` | one per facet, verbatim (underscore form) |
| FacetLimit | `facet.limit` | single global value |
| FacetSort=value | `facet.sort` | `value` (omit for default count) |
| FacetAnd | `facetBoolType` | `and` (omit for or/default) |
| Sort | `sortBy` + `sortAsc` | field verbatim; `sortAsc=false` when Desc |
| Page/Size | `page`, `rows`, `start` | `start=(page-1)*size`, `rows=size` |
| Collapse | `collapseOn`, `collapseSize`, `collapseSort` | verbatim |
| always | `itemFormat` | `semantic` |

- [ ] **Step 1: Write the failing tests**

```go
package v2bridge

import (
	"testing"

	"github.com/matryer/is"

	semanticv1 "github.com/delving/hub3/ikuzo/service/x/semanticv1"
)

func TestTranslateToV2(t *testing.T) {
	tests := []struct {
		name string
		opts semanticv1.QueryOptions
		want map[string][]string
	}{
		{
			name: "namespaced field filter stays underscore",
			opts: semanticv1.QueryOptions{Page: 1, Size: 20,
				Filters: []semanticv1.Filter{{Field: "dc_creator", Operator: semanticv1.OpEq, Values: []string{"Rembrandt"}}}},
			want: map[string][]string{
				"qf": {"dc_creator:Rembrandt"}, "itemFormat": {"semantic"},
				"page": {"1"}, "rows": {"20"}, "start": {"0"},
			},
		},
		{
			name: "exists and between",
			opts: semanticv1.QueryOptions{Page: 1, Size: 20,
				Filters: []semanticv1.Filter{
					{Field: "nave_thumbnail", Operator: semanticv1.OpExists},
					{Field: "dc_date", Operator: semanticv1.OpBetween, Values: []string{"1600", "1700"}},
				}},
			want: map[string][]string{
				"qf": {"_exists_:nave_thumbnail", "dc_date:[1600 TO 1700]"}, "itemFormat": {"semantic"},
				"page": {"1"}, "rows": {"20"}, "start": {"0"},
			},
		},
		{
			name: "hidden filter goes to hqf",
			opts: semanticv1.QueryOptions{Page: 1, Size: 20,
				Filters: []semanticv1.Filter{{Field: "meta_tags", Operator: semanticv1.OpEq, Values: []string{"x"}, Hidden: true}}},
			want: map[string][]string{
				"hqf": {"meta_tags:x"}, "itemFormat": {"semantic"},
				"page": {"1"}, "rows": {"20"}, "start": {"0"},
			},
		},
		{
			name: "facets sort paging fuzzy",
			opts: semanticv1.QueryOptions{Page: 3, Size: 50,
				Query:  &semanticv1.TextQuery{Value: "night watch", Operator: "AND", Fuzzy: true},
				Facets: []string{"dc_creator", "edm_dataProvider"}, FacetLimit: 25, FacetAnd: true,
				Sort: &semanticv1.Sort{Field: "dc_date", Desc: true}},
			want: map[string][]string{
				"query": {"night~ AND watch~"}, "itemFormat": {"semantic"},
				"facet.field": {"dc_creator", "edm_dataProvider"}, "facet.limit": {"25"}, "facetBoolType": {"and"},
				"sortBy": {"dc_date"}, "sortAsc": {"false"},
				"page": {"3"}, "rows": {"50"}, "start": {"100"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			got, err := translateToV2(&tt.opts)
			is.NoErr(err)
			is.Equal(map[string][]string(got), tt.want)
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./ikuzo/service/x/semanticv1/internal/v2bridge/ -run TestTranslateToV2 -v`
Expected: FAIL (undefined: translateToV2)

- [ ] **Step 3: Implement the translator (`translate.go`)**

```go
// Package v2bridge is the RETIRABLE compatibility layer between the
// semanticv1 contract and v2 search. It enters v2 through its public
// param contract (url.Values -> fragments.NewSearchRequest) and leaves
// through v2's decoded ScrollResultV4 response. Nothing outside
// semanticv1 can import it. Delete this package (plus semanticv1/v2.go)
// when a native backend passes the contract suite.
package v2bridge

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	semanticv1 "github.com/delving/hub3/ikuzo/service/x/semanticv1"
)

func translateToV2(opts *semanticv1.QueryOptions) (url.Values, error) {
	p := url.Values{}
	p.Set("itemFormat", "semantic")

	if q := opts.Query; q != nil && q.Value != "" {
		terms := strings.Fields(q.Value)
		if q.Fuzzy {
			for i := range terms {
				terms[i] += "~"
			}
		}
		joiner := " AND "
		if q.Operator == "OR" {
			joiner = " OR "
		}
		p.Set("query", strings.Join(terms, joiner))
		if len(q.Fields) > 0 {
			p.Set("searchFields", strings.Join(q.Fields, ","))
		}
	}

	for _, f := range opts.Filters {
		qf, err := encodeQF(f)
		if err != nil {
			return nil, err
		}
		key := "qf"
		if f.Hidden {
			key = "hqf"
		}
		for _, clause := range qf {
			p.Add(key, clause)
		}
	}

	for _, f := range opts.Facets {
		p.Add("facet.field", f)
	}
	if len(opts.Facets) > 0 {
		p.Set("facet.limit", strconv.Itoa(opts.FacetLimit))
		if opts.FacetSort == "value" {
			p.Set("facet.sort", "value")
		}
	}
	if opts.FacetAnd {
		p.Set("facetBoolType", "and")
	}

	if s := opts.Sort; s != nil {
		p.Set("sortBy", s.Field)
		p.Set("sortAsc", strconv.FormatBool(!s.Desc))
	}

	if c := opts.Collapse; c != nil {
		p.Set("collapseOn", c.Field)
		if c.Size > 0 {
			p.Set("collapseSize", strconv.Itoa(c.Size))
		}
		if c.Sort != "" {
			p.Set("collapseSort", c.Sort)
		}
	}

	p.Set("page", strconv.Itoa(opts.Page))
	p.Set("rows", strconv.Itoa(opts.Size))
	p.Set("start", strconv.Itoa((opts.Page-1)*opts.Size))

	return p, nil
}

func encodeQF(f semanticv1.Filter) ([]string, error) {
	switch f.Operator {
	case semanticv1.OpEq:
		out := make([]string, len(f.Values))
		for i, v := range f.Values {
			out[i] = f.Field + ":" + v
		}
		return out, nil
	case semanticv1.OpNeq:
		out := make([]string, len(f.Values))
		for i, v := range f.Values {
			out[i] = "-" + f.Field + ":" + v
		}
		return out, nil
	case semanticv1.OpIn:
		return []string{f.Field + ":(" + strings.Join(f.Values, " OR ") + ")"}, nil
	case semanticv1.OpNin:
		return []string{"-" + f.Field + ":(" + strings.Join(f.Values, " OR ") + ")"}, nil
	case semanticv1.OpGt:
		return []string{f.Field + ":{" + f.Values[0] + " TO *}"}, nil
	case semanticv1.OpGte:
		return []string{f.Field + ":[" + f.Values[0] + " TO *]"}, nil
	case semanticv1.OpLt:
		return []string{f.Field + ":{* TO " + f.Values[0] + "}"}, nil
	case semanticv1.OpLte:
		return []string{f.Field + ":[* TO " + f.Values[0] + "]"}, nil
	case semanticv1.OpBetween:
		return []string{f.Field + ":[" + f.Values[0] + " TO " + f.Values[1] + "]"}, nil
	case semanticv1.OpContains:
		return []string{f.Field + ":*" + f.Values[0] + "*"}, nil
	case semanticv1.OpStartsWith:
		return []string{f.Field + ":" + f.Values[0] + "*"}, nil
	case semanticv1.OpExists:
		return []string{"_exists_:" + f.Field}, nil
	default:
		return nil, fmt.Errorf("unsupported operator: %s", f.Operator)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ikuzo/service/x/semanticv1/internal/v2bridge/ -run TestTranslateToV2 -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ikuzo/service/x/semanticv1/internal/v2bridge/
git commit -m "feat(semanticv1): internal v2bridge query translation, underscore fields verbatim"
```

### Task 6: v2 bridge — execution, ScrollResultV4 mapping, re-export

**Files:**
- Create: `ikuzo/service/x/semanticv1/internal/v2bridge/store.go` (store struct, Search, GetByID)
- Create: `ikuzo/service/x/semanticv1/internal/v2bridge/response.go` (`fromScrollResult`)
- Create: `ikuzo/service/x/semanticv1/v2.go` (the ONE re-export — wiring entry point and the bridge's only doorway)
- Test: `ikuzo/service/x/semanticv1/internal/v2bridge/response_test.go` (+ integration test gated on live ES, pattern of `v2adapter/adapter_integration_test.go`)

**Interfaces:**
- Consumes: `translateToV2` (Task 5); v2's own machinery: `fragments.NewSearchRequest` + `ExecuteWithParallelAggregations` for execution, and v2's response decode producing `*fragments.ScrollResultV4` — port `decodeV2Results` from the old `ikuzo/storage/x/v2adapter/adapter.go` verbatim (it calls v2's own hit/facet decoding); port `GetByID` from `adapter.go:143-238` (`_id` get for hubIDs, `meta.entryURI` term query for URIs, index `{orgID}v2`).
- Produces:
  - `v2bridge.NewStore(client *elastic.Client, log zerolog.Logger) *Store` implementing `semanticv1.SearchStore`
  - `fromScrollResult(sr *fragments.ScrollResultV4, opts *semanticv1.QueryOptions) (*semanticv1.SearchResult, error)` — the pure response-side mapping
  - `semanticv1.NewV2SearchStore(client *elastic.Client, log zerolog.Logger) SearchStore` (in `v2.go`) — the only symbol the wiring layer sees

- [ ] **Step 1: Write the failing mapping test** — pure function over a canned `ScrollResultV4` JSON fixture. Create `testdata/scroll_result_v4.json` by capturing a real v2 response (`/api/search/v2?query=*&rows=2&itemFormat=semantic&facet.field=dc_creator` against the dev stack, or serialize one in a bootstrap test using the old adapter):

```go
package v2bridge

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/matryer/is"

	"github.com/delving/hub3/hub3/fragments"
	semanticv1 "github.com/delving/hub3/ikuzo/service/x/semanticv1"
)

func TestFromScrollResult(t *testing.T) {
	is := is.New(t)

	raw, err := os.ReadFile("testdata/scroll_result_v4.json")
	is.NoErr(err)

	var sr fragments.ScrollResultV4
	is.NoErr(json.Unmarshal(raw, &sr))

	opts := &semanticv1.QueryOptions{Page: 1, Size: 20,
		Filters: []semanticv1.Filter{{Field: "dc_creator", Operator: semanticv1.OpEq, Values: []string{"Rembrandt"}}}}

	res, err := fromScrollResult(&sr, opts)
	is.NoErr(err)
	is.True(res.Total > 0)
	is.True(len(res.Items) > 0)

	// member items are the semantic view verbatim
	var first map[string]any
	is.NoErr(json.Unmarshal(res.Items[0], &first))
	_, hasID := first["@id"]
	is.True(hasID)

	// facets carry selected-ness derived from opts.Filters
	for _, f := range res.Facets {
		if f.Field != "dc_creator" {
			continue
		}
		for _, v := range f.Values {
			if v.Value == "Rembrandt" {
				is.True(v.Selected)
			}
		}
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./ikuzo/service/x/semanticv1/internal/v2bridge/ -run TestFromScrollResult -v`
Expected: FAIL (undefined: fromScrollResult / missing fixture — create the fixture in this step)

- [ ] **Step 3: Implement `response.go`, `store.go`, and the re-export.**

`response.go` — the pure mapping. `ScrollResultV4` items are `*fragments.FragmentGraph`; with `itemFormat=semantic` the item's `Semantic` field holds the member document. Facets come as `[]*fragments.QueryFacet` with `Links []*fragments.FacetLink` carrying value/count/selected:

```go
package v2bridge

import (
	"encoding/json"
	"fmt"

	"github.com/delving/hub3/hub3/fragments"
	semanticv1 "github.com/delving/hub3/ikuzo/service/x/semanticv1"
)

// fromScrollResult maps v2's decoded response onto the semanticv1 result.
// It is intentionally dumb: v2 has already done all hit and facet decoding;
// this only reshapes. Verify exact field names against
// fragments.ScrollResultV4 / QueryFacet / FacetLink at implementation time.
func fromScrollResult(sr *fragments.ScrollResultV4, opts *semanticv1.QueryOptions) (*semanticv1.SearchResult, error) {
	res := &semanticv1.SearchResult{}
	if p := sr.GetPager(); p != nil {
		res.Total = int64(p.GetNrOfResults())
	}

	for _, item := range sr.GetItems() {
		if item.GetSemantic() == nil {
			continue
		}
		raw, err := json.Marshal(item.GetSemantic())
		if err != nil {
			return nil, fmt.Errorf("marshal semantic item: %w", err)
		}
		res.Items = append(res.Items, raw)
	}

	for _, qf := range sr.GetFacets() {
		facet := semanticv1.Facet{Field: qf.GetField(), Total: int(qf.GetTotal())}
		for _, link := range qf.GetLinks() {
			facet.Values = append(facet.Values, semanticv1.FacetValue{
				Value:    link.GetValue(),
				Count:    int(link.GetCount()),
				Selected: link.GetIsSelected(),
			})
		}
		res.Facets = append(res.Facets, facet)
	}

	return res, nil
}
```

`store.go`:

```go
package v2bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	elastic "github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"

	"github.com/delving/hub3/hub3/fragments"
	semanticv1 "github.com/delving/hub3/ikuzo/service/x/semanticv1"
)

type Store struct {
	client *elastic.Client
	log    zerolog.Logger
}

func NewStore(client *elastic.Client, log zerolog.Logger) *Store {
	return &Store{client: client, log: log.With().Str("component", "semanticv1_v2bridge").Logger()}
}

func (s *Store) Search(ctx context.Context, orgID string, opts *semanticv1.QueryOptions) (*semanticv1.SearchResult, error) {
	params, err := translateToV2(opts)
	if err != nil {
		return nil, err
	}

	sr, err := fragments.NewSearchRequest(orgID, params)
	if err != nil {
		return nil, fmt.Errorf("v2 search request: %w", err)
	}

	start := time.Now()
	esRes, err := sr.ExecuteWithParallelAggregations(s.client, ctx)
	if err != nil {
		return nil, fmt.Errorf("v2 search execution: %w", err)
	}

	// decodeV2Results: verbatim port from the old
	// ikuzo/storage/x/v2adapter/adapter.go — produces *fragments.ScrollResultV4
	// using v2's own hit and facet decoding.
	scroll, err := decodeV2Results(esRes, sr)
	if err != nil {
		return nil, fmt.Errorf("v2 response decode: %w", err)
	}

	res, err := fromScrollResult(scroll, opts)
	if err != nil {
		return nil, err
	}
	res.TookMS = time.Since(start).Milliseconds()

	if opts.Debug {
		res.Debug = map[string]any{"v2Params": map[string][]string(params)}
	}
	return res, nil
}

func (s *Store) GetByID(ctx context.Context, orgID, id string) (json.RawMessage, error) {
	// verbatim port of the old adapter.go GetByID (adapter.go:143-238):
	// _id get for hubIDs, meta.entryURI term query for http(s) URIs, index
	// name {orgID}v2; extract the "semantic" key verbatim;
	// elastic.IsNotFound -> fmt.Errorf("%w: %s", semanticv1.ErrNotFound, id)
}
```

`ikuzo/service/x/semanticv1/v2.go` — the one doorway:

```go
package semanticv1

import (
	elastic "github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"

	"github.com/delving/hub3/ikuzo/service/x/semanticv1/internal/v2bridge"
)

// NewV2SearchStore returns the SearchStore backed by v2 search via the
// internal, retirable compatibility bridge. Delete this file together
// with internal/v2bridge when a native backend passes the contract suite.
func NewV2SearchStore(client *elastic.Client, log zerolog.Logger) SearchStore {
	return v2bridge.NewStore(client, log)
}
```

- [ ] **Step 4: Run package tests**

Run: `go test ./ikuzo/service/x/semanticv1/... -v 2>&1 | grep -E "^(--- |ok|FAIL)"`
Expected: PASS (integration tests skip without ES; run with local ES for the full check)

- [ ] **Step 5: Commit**

```bash
git add ikuzo/service/x/semanticv1/
git commit -m "feat(semanticv1): v2bridge execution via ScrollResultV4, single re-export doorway"
```

### Task 7: Hydra envelope builder

**Files:**
- Create: `ikuzo/service/x/semanticv1/envelope.go`
- Test: `ikuzo/service/x/semanticv1/envelope_test.go`

**Interfaces:**
- Consumes: `QueryOptions`, `EncodeQuery`, `SearchResult`, `Facet`, `ContextURL` from earlier tasks.
- Produces: `func BuildCollection(baseURL string, opts *QueryOptions, res *SearchResult) map[string]any` and `func BuildError(baseURL string, cerr *ContractError) map[string]any`.

- [ ] **Step 1: Write the failing tests**

```go
package semanticv1

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/matryer/is"
)

const testBase = "https://example.org/api/semantic/v1"

func TestBuildCollection(t *testing.T) {
	is := is.New(t)

	opts := defaultOptions()
	opts.Query = &TextQuery{Value: "rembrandt", Operator: "AND"}
	opts.Filters = []Filter{{Field: "dc_creator", Operator: OpEq, Values: []string{"Rembrandt"}}}
	opts.Facets = []string{"dc_creator"}
	opts.Page = 2

	res := &SearchResult{
		Total: 100,
		Items: []json.RawMessage{json.RawMessage(`{"@id":"urn:x:1","dc_title":"The Night Watch"}`)},
		Facets: []Facet{{Field: "dc_creator", Total: 2, Values: []FacetValue{
			{Value: "Rembrandt", Count: 41, Selected: true},
			{Value: "Vermeer", Count: 12},
		}}},
		TookMS: 12,
	}

	doc := BuildCollection(testBase, opts, res)

	is.Equal(doc["@context"], testBase+"/contexts/hub3/1.0/context.jsonld") // versioned context URL, not inline map
	is.Equal(doc["hydra:totalItems"], int64(100))

	view := doc["hydra:view"].(map[string]any)
	next := view["hydra:next"].(string)
	prev := view["hydra:previous"].(string)
	// pagination links carry the FULL canonical query — works for POST too
	is.True(strings.Contains(next, "page=3"))
	is.True(strings.Contains(next, "query=rembrandt"))
	is.True(strings.Contains(next, "filter%5Bdc_creator%5D%5Beq%5D=Rembrandt"))
	is.True(strings.Contains(prev, "query=rembrandt")) // page=1 omitted (default), query kept

	// member content is byte-identical passthrough
	members := doc["hydra:member"].([]json.RawMessage)
	is.Equal(string(members[0]), `{"@id":"urn:x:1","dc_title":"The Night Watch"}`)

	// facet apply/remove URLs derive from canonical encoding
	facets := doc["hub3:facets"].([]map[string]any)
	vals := facets[0]["hub3:values"].([]map[string]any)
	is.True(vals[0]["hub3:selected"].(bool))
	is.True(strings.Contains(vals[0]["hub3:removeURL"].(string), "query=rembrandt"))
	is.True(!strings.Contains(vals[0]["hub3:removeURL"].(string), "dc_creator")) // filter removed
	is.True(strings.Contains(vals[1]["hub3:applyURL"].(string), "filter%5Bdc_creator%5D%5Beq%5D=Vermeer"))

	// whole document must marshal
	_, err := json.Marshal(doc)
	is.NoErr(err)

	// no debug block unless requested
	_, hasDebug := doc["hub3:debug"]
	is.True(!hasDebug)
}

func TestBuildCollectionLastPageHasNoNext(t *testing.T) {
	is := is.New(t)
	opts := defaultOptions() // page 1, size 20
	doc := BuildCollection(testBase, opts, &SearchResult{Total: 15})
	view := doc["hydra:view"].(map[string]any)
	_, hasNext := view["hydra:next"]
	is.True(!hasNext)
	_, hasPrev := view["hydra:previous"]
	is.True(!hasPrev)
}

func TestBuildError(t *testing.T) {
	is := is.New(t)
	doc := BuildError(testBase, badRequest("Invalid request", "unknown parameter: %s", "detailLevel"))
	is.Equal(doc["@type"], "hydra:Error")
	is.Equal(doc["hydra:statusCode"], 400)
	is.True(strings.Contains(doc["hydra:description"].(string), "detailLevel"))
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./ikuzo/service/x/semanticv1/ -run "TestBuildCollection|TestBuildError" -v`
Expected: FAIL (undefined: BuildCollection)

- [ ] **Step 3: Implement (`envelope.go`)**

```go
package semanticv1

import (
	"encoding/json"
	"math"
)

// searchURL renders the canonical URL for opts at a given page.
func searchURL(baseURL string, opts *QueryOptions, page int) string {
	clone := *opts
	clone.Page = page
	q := EncodeQuery(&clone).Encode()
	u := baseURL + "/search"
	if q != "" {
		u += "?" + q
	}
	return u
}

// BuildCollection wraps a SearchResult in the Hydra collection envelope.
// All links derive from the canonical encoding of opts, so GET and POST
// requests produce identical, followable pagination and facet URLs.
func BuildCollection(baseURL string, opts *QueryOptions, res *SearchResult) map[string]any {
	doc := map[string]any{
		"@context":         ContextURL(baseURL),
		"@id":              searchURL(baseURL, opts, opts.Page),
		"@type":            []string{"hydra:Collection", "schema:SearchResultsPage"},
		"hydra:totalItems": res.Total,
		"hydra:member":     memberList(res.Items),
	}

	totalPages := 1
	if opts.Size > 0 {
		totalPages = int(math.Ceil(float64(res.Total) / float64(opts.Size)))
	}

	view := map[string]any{
		"@id":         searchURL(baseURL, opts, opts.Page),
		"@type":       "hydra:PartialCollectionView",
		"hydra:first": searchURL(baseURL, opts, 1),
	}
	if opts.Page > 1 {
		view["hydra:previous"] = searchURL(baseURL, opts, opts.Page-1)
	}
	if opts.Page < totalPages {
		view["hydra:next"] = searchURL(baseURL, opts, opts.Page+1)
	}
	doc["hydra:view"] = view

	if len(res.Facets) > 0 {
		doc["hub3:facets"] = buildFacetBlocks(baseURL, opts, res.Facets)
	}
	if active := buildActiveFilters(baseURL, opts); len(active) > 0 {
		doc["hub3:activeFilters"] = active
	}

	doc["hub3:timing"] = map[string]any{"hub3:took": res.TookMS, "hub3:unit": "ms"}

	if res.Debug != nil {
		dbg := map[string]any{}
		for k, v := range res.Debug {
			dbg["hub3:"+k] = v
		}
		doc["hub3:debug"] = dbg
	}

	return doc
}

func memberList(items []json.RawMessage) []json.RawMessage {
	if items == nil {
		return []json.RawMessage{}
	}
	return items
}

func buildFacetBlocks(baseURL string, opts *QueryOptions, facets []Facet) []map[string]any {
	out := make([]map[string]any, 0, len(facets))
	for _, f := range facets {
		vals := make([]map[string]any, 0, len(f.Values))
		for _, v := range f.Values {
			entry := map[string]any{
				"hub3:value":    v.Value,
				"hub3:count":    v.Count,
				"hub3:selected": v.Selected,
			}
			if v.Selected {
				entry["hub3:removeURL"] = searchURL(baseURL, withoutFilter(opts, f.Field, v.Value), 1)
			} else {
				entry["hub3:applyURL"] = searchURL(baseURL, withFilter(opts, f.Field, v.Value), 1)
			}
			vals = append(vals, entry)
		}
		out = append(out, map[string]any{
			"hub3:field":  f.Field,
			"hub3:values": vals,
		})
	}
	return out
}

func buildActiveFilters(baseURL string, opts *QueryOptions) []map[string]any {
	out := []map[string]any{}
	for _, f := range opts.Filters {
		if f.Hidden {
			continue
		}
		out = append(out, map[string]any{
			"hub3:field":     f.Field,
			"hub3:operator":  string(f.Operator),
			"hub3:values":    f.Values,
			"hub3:removeURL": searchURL(baseURL, withoutFilterExact(opts, f), 1),
		})
	}
	return out
}

// withFilter returns a copy of opts with an eq filter added.
func withFilter(opts *QueryOptions, field, value string) *QueryOptions {
	clone := *opts
	clone.Filters = append(append([]Filter{}, opts.Filters...),
		Filter{Field: field, Operator: OpEq, Values: []string{value}})
	sortFilters(clone.Filters)
	return &clone
}

// withoutFilter returns a copy of opts with any visible filter on field
// matching value removed.
func withoutFilter(opts *QueryOptions, field, value string) *QueryOptions {
	clone := *opts
	clone.Filters = nil
	for _, f := range opts.Filters {
		if !f.Hidden && f.Field == field && len(f.Values) == 1 && f.Values[0] == value {
			continue
		}
		clone.Filters = append(clone.Filters, f)
	}
	return &clone
}

func withoutFilterExact(opts *QueryOptions, target Filter) *QueryOptions {
	clone := *opts
	clone.Filters = nil
	for _, f := range opts.Filters {
		if f.Field == target.Field && f.Operator == target.Operator && !f.Hidden {
			continue
		}
		clone.Filters = append(clone.Filters, f)
	}
	return &clone
}

// BuildError renders a ContractError as a hydra:Error document.
func BuildError(baseURL string, cerr *ContractError) map[string]any {
	return map[string]any{
		"@context":          ContextURL(baseURL),
		"@type":             "hydra:Error",
		"hydra:title":       cerr.Title,
		"hydra:description": cerr.Description,
		"hydra:statusCode":  cerr.Status,
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ikuzo/service/x/semanticv1/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ikuzo/service/x/semanticv1/
git commit -m "feat(semanticv1): Hydra envelope with canonical links for GET and POST"
```

### Task 8: HTTP handlers + contract test suite

**Files:**
- Create: `ikuzo/service/x/semanticv1/service.go` (Service struct, options, routes, handlers)
- Create: `ikuzo/service/x/semanticv1/contract_test.go`

**Interfaces:**
- Consumes: everything above; `domain.GetOrganizationFromCtx` (same org-middleware contract as the old service, see old `service.go` handleSearch); chi router (`github.com/go-chi/chi/v5`).
- Produces:
  - `func NewService(opts ...Option) (*Service, error)`
  - `func WithSearchStore(s SearchStore) Option`, `func WithBaseURL(u string) Option` (default `/api/semantic/v1`)
  - `func (s *Service) Routes(pattern string, r chi.Router)` mounting: `GET /`, `GET /docs`, `GET+POST /search`, `GET /resource/{id}`, `GET {ContextURLPath}`
  - implements `domain.Service` the same way the old service does (check `ikuzo/service/x/semantic/service.go` for the interface methods: `Metrics()`, `Shutdown(ctx)` — mirror them).

- [ ] **Step 1: Write the contract test suite (this is the API's acceptance contract — D9)**

```go
package semanticv1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/matryer/is"

	"github.com/delving/hub3/ikuzo/domain"
	semanticv1 "github.com/delving/hub3/ikuzo/service/x/semanticv1"
)

// fakeStore records the QueryOptions it receives and returns canned results.
type fakeStore struct {
	lastOpts *semanticv1.QueryOptions
	result   *semanticv1.SearchResult
	item     json.RawMessage
}

func (f *fakeStore) Search(_ context.Context, _ string, opts *semanticv1.QueryOptions) (*semanticv1.SearchResult, error) {
	f.lastOpts = opts
	return f.result, nil
}

func (f *fakeStore) GetByID(_ context.Context, _ string, id string) (json.RawMessage, error) {
	if f.item == nil {
		return nil, semanticv1.ErrNotFound
	}
	return f.item, nil
}

func newTestServer(t *testing.T, store *semanticv1.SearchStore) (*httptest.Server, *fakeStore) {
	t.Helper()
	fs := &fakeStore{result: &semanticv1.SearchResult{Total: 1,
		Items: []json.RawMessage{json.RawMessage(`{"@id":"urn:x:1"}`)}}}

	svc, err := semanticv1.NewService(semanticv1.WithSearchStore(fs))
	if err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { // org middleware stand-in
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			org := domain.Organization{ID: "testorg"}
			next.ServeHTTP(w, req.WithContext(domain.SetOrganizationInContext(req.Context(), org)))
		})
	})
	svc.Routes("", r)
	return httptest.NewServer(r), fs
}

func TestContractSearchGET(t *testing.T) {
	is := is.New(t)
	srv, fs := newTestServer(t, nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/semantic/v1/search?query=rembrandt&filter%5Bdc_creator%5D%5Beq%5D=Rembrandt")
	is.NoErr(err)
	is.Equal(resp.StatusCode, 200)
	is.Equal(resp.Header.Get("Content-Type"), "application/ld+json")

	is.Equal(fs.lastOpts.Query.Value, "rembrandt")
	is.Equal(fs.lastOpts.Filters[0].Field, "dc_creator")

	var doc map[string]any
	is.NoErr(json.NewDecoder(resp.Body).Decode(&doc))
	is.True(strings.HasSuffix(doc["@context"].(string), "/contexts/hub3/1.0/context.jsonld"))
}

func TestContractSearchPOSTEqualsGET(t *testing.T) {
	is := is.New(t)
	srv, fs := newTestServer(t, nil)
	defer srv.Close()

	_, err := http.Get(srv.URL + "/api/semantic/v1/search?query=x&filter%5Bdc_creator%5D%5Beq%5D=R&page=2")
	is.NoErr(err)
	getOpts := fs.lastOpts

	body := `{"query":{"value":"x"},"filters":[{"field":"dc_creator","operator":"eq","values":["R"]}],"page":2}`
	resp, err := http.Post(srv.URL+"/api/semantic/v1/search", "application/ld+json", strings.NewReader(body))
	is.NoErr(err)
	is.Equal(resp.StatusCode, 200)

	is.Equal(fs.lastOpts, getOpts) // identical QueryOptions

	// POST response pagination links are followable GET URLs carrying the query
	var doc map[string]any
	is.NoErr(json.NewDecoder(resp.Body).Decode(&doc))
	view := doc["hydra:view"].(map[string]any)
	is.True(strings.Contains(view["hydra:next"].(string), "query=x"))
}

func TestContractUnknownParamRejected(t *testing.T) {
	is := is.New(t)
	srv, _ := newTestServer(t, nil)
	defer srv.Close()

	for _, dead := range []string{"detailLevel=full", "cursor=abc", "contextIndex=x", "languages=nl",
		"expand=a", "fields=dc_title", "peek=dc_creator", "backend=es8"} {
		resp, err := http.Get(srv.URL + "/api/semantic/v1/search?" + dead)
		is.NoErr(err)
		is.Equal(resp.StatusCode, 400) // dead param must be rejected: dead

		var doc map[string]any
		is.NoErr(json.NewDecoder(resp.Body).Decode(&doc))
		is.Equal(doc["@type"], "hydra:Error")
	}
}

func TestContractResourceDetailPassthrough(t *testing.T) {
	is := is.New(t)
	srv, fs := newTestServer(t, nil)
	defer srv.Close()
	fs.item = json.RawMessage(`{"@id":"urn:x:1","dc_title":"t"}`)

	resp, err := http.Get(srv.URL + "/api/semantic/v1/resource/testorg_ds_1")
	is.NoErr(err)
	is.Equal(resp.StatusCode, 200)

	var doc map[string]any
	is.NoErr(json.NewDecoder(resp.Body).Decode(&doc))
	member := doc["hydra:member"].([]any)[0].(map[string]any)
	is.Equal(member["@id"], "urn:x:1")
	_, mutated := member["hub3:navigation"]
	is.True(!mutated) // content is never mutated
}

func TestContractNotFound(t *testing.T) {
	is := is.New(t)
	srv, _ := newTestServer(t, nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/semantic/v1/resource/nope")
	is.NoErr(err)
	is.Equal(resp.StatusCode, 404)
}

func TestContractContextServed(t *testing.T) {
	is := is.New(t)
	srv, _ := newTestServer(t, nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/semantic/v1/contexts/hub3/1.0/context.jsonld")
	is.NoErr(err)
	is.Equal(resp.StatusCode, 200)
	is.Equal(resp.Header.Get("Content-Type"), "application/ld+json")
}
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./ikuzo/service/x/semanticv1/ -run TestContract -v`
Expected: FAIL (undefined: NewService)

- [ ] **Step 3: Implement `service.go`.** Handler skeleton:

```go
package semanticv1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/delving/hub3/ikuzo/domain"
)

type Service struct {
	store   SearchStore
	baseURL string
}

type Option func(*Service) error

func WithSearchStore(s SearchStore) Option {
	return func(svc *Service) error { svc.store = s; return nil }
}

func WithBaseURL(u string) Option {
	return func(svc *Service) error { svc.baseURL = strings.TrimSuffix(u, "/"); return nil }
}

func NewService(opts ...Option) (*Service, error) {
	svc := &Service{baseURL: "/api/semantic/v1"}
	for _, o := range opts {
		if err := o(svc); err != nil {
			return nil, err
		}
	}
	if svc.store == nil {
		return nil, fmt.Errorf("semanticv1: WithSearchStore is required")
	}
	return svc, nil
}

func (s *Service) Routes(_ string, r chi.Router) {
	r.Route(s.baseURL, func(r chi.Router) {
		r.Get("/", s.handleEntryPoint)
		r.Get("/docs", s.handleDocs)
		r.Get("/search", s.handleSearchGET)
		r.Post("/search", s.handleSearchPOST)
		r.Get("/resource/{id}", s.handleResource)
		r.Get(ContextURLPath, s.handleContext)
	})
}

// absBase derives the absolute API base from the request (honouring
// X-Forwarded-Proto — old code ignored it and emitted http:// behind TLS).
func (s *Service) absBase(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host + s.baseURL
}

func (s *Service) handleSearchGET(w http.ResponseWriter, r *http.Request) {
	opts, err := ParseQuery(r.URL.Query())
	s.executeSearch(w, r, opts, err)
}

func (s *Service) handleSearchPOST(w http.ResponseWriter, r *http.Request) {
	opts, err := ParseSearchBody(r.Body)
	s.executeSearch(w, r, opts, err)
}

func (s *Service) executeSearch(w http.ResponseWriter, r *http.Request, opts *QueryOptions, err error) {
	base := s.absBase(r)
	if err != nil {
		s.writeError(w, base, err)
		return
	}

	org, ok := domain.GetOrganizationFromCtx(r.Context())
	if !ok || org.ID == "" {
		s.writeError(w, base, &ContractError{Status: 500, Title: "Server error", Description: "organization not resolved"})
		return
	}

	res, err := s.store.Search(r.Context(), org.ID.String(), opts)
	if err != nil {
		s.writeError(w, base, err)
		return
	}

	s.writeJSON(w, 200, BuildCollection(base, opts, res))
}

func (s *Service) handleResource(w http.ResponseWriter, r *http.Request) {
	base := s.absBase(r)

	id, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil || id == "" {
		s.writeError(w, base, badRequest("Invalid request", "missing or malformed resource id"))
		return
	}

	org, ok := domain.GetOrganizationFromCtx(r.Context())
	if !ok || org.ID == "" {
		s.writeError(w, base, &ContractError{Status: 500, Title: "Server error", Description: "organization not resolved"})
		return
	}

	item, err := s.store.GetByID(r.Context(), org.ID.String(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.writeError(w, base, &ContractError{Status: 404, Title: "Not found", Description: id})
			return
		}
		s.writeError(w, base, err)
		return
	}

	// single-member collection: envelope beside content, never inside it
	s.writeJSON(w, 200, map[string]any{
		"@context":         ContextURL(base),
		"@id":              base + "/resource/" + url.PathEscape(id),
		"@type":            "hydra:Collection",
		"hydra:totalItems": 1,
		"hydra:member":     []json.RawMessage{item},
	})
}

func (s *Service) handleContext(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/ld+json")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(ContextJSONLD)
}

func (s *Service) writeJSON(w http.ResponseWriter, status int, doc map[string]any) {
	w.Header().Set("Content-Type", "application/ld+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(doc)
}

func (s *Service) writeError(w http.ResponseWriter, base string, err error) {
	var cerr *ContractError
	if !errors.As(err, &cerr) {
		cerr = &ContractError{Status: 500, Title: "Server error", Description: err.Error()}
	}
	s.writeJSON(w, cerr.Status, BuildError(base, cerr))
}
```

(imports include `"net/url"`. `handleEntryPoint`/`handleDocs`: Task 9. For this task register them as minimal JSON documents so routes exist: entrypoint returns `{"@context": ctx, "@id": base, "@type": "hydra:EntryPoint", "hub3:search": base+"/search"}`; docs likewise minimal — completed next task. Mirror the old service's `Metrics()`/`Shutdown()` methods for `domain.Service` conformance.)

- [ ] **Step 4: Run the full package**

Run: `go test ./ikuzo/service/x/semanticv1/ -v 2>&1 | grep -E "^(--- |ok|FAIL)"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ikuzo/service/x/semanticv1/
git commit -m "feat(semanticv1): HTTP service, handlers, and the contract test suite"
```

### Task 9: EntryPoint + ApiDocumentation

**Files:**
- Create: `ikuzo/service/x/semanticv1/docs.go` (replaces the minimal stubs from Task 8)
- Test: `ikuzo/service/x/semanticv1/docs_test.go`

**Interfaces:**
- Consumes: `ContextURL`, the closed parameter list (mirror the `scalarParams` map + filter/facet syntax from Task 2).
- Produces: `handleEntryPoint`, `handleDocs` emitting `hydra:EntryPoint` and `hydra:ApiDocumentation` documents that enumerate ONLY the shipped surface: the 12 operators, the closed GET param list, the POST body shape. Source the operator list from `allOperators` and the param list from `scalarParams` so docs cannot drift from the parser.

- [ ] **Step 1: Write the failing test**

```go
func TestDocsMatchParser(t *testing.T) {
	is := is.New(t)
	srv, _ := newTestServer(t, nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/semantic/v1/docs")
	is.NoErr(err)
	is.Equal(resp.StatusCode, 200)

	raw, err := io.ReadAll(resp.Body)
	is.NoErr(err)
	body := string(raw)

	// every documented scalar param must be parseable — and vice versa
	for _, param := range []string{"query", "facetLimit", "facetSort", "facetBool", "sort", "page", "size", "collapse", "debug"} {
		is.True(strings.Contains(body, `"`+param+`"`)) // param documented: param
	}
	// dead params of the old API must NOT be documented
	for _, dead := range []string{"detailLevel", "cursor", "contextIndex", "languages", "peek", "backend"} {
		is.True(!strings.Contains(body, `"`+dead+`"`)) // dead param leaked into docs: dead
	}
	// all operators documented
	for _, op := range []string{"eq", "neq", "in", "nin", "between", "exists", "contains", "startswith"} {
		is.True(strings.Contains(body, `"`+op+`"`))
	}
}
```

- [ ] **Step 2: Run to verify failure** — `go test ./ikuzo/service/x/semanticv1/ -run TestDocsMatchParser -v` → FAIL
- [ ] **Step 3: Implement `docs.go`** — build the ApiDocumentation map by ranging over `scalarParams` and `allOperators` (single source of truth with the parser); document `filter[{field}][{op}]`, `hfilter[{field}][{op}]`, `facet` syntax and the POST body shape as `hub3:SearchQuery` supported properties.
- [ ] **Step 4: Run full package** — PASS
- [ ] **Step 5: Commit**

```bash
git add ikuzo/service/x/semanticv1/
git commit -m "feat(semanticv1): entrypoint and self-consistent API documentation"
```

### Task 10: Wiring + e2e verification

**Files:**
- Modify: `ikuzo/ikuzoctl/cmd/config/semantic.go` — replace the old service construction with `semanticv1.NewService(semanticv1.WithSearchStore(semanticv1.NewV2SearchStore(esClient.SearchClient(), logger)))`. The wiring layer never sees the bridge — only the re-export. Keep the config struct fields; `UseES8Backend` keeps its warn-and-ignore behavior.
- Create: `test_semanticv1_e2e.sh` — port `test_semantic_e2e.sh` sections 1–13 to the new surface (drop backend/context-navigation/introspect sections; add checks that each dead param returns 400 with `hydra:Error`; add a namespaced-field filter check `filter[dc_creator][eq]=...` asserting non-empty results against seeded data — the check the old suite skipped).

- [ ] **Step 1: Wire the service**, `go build ./...` → success
- [ ] **Step 2: Run unit + contract suites** — `go test ./ikuzo/service/x/semanticv1/ ./ikuzo/storage/x/v2adapter/` → PASS
- [ ] **Step 3: Run e2e against local stack** — `./test_semanticv1_e2e.sh` with dev server + ES running; all sections pass
- [ ] **Step 4: Commit**

```bash
git add ikuzo/ikuzoctl/cmd/config/semantic.go test_semanticv1_e2e.sh
git commit -m "feat(semanticv1): wire greenfield service; e2e suite for the frozen contract"
```

### Task 11: Cutover — delete the old stack (gated on user review)

**Do not start this task until the user has reviewed a running instance of the new service.**

**Files:**
- Delete: `ikuzo/service/x/semantic/` (entire package)
- Delete: `ikuzo/domain/semantic/` (entire package)
- Delete: `ikuzo/storage/x/v2adapter/` (entire package — the new bridge lives in `semanticv1/internal/v2bridge`, so nothing remains here)
- Delete: `ikuzo/storage/x/elasticsearch8/` (deferred native backend — restored from git when that phase starts)
- Delete: `ikuzo/storage/x/elasticsearch/semantic_store.go`, `semantic_store_test.go`, `facets.go` (verify `facets.go` has no other importers first: `grep -rn "facets\." ikuzo/storage/x/elasticsearch/`)
- Modify: `tools/cmd/gendocs/main.go` — regenerate `docs/semantic-api-reference.md` from the new `/docs` output or delete the tool if superseded
- Modify: `docs/v2-feature-freeze.md`, `docs/v2-to-semantic-migration-guide.md` — rewrite param tables to the frozen surface; every row states works/removed, no "Planned" ambiguity
- Delete: `test_semantic_e2e.sh` (superseded by `test_semanticv1_e2e.sh`)

- [ ] **Step 1: Delete + fix compilation** — `go build ./...` → success
- [ ] **Step 2: Full test sweep** — `go test ./...` → zero failures
- [ ] **Step 3: Update docs** as listed
- [ ] **Step 4: Single cutover commit** (revertible as one unit)

```bash
git add -A
git commit -m "feat(semantic)!: cut over to greenfield semanticv1, remove old semantic stack

BREAKING CHANGE: /api/semantic/v1 now rejects the previously accepted
dead parameters (detailLevel, cursor, contextIndex, languages, expand,
fields, peek) with 400 hydra:Error. Filter field names use the
underscore form only. Introspection, typed search, and search-context
navigation are removed pending phase 2."
```

---

## Self-review notes

- Spec coverage: every Part 1 contract row maps to Task 2/3 (parsing), Task 5 (v2 mapping), Task 7/8 (envelope + endpoints), Task 9 (docs), Task 10 (e2e). D1–D9 all encoded in tasks; D9 = Task 8's contract suite.
- Known intentional gaps: `hub3.delving.org/ns/hub3#` vocabulary IRI in the context file is a placeholder namespace choice the user should confirm (it only needs to be stable, not resolvable, for v1). Task 6 ports `decodeV2Results`/`GetByID` by explicit reference to the old `v2adapter/adapter.go` — the implementer must verify exact helper and protobuf getter names (`ScrollResultV4`, `QueryFacet`, `FacetLink`) against the real types at execution time; the old package remains present until Task 11 precisely so it can be ported from.
- Bridge retirement path (D10): when a native backend passes `contract_test.go`, delete `ikuzo/service/x/semanticv1/internal/v2bridge/` and `ikuzo/service/x/semanticv1/v2.go`, swap the constructor in the wiring — nothing else can reference the bridge, by compiler rule.
- Type consistency verified: `QueryOptions`/`Filter`/`SearchResult` field names match across Tasks 2–8; `ContextURL`/`ContextURLPath` (Task 1) used in Tasks 7–9.
