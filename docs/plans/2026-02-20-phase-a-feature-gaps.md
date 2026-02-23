# Phase A: Feature Gap Closure — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Close the 6 remaining search feature gaps (collapse, facet bool logic, facet cursor, hidden filters, peek, debug) so the semantic API can fully replace v2.

**Architecture:** Each feature follows the same 3-layer pattern: (1) add domain types in `ikuzo/domain/semantic/`, (2) add GET/POST parsing in `ikuzo/service/x/semantic/parser.go`, (3) add v2 adapter translation in `ikuzo/storage/x/v2adapter/query_translator.go`. All features are additive — they extend `QueryOptions` with new optional fields and don't break existing functionality.

**Tech Stack:** Go 1.22, chi router, Elasticsearch via v2adapter bridge, table-driven tests

**Key files reference:**
- Domain types: `ikuzo/domain/semantic/query.go` (QueryOptions struct, lines 10-37)
- Filter types: `ikuzo/domain/semantic/filter.go` (PropertyFilter, lines 27-33)
- Parser: `ikuzo/service/x/semantic/parser.go` (parseQueryParams, lines 44-97)
- V2 translator: `ikuzo/storage/x/v2adapter/query_translator.go` (TranslateToV2Query, lines 25-68)
- Response builder: `ikuzo/service/x/semantic/response.go` (buildCollection, lines 15-65)
- Store interface: `ikuzo/domain/semantic/store.go` (SearchStore, lines 11-32)

---

### Task 1: Add CollapseOptions to domain types

**Files:**
- Modify: `ikuzo/domain/semantic/query.go:10-37`
- Test: `ikuzo/domain/semantic/query_test.go` (create if not exists)

**Step 1: Write the failing test**

Create test file `ikuzo/domain/semantic/query_test.go`:

```go
package semantic

import (
	"testing"
)

func TestCollapseOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *CollapseOptions
		wantErr bool
	}{
		{
			name:    "nil is valid (no collapse)",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid collapse",
			opts:    &CollapseOptions{Field: "edm:dataProvider"},
			wantErr: false,
		},
		{
			name:    "valid with size",
			opts:    &CollapseOptions{Field: "edm:dataProvider", Size: 3},
			wantErr: false,
		},
		{
			name:    "empty field is invalid",
			opts:    &CollapseOptions{Field: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CollapseOptions.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueryOptions_HasCollapse(t *testing.T) {
	opts := &QueryOptions{}
	if opts.Collapse != nil {
		t.Error("expected nil Collapse on new QueryOptions")
	}

	opts.Collapse = &CollapseOptions{Field: "edm:dataProvider"}
	if opts.Collapse == nil {
		t.Error("expected non-nil Collapse")
	}
	if opts.Collapse.Field != "edm:dataProvider" {
		t.Errorf("got field %q, want %q", opts.Collapse.Field, "edm:dataProvider")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/domain/semantic/ -run "TestCollapse|TestQueryOptions_HasCollapse" -v`
Expected: FAIL — `CollapseOptions` type not defined

**Step 3: Write minimal implementation**

Add to `ikuzo/domain/semantic/query.go` — add `Collapse` field to `QueryOptions` (after line 36) and new `CollapseOptions` struct:

```go
// In QueryOptions struct, add after Fields:
	// Collapse groups results by a field value (e.g., dedup by data provider).
	Collapse *CollapseOptions

// New struct after QueryOptions:

// CollapseOptions configures result collapsing (grouping/deduplication).
type CollapseOptions struct {
	// Field is the field to collapse on (required).
	Field string `json:"field"`
	// Size is the number of inner hits per group (default: 1).
	Size int `json:"size,omitempty"`
	// Sort specifies sort order for inner hits.
	Sort []SortField `json:"sort,omitempty"`
}

// Validate checks if the collapse options are valid.
func (co *CollapseOptions) Validate() error {
	if co == nil {
		return nil
	}
	if co.Field == "" {
		return fmt.Errorf("collapse field cannot be empty")
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/domain/semantic/ -run "TestCollapse|TestQueryOptions_HasCollapse" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/domain/semantic/query.go ikuzo/domain/semantic/query_test.go
git commit -m "feat(semantic): add CollapseOptions to QueryOptions"
```

---

### Task 2: Add FacetBoolType, facet Cursor, Hidden flag, Peek, and Debug to domain types

These are all small additions to `QueryOptions`. Bundle them in one task to avoid excessive churn.

**Files:**
- Modify: `ikuzo/domain/semantic/query.go:10-37`
- Modify: `ikuzo/domain/semantic/filter.go:27-33` (add Hidden to PropertyFilter)
- Test: `ikuzo/domain/semantic/query_test.go`

**Step 1: Write the failing tests**

Append to `ikuzo/domain/semantic/query_test.go`:

```go
func TestFacetBoolType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		fbt   FacetBoolType
		valid bool
	}{
		{"or", FacetBoolOr, true},
		{"and", FacetBoolAnd, true},
		{"empty defaults to or", "", true},
		{"invalid", FacetBoolType("xor"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fbt.IsValid(); got != tt.valid {
				t.Errorf("FacetBoolType(%q).IsValid() = %v, want %v", tt.fbt, got, tt.valid)
			}
		})
	}
}

func TestQueryOptions_NewFields(t *testing.T) {
	opts := &QueryOptions{
		FacetBoolType: FacetBoolAnd,
		Peek:          true,
		Debug:         "query",
	}

	if opts.FacetBoolType != FacetBoolAnd {
		t.Errorf("FacetBoolType = %q, want %q", opts.FacetBoolType, FacetBoolAnd)
	}
	if !opts.Peek {
		t.Error("Peek should be true")
	}
	if opts.Debug != "query" {
		t.Errorf("Debug = %q, want %q", opts.Debug, "query")
	}
}

func TestFacetRequest_Cursor(t *testing.T) {
	fr := FacetRequest{
		Field:  "dc:creator",
		Limit:  50,
		Cursor: "abc123",
	}
	if fr.Cursor != "abc123" {
		t.Errorf("FacetRequest.Cursor = %q, want %q", fr.Cursor, "abc123")
	}
}

func TestPropertyFilter_Hidden(t *testing.T) {
	pf := &PropertyFilter{
		FieldName:    "orgID",
		OperatorType: OpEqual,
		Value:        "museum-x",
		Hidden:       true,
	}
	if !pf.Hidden {
		t.Error("PropertyFilter.Hidden should be true")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/domain/semantic/ -run "TestFacetBoolType|TestQueryOptions_NewFields|TestFacetRequest_Cursor|TestPropertyFilter_Hidden" -v`
Expected: FAIL — types/fields not defined

**Step 3: Write minimal implementation**

Add to `ikuzo/domain/semantic/query.go`:

```go
// FacetBoolType controls how multiple facet selections combine.
type FacetBoolType string

const (
	FacetBoolOr  FacetBoolType = "or"  // Selected values broaden results (default)
	FacetBoolAnd FacetBoolType = "and" // Selected values narrow results
)

// IsValid returns true if the facet bool type is valid.
func (fbt FacetBoolType) IsValid() bool {
	switch fbt {
	case FacetBoolOr, FacetBoolAnd, "":
		return true
	default:
		return false
	}
}
```

Add these fields to `QueryOptions` struct (after `Collapse`):

```go
	// FacetBoolType controls AND/OR logic for facet selections.
	FacetBoolType FacetBoolType

	// Peek when true returns only facets with zero items (size=0).
	Peek bool

	// Debug when set returns diagnostic information (e.g., "query" shows ES query).
	Debug string
```

Add `Cursor` field to `FacetRequest` struct (after `Sort`):

```go
	Cursor string `json:"cursor,omitempty"` // Opaque cursor for paginating facet values
```

Add `Hidden` field to `PropertyFilter` in `ikuzo/domain/semantic/filter.go` (after `Value`):

```go
	Hidden bool `json:"hidden,omitempty"` // If true, not shown in activeFilters response
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/domain/semantic/ -run "TestFacetBoolType|TestQueryOptions_NewFields|TestFacetRequest_Cursor|TestPropertyFilter_Hidden" -v`
Expected: PASS

**Step 5: Run all existing tests to verify no regressions**

Run: `go test ./ikuzo/domain/semantic/ -v`
Expected: All existing tests pass

**Step 6: Commit**

```bash
git add ikuzo/domain/semantic/query.go ikuzo/domain/semantic/filter.go ikuzo/domain/semantic/query_test.go
git commit -m "feat(semantic): add FacetBoolType, facet cursor, Hidden filter flag, Peek, Debug to domain types"
```

---

### Task 3: Parse collapse from GET/POST requests

**Files:**
- Modify: `ikuzo/service/x/semantic/parser.go:44-97`
- Test: `ikuzo/service/x/semantic/parser_test.go`

**Step 1: Write the failing test**

Add to `ikuzo/service/x/semantic/parser_test.go` (read file first to find appropriate location):

```go
func TestParseQueryParams_Collapse(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantField string
		wantSize  int
		wantNil   bool
	}{
		{
			name:    "no collapse",
			query:   "query=test",
			wantNil: true,
		},
		{
			name:      "collapse field only",
			query:     "collapse=edm_dataProvider",
			wantField: "edm:dataProvider",
			wantSize:  0,
		},
		{
			name:      "collapse with size",
			query:     "collapse=edm_dataProvider&collapse.size=3",
			wantField: "edm:dataProvider",
			wantSize:  3,
		},
		{
			name:      "collapse with sort",
			query:     "collapse=edm_dataProvider&collapse.sort=-dc_date",
			wantField: "edm:dataProvider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/search?"+tt.query, nil)
			opts, err := parseQueryParams(r)
			if err != nil {
				t.Fatalf("parseQueryParams() error = %v", err)
			}
			if tt.wantNil {
				if opts.Collapse != nil {
					t.Error("expected nil Collapse")
				}
				return
			}
			if opts.Collapse == nil {
				t.Fatal("expected non-nil Collapse")
			}
			if opts.Collapse.Field != tt.wantField {
				t.Errorf("Collapse.Field = %q, want %q", opts.Collapse.Field, tt.wantField)
			}
			if tt.wantSize > 0 && opts.Collapse.Size != tt.wantSize {
				t.Errorf("Collapse.Size = %d, want %d", opts.Collapse.Size, tt.wantSize)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestParseQueryParams_Collapse" -v`
Expected: FAIL — Collapse not parsed

**Step 3: Write minimal implementation**

Add `parseCollapseFromQuery` function to `ikuzo/service/x/semantic/parser.go` and call it from `parseQueryParams` (after sort parsing, around line 79):

```go
// In parseQueryParams, after sort parsing:
	// Parse collapse
	if err := parseCollapseFromQuery(query, opts); err != nil {
		return nil, fmt.Errorf("failed to parse collapse: %w", err)
	}
```

New function:

```go
// parseCollapseFromQuery parses collapse parameters from URL query.
// Format: collapse=field&collapse.size=3&collapse.sort=-dc_date
func parseCollapseFromQuery(query url.Values, opts *semantic.QueryOptions) error {
	collapseField := query.Get("collapse")
	if collapseField == "" {
		return nil
	}

	opts.Collapse = &semantic.CollapseOptions{
		Field: fromURLFieldName(collapseField),
	}

	if sizeStr := query.Get("collapse.size"); sizeStr != "" {
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return fmt.Errorf("invalid collapse.size: %w", err)
		}
		opts.Collapse.Size = size
	}

	// Parse collapse sort (same format as main sort: -field for desc)
	if sortParam := query.Get("collapse.sort"); sortParam != "" {
		direction := semantic.SortAsc
		field := sortParam
		if strings.HasPrefix(sortParam, "-") {
			field = strings.TrimPrefix(sortParam, "-")
			direction = semantic.SortDesc
		}
		opts.Collapse.Sort = []semantic.SortField{{
			Field:     fromURLFieldName(field),
			Direction: direction,
		}}
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestParseQueryParams_Collapse" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/service/x/semantic/parser.go ikuzo/service/x/semantic/parser_test.go
git commit -m "feat(semantic): parse collapse parameters from GET requests"
```

---

### Task 4: Parse facetBool, peek, debug, hidden filters, and facet cursor from GET requests

**Files:**
- Modify: `ikuzo/service/x/semantic/parser.go`
- Test: `ikuzo/service/x/semantic/parser_test.go`

**Step 1: Write the failing tests**

Append to `parser_test.go`:

```go
func TestParseQueryParams_FacetBool(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?facetBool=and", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("parseQueryParams() error = %v", err)
	}
	if opts.FacetBoolType != semantic.FacetBoolAnd {
		t.Errorf("FacetBoolType = %q, want %q", opts.FacetBoolType, semantic.FacetBoolAnd)
	}
}

func TestParseQueryParams_Peek(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?peek=dc_creator,dc_type", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("parseQueryParams() error = %v", err)
	}
	if !opts.Peek {
		t.Error("Peek should be true")
	}
	// Peek also creates facet requests for specified fields
	if len(opts.Facets) != 2 {
		t.Fatalf("expected 2 facets from peek, got %d", len(opts.Facets))
	}
	if opts.Facets[0].Field != "dc:creator" {
		t.Errorf("Facets[0].Field = %q, want %q", opts.Facets[0].Field, "dc:creator")
	}
	// Peek sets pagination size to 0
	if opts.Pagination == nil || opts.Pagination.Size != 0 {
		t.Error("Peek should set pagination size to 0")
	}
}

func TestParseQueryParams_Debug(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?debug=query", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("parseQueryParams() error = %v", err)
	}
	if opts.Debug != "query" {
		t.Errorf("Debug = %q, want %q", opts.Debug, "query")
	}
}

func TestParseQueryParams_HiddenFilter(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?hfilter[orgID][eq]=museum-x", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("parseQueryParams() error = %v", err)
	}
	if len(opts.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(opts.Filters))
	}
	pf, ok := opts.Filters[0].(*semantic.PropertyFilter)
	if !ok {
		t.Fatal("expected PropertyFilter")
	}
	if !pf.Hidden {
		t.Error("filter should be hidden")
	}
	if pf.FieldName != "orgID" {
		t.Errorf("field = %q, want %q", pf.FieldName, "orgID")
	}
}

func TestParseQueryParams_FacetCursor(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?facet[dc_creator]=50&facet[dc_creator].cursor=abc123", nil)
	opts, err := parseQueryParams(r)
	if err != nil {
		t.Fatalf("parseQueryParams() error = %v", err)
	}
	if len(opts.Facets) == 0 {
		t.Fatal("expected at least 1 facet")
	}
	// Find the dc:creator facet
	var found bool
	for _, f := range opts.Facets {
		if f.Field == "dc:creator" {
			found = true
			if f.Cursor != "abc123" {
				t.Errorf("Cursor = %q, want %q", f.Cursor, "abc123")
			}
			if f.Limit != 50 {
				t.Errorf("Limit = %d, want %d", f.Limit, 50)
			}
		}
	}
	if !found {
		t.Error("dc:creator facet not found")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestParseQueryParams_(FacetBool|Peek|Debug|HiddenFilter|FacetCursor)" -v`
Expected: FAIL — new params not parsed

**Step 3: Write minimal implementation**

Modify `parseQueryParams` in `parser.go` — add after the existing "Parse other options" section (around line 81):

```go
	// Parse facet bool type
	if fbt := query.Get("facetBool"); fbt != "" {
		opts.FacetBoolType = semantic.FacetBoolType(fbt)
		if !opts.FacetBoolType.IsValid() {
			return nil, fmt.Errorf("invalid facetBool value: %q (must be 'and' or 'or')", fbt)
		}
	}

	// Parse debug mode
	opts.Debug = query.Get("debug")

	// Parse peek mode: ?peek=field1,field2 (returns only facets, zero items)
	if peekFields := query.Get("peek"); peekFields != "" {
		opts.Peek = true
		for _, urlField := range strings.Split(peekFields, ",") {
			field := fromURLFieldName(strings.TrimSpace(urlField))
			if field != "" {
				opts.Facets = append(opts.Facets, semantic.FacetRequest{Field: field})
			}
		}
		// Peek forces zero items
		if opts.Pagination == nil {
			opts.Pagination = &semantic.Pagination{}
		}
		opts.Pagination.Size = 0
	}
```

Modify `parseFiltersFromQuery` to also handle `hfilter[field][op]=value` (hidden filters). Add at the top of the loop body, before the existing `filter[` check:

```go
	for key, values := range query {
		hidden := false
		prefix := "filter["
		if strings.HasPrefix(key, "hfilter[") {
			hidden = true
			prefix = "hfilter["
		} else if !strings.HasPrefix(key, "filter[") {
			continue
		}

		// Extract field and operator from (h)filter[field][operator]
		parts := strings.Split(key, "][")
		if len(parts) != 2 {
			continue
		}

		urlField := strings.TrimPrefix(parts[0], prefix)
		operator := strings.TrimSuffix(parts[1], "]")
		// ... rest is same, but set Hidden on PropertyFilter
```

After creating the `PropertyFilter` (around line 142), set the `Hidden` flag:

```go
		pf := &semantic.PropertyFilter{
			FieldName:    field,
			OperatorType: op,
			Value:        value,
			Hidden:       hidden,
		}
		opts.Filters = append(opts.Filters, pf)
```

Modify `parseFacetsFromQuery` to also handle `facet[field].cursor=abc` and `facet[field].sort=count`:

```go
func parseFacetsFromQuery(query url.Values, opts *semantic.QueryOptions) error {
	// First pass: collect facet requests by field
	facetMap := make(map[string]*semantic.FacetRequest)

	for key, values := range query {
		if !strings.HasPrefix(key, "facet[") {
			continue
		}

		// Check for sub-parameters: facet[field].cursor, facet[field].sort
		fullKey := strings.TrimPrefix(key, "facet[")
		if idx := strings.Index(fullKey, "]."); idx >= 0 {
			urlField := fullKey[:idx]
			subParam := fullKey[idx+2:]
			field := fromURLFieldName(urlField)

			fr, ok := facetMap[field]
			if !ok {
				fr = &semantic.FacetRequest{Field: field}
				facetMap[field] = fr
			}

			switch subParam {
			case "cursor":
				if len(values) > 0 {
					fr.Cursor = values[0]
				}
			case "sort":
				if len(values) > 0 {
					fr.Sort = values[0]
				}
			}
			continue
		}

		// Main facet parameter: facet[field]=limit
		urlField := strings.TrimSuffix(fullKey, "]")
		field := fromURLFieldName(urlField)

		if len(values) == 0 {
			continue
		}

		fr, ok := facetMap[field]
		if !ok {
			fr = &semantic.FacetRequest{Field: field}
			facetMap[field] = fr
		}

		if limit, err := strconv.Atoi(values[0]); err == nil && limit > 0 {
			fr.Limit = limit
		}
	}

	for _, fr := range facetMap {
		opts.Facets = append(opts.Facets, *fr)
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestParseQueryParams_(FacetBool|Peek|Debug|HiddenFilter|FacetCursor)" -v`
Expected: PASS

**Step 5: Run all parser tests to verify no regressions**

Run: `go test ./ikuzo/service/x/semantic/ -v`
Expected: All existing tests pass

**Step 6: Commit**

```bash
git add ikuzo/service/x/semantic/parser.go ikuzo/service/x/semantic/parser_test.go
git commit -m "feat(semantic): parse facetBool, peek, debug, hidden filters, facet cursor from GET"
```

---

### Task 5: Translate new features to v2 adapter

**Files:**
- Modify: `ikuzo/storage/x/v2adapter/query_translator.go:25-68`
- Test: `ikuzo/storage/x/v2adapter/query_translator_test.go`

**Step 1: Write the failing tests**

Add to or create `ikuzo/storage/x/v2adapter/query_translator_test.go`:

```go
package v2adapter

import (
	"testing"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

func TestTranslateToV2Query_Collapse(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		Collapse: &semantic.CollapseOptions{
			Field: "edm:dataProvider",
			Size:  3,
		},
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("TranslateToV2Query() error = %v", err)
	}

	if got := params.Get("collapseOn"); got != "edm:dataProvider" {
		t.Errorf("collapseOn = %q, want %q", got, "edm:dataProvider")
	}
	if got := params.Get("collapseSize"); got != "3" {
		t.Errorf("collapseSize = %q, want %q", got, "3")
	}
}

func TestTranslateToV2Query_FacetBoolType(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		FacetBoolType: semantic.FacetBoolAnd,
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("TranslateToV2Query() error = %v", err)
	}

	if got := params.Get("facetBoolType"); got != "and" {
		t.Errorf("facetBoolType = %q, want %q", got, "and")
	}
}

func TestTranslateToV2Query_HiddenFilter(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		Filters: []semantic.Filter{
			&semantic.PropertyFilter{
				FieldName:    "orgID",
				OperatorType: semantic.OpEqual,
				Value:        "museum-x",
				Hidden:       true,
			},
			&semantic.PropertyFilter{
				FieldName:    "dc:type",
				OperatorType: semantic.OpEqual,
				Value:        "painting",
			},
		},
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("TranslateToV2Query() error = %v", err)
	}

	// Hidden filter should use hqf param
	hqfValues := params["hqf"]
	if len(hqfValues) != 1 {
		t.Fatalf("expected 1 hqf value, got %d: %v", len(hqfValues), hqfValues)
	}
	if hqfValues[0] != "orgID:museum-x" {
		t.Errorf("hqf[0] = %q, want %q", hqfValues[0], "orgID:museum-x")
	}

	// Normal filter should use qf param
	qfValues := params["qf"]
	if len(qfValues) != 1 {
		t.Fatalf("expected 1 qf value, got %d: %v", len(qfValues), qfValues)
	}
}

func TestTranslateToV2Query_Peek(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		Peek: true,
		Facets: []semantic.FacetRequest{
			{Field: "dc:creator"},
			{Field: "dc:type"},
		},
		Pagination: &semantic.Pagination{Size: 0},
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("TranslateToV2Query() error = %v", err)
	}

	if got := params.Get("peek"); got != "dc:creator,dc:type" {
		t.Errorf("peek = %q, want %q", got, "dc:creator,dc:type")
	}
}

func TestTranslateToV2Query_Debug(t *testing.T) {
	qt := NewQueryTranslator("test-org")
	opts := &semantic.QueryOptions{
		Debug: "query",
	}

	params, err := qt.TranslateToV2Query(opts)
	if err != nil {
		t.Fatalf("TranslateToV2Query() error = %v", err)
	}

	if got := params.Get("echo"); got != "searchResponse" {
		t.Errorf("echo = %q, want %q (debug=query maps to echo=searchResponse)", got, "searchResponse")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/storage/x/v2adapter/ -run "TestTranslateToV2Query_(Collapse|FacetBoolType|HiddenFilter|Peek|Debug)" -v`
Expected: FAIL — features not translated

**Step 3: Write minimal implementation**

Modify `TranslateToV2Query` in `query_translator.go` — add after the sort translation (around line 65):

```go
	// Add collapse
	if opts.Collapse != nil {
		qt.translateCollapse(opts.Collapse, params)
	}

	// Add facet bool type
	if opts.FacetBoolType != "" {
		params.Set("facetBoolType", string(opts.FacetBoolType))
	}

	// Add peek mode
	if opts.Peek && len(opts.Facets) > 0 {
		peekFields := make([]string, len(opts.Facets))
		for i, f := range opts.Facets {
			peekFields[i] = f.Field
		}
		params.Set("peek", strings.Join(peekFields, ","))
	}

	// Add debug mode
	if opts.Debug != "" {
		// v2 uses echo=searchResponse for query debugging
		params.Set("echo", "searchResponse")
	}
```

Add new method:

```go
// translateCollapse converts collapse options to v2 parameters.
func (qt *QueryTranslator) translateCollapse(co *semantic.CollapseOptions, params url.Values) {
	params.Set("collapseOn", co.Field)
	if co.Size > 0 {
		params.Set("collapseSize", fmt.Sprintf("%d", co.Size))
	}
	if len(co.Sort) > 0 {
		sort := co.Sort[0]
		params.Set("collapseSort", sort.Field)
	}
}
```

Modify `translateFilters` to handle hidden filters — change line 105 from `params.Add("qf", ...)` to check the hidden flag:

```go
func (qt *QueryTranslator) translateFilters(filters []semantic.Filter, params url.Values) error {
	for _, filter := range filters {
		filterStr, err := qt.translateFilter(filter)
		if err != nil {
			return err
		}

		if filterStr != "" {
			// Use hqf for hidden filters, qf for normal
			paramKey := "qf"
			if pf, ok := filter.(*semantic.PropertyFilter); ok && pf.Hidden {
				paramKey = "hqf"
			}
			params.Add(paramKey, filterStr)
		}
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/storage/x/v2adapter/ -run "TestTranslateToV2Query_(Collapse|FacetBoolType|HiddenFilter|Peek|Debug)" -v`
Expected: PASS

**Step 5: Run all v2adapter tests to verify no regressions**

Run: `go test ./ikuzo/storage/x/v2adapter/ -run "TestTranslate" -v`
Expected: All translator tests pass (skip integration tests that need live ES)

**Step 6: Commit**

```bash
git add ikuzo/storage/x/v2adapter/query_translator.go ikuzo/storage/x/v2adapter/query_translator_test.go
git commit -m "feat(v2adapter): translate collapse, facetBool, hidden filters, peek, debug to v2 params"
```

---

### Task 6: Handle hidden filters in response breadcrumbs

Hidden filters should not appear in the `activeFilters` breadcrumb list.

**Files:**
- Modify: `ikuzo/service/x/semantic/response.go:152-170`
- Test: `ikuzo/service/x/semantic/response_test.go` (create or append)

**Step 1: Write the failing test**

```go
func TestBuildBreadcrumbs_HidesHiddenFilters(t *testing.T) {
	svc := &Service{
		registry: semantic.DefaultRegistry(),
		baseURL:  "/api/semantic/v1",
	}

	opts := &semantic.QueryOptions{
		Filters: []semantic.Filter{
			&semantic.PropertyFilter{
				FieldName:    "dc:type",
				OperatorType: semantic.OpEqual,
				Value:        "painting",
			},
			&semantic.PropertyFilter{
				FieldName:    "orgID",
				OperatorType: semantic.OpEqual,
				Value:        "museum-x",
				Hidden:       true,
			},
		},
	}

	r := httptest.NewRequest("GET", "/search?filter[dc_type][eq]=painting&hfilter[orgID][eq]=museum-x", nil)
	breadcrumbs := svc.buildBreadcrumbs(r, opts)

	// Only the non-hidden filter should appear
	if len(breadcrumbs.ItemListElement) != 1 {
		t.Fatalf("expected 1 breadcrumb, got %d", len(breadcrumbs.ItemListElement))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestBuildBreadcrumbs_HidesHiddenFilters" -v`
Expected: FAIL — hidden filter appears in breadcrumbs (2 items instead of 1)

**Step 3: Write minimal implementation**

Modify `buildBreadcrumbs` in `response.go` (lines 152-170) to skip hidden filters:

```go
func (s *Service) buildBreadcrumbs(
	r *http.Request,
	opts *semantic.QueryOptions,
) *semantic.BreadcrumbList {
	var items []semantic.BreadcrumbListItem

	for _, filter := range opts.Filters {
		// Skip hidden filters
		if pf, ok := filter.(*semantic.PropertyFilter); ok && pf.Hidden {
			continue
		}

		items = append(items, semantic.BreadcrumbListItem{
			Position:  len(items) + 1,
			Name:      s.buildFilterLabel(filter),
			RemoveURL: s.buildFilterRemoveURL(r, filter, opts),
		})
	}

	return &semantic.BreadcrumbList{
		ItemListElement: items,
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestBuildBreadcrumbs_HidesHiddenFilters" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/service/x/semantic/response.go ikuzo/service/x/semantic/response_test.go
git commit -m "feat(semantic): exclude hidden filters from breadcrumb response"
```

---

### Task 7: Add debug section to search response

When `debug=query` is set, include a `hub3:debug` section in the Hydra Collection response showing the generated query.

**Files:**
- Modify: `ikuzo/service/x/semantic/response.go:15-65` (buildCollection)
- Modify: `ikuzo/domain/semantic/collection.go` (add Debug field to Collection if not present — check file first)
- Test: `ikuzo/service/x/semantic/response_test.go`

**Step 1: Write the failing test**

```go
func TestBuildCollection_Debug(t *testing.T) {
	svc := &Service{
		registry: semantic.DefaultRegistry(),
		baseURL:  "/api/semantic/v1",
	}

	result := &semantic.SearchResult{
		Total:   5,
		Results: []map[string]any{},
		Metadata: map[string]any{
			"v2_query": "some ES query string",
		},
	}

	opts := &semantic.QueryOptions{
		Debug:      "query",
		Pagination: &semantic.Pagination{Page: 1, Size: 20},
	}

	config := svc.registry.GetDefault()
	r := httptest.NewRequest("GET", "/search?debug=query", nil)

	collection := svc.buildCollection(r, result, opts, config)

	// Collection should have debug info when debug mode is set
	if collection.Debug == nil {
		t.Fatal("expected debug section in collection")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestBuildCollection_Debug" -v`
Expected: FAIL — Debug field doesn't exist or not populated

**Step 3: Write minimal implementation**

First, check and modify the Collection type to include a Debug field. Then modify `buildCollection` to populate it:

In `buildCollection` (response.go), add after the timing section (around line 62):

```go
	// Add debug info if debug mode is enabled
	if opts.Debug != "" && result.Metadata != nil {
		collection.Debug = result.Metadata
	}
```

If `Collection` struct in `ikuzo/domain/semantic/collection.go` doesn't have a `Debug` field, add:

```go
	// Debug contains diagnostic information when debug mode is enabled.
	Debug map[string]any `json:"hub3:debug,omitempty"`
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestBuildCollection_Debug" -v`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./ikuzo/service/x/semantic/ -v`
Expected: All pass

**Step 6: Commit**

```bash
git add ikuzo/domain/semantic/collection.go ikuzo/service/x/semantic/response.go ikuzo/service/x/semantic/response_test.go
git commit -m "feat(semantic): include debug metadata in search response"
```

---

### Task 8: Add FacetResult cursor support in response

When facet results have more values than returned, the response should include a `hub3:nextCursor` for pagination.

**Files:**
- Modify: `ikuzo/domain/semantic/store.go:58-67` (FacetResult — add NextCursor field)
- Modify: `ikuzo/service/x/semantic/response.go:104-150` (buildFacets)
- Test: `ikuzo/service/x/semantic/response_test.go`

**Step 1: Write the failing test**

```go
func TestBuildFacets_IncludesCursor(t *testing.T) {
	svc := &Service{
		registry: semantic.DefaultRegistry(),
		baseURL:  "/api/semantic/v1",
	}

	results := []semantic.FacetResult{
		{
			Field:       "dc:creator",
			FacetType:   "enum",
			TotalValues: 500,
			NextCursor:  "eyJjcmVhdG9yIjoiUmVtYnJhbmR0In0=",
			Values: []semantic.FacetValueResult{
				{Value: "Rembrandt", Count: 42},
			},
		},
	}

	opts := &semantic.QueryOptions{}
	r := httptest.NewRequest("GET", "/search", nil)

	facets := svc.buildFacets(r, results, opts)
	if len(facets) != 1 {
		t.Fatalf("expected 1 facet, got %d", len(facets))
	}
	if facets[0].NextCursor != "eyJjcmVhdG9yIjoiUmVtYnJhbmR0In0=" {
		t.Errorf("NextCursor = %q, want encoded cursor", facets[0].NextCursor)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestBuildFacets_IncludesCursor" -v`
Expected: FAIL — NextCursor field doesn't exist

**Step 3: Write minimal implementation**

Add `NextCursor` field to `FacetResult` in `store.go`:

```go
type FacetResult struct {
	Field       string
	FacetType   string
	TotalValues int64
	Values      []FacetValueResult
	SumOther    int64
	Missing     int64
	Error       string
	NextCursor  string // Opaque cursor for fetching next page of facet values
}
```

Add `NextCursor` to `Facet` in `collection.go` (check the Facet struct first):

```go
	NextCursor string `json:"hub3:nextCursor,omitempty"`
```

In `buildFacets` (response.go), add after populating facet values:

```go
		if result.NextCursor != "" {
			facet.NextCursor = result.NextCursor
		}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/service/x/semantic/ -run "TestBuildFacets_IncludesCursor" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add ikuzo/domain/semantic/store.go ikuzo/domain/semantic/collection.go ikuzo/service/x/semantic/response.go ikuzo/service/x/semantic/response_test.go
git commit -m "feat(semantic): add facet cursor support in response"
```

---

### Task 9: Validate all new query options

The `ResourceConfig.ValidateQueryOptions` should validate collapse field, facet bool type, and debug values.

**Files:**
- Modify: `ikuzo/domain/semantic/config.go` (ValidateQueryOptions method)
- Test: `ikuzo/domain/semantic/config_test.go`

**Step 1: Write the failing tests**

```go
func TestValidateQueryOptions_Collapse(t *testing.T) {
	rc := semantic.NewResourceConfig("schema:CreativeWork", "Creative Work")

	tests := []struct {
		name    string
		opts    *semantic.QueryOptions
		wantErr bool
	}{
		{
			name:    "nil collapse is valid",
			opts:    &semantic.QueryOptions{},
			wantErr: false,
		},
		{
			name: "valid collapse",
			opts: &semantic.QueryOptions{
				Collapse: &semantic.CollapseOptions{Field: "edm:dataProvider"},
			},
			wantErr: false,
		},
		{
			name: "empty collapse field is invalid",
			opts: &semantic.QueryOptions{
				Collapse: &semantic.CollapseOptions{Field: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rc.ValidateQueryOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQueryOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateQueryOptions_FacetBoolType(t *testing.T) {
	rc := semantic.NewResourceConfig("schema:CreativeWork", "Creative Work")

	opts := &semantic.QueryOptions{
		FacetBoolType: semantic.FacetBoolType("xor"),
	}
	if err := rc.ValidateQueryOptions(opts); err == nil {
		t.Error("expected error for invalid facet bool type")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./ikuzo/domain/semantic/ -run "TestValidateQueryOptions_(Collapse|FacetBoolType)" -v`
Expected: FAIL — validation not implemented

**Step 3: Write minimal implementation**

Add to `ValidateQueryOptions` in `config.go` (find the method, add at the end before return):

```go
	// Validate collapse options
	if opts.Collapse != nil {
		if err := opts.Collapse.Validate(); err != nil {
			return fmt.Errorf("invalid collapse options: %w", err)
		}
	}

	// Validate facet bool type
	if !opts.FacetBoolType.IsValid() {
		return fmt.Errorf("invalid facetBoolType: %q (must be 'and' or 'or')", opts.FacetBoolType)
	}
```

**Step 4: Run test to verify it passes**

Run: `go test ./ikuzo/domain/semantic/ -run "TestValidateQueryOptions_(Collapse|FacetBoolType)" -v`
Expected: PASS

**Step 5: Run all domain tests**

Run: `go test ./ikuzo/domain/semantic/ -v`
Expected: All pass

**Step 6: Commit**

```bash
git add ikuzo/domain/semantic/config.go ikuzo/domain/semantic/config_test.go
git commit -m "feat(semantic): validate collapse and facetBoolType in query options"
```

---

### Task 10: Full build verification and integration

**Step 1: Run full build**

Run: `make build`
Expected: exit 0

**Step 2: Run all affected test packages**

Run: `go test ./ikuzo/domain/semantic/ ./ikuzo/service/x/semantic/ ./ikuzo/storage/x/v2adapter/ -v -count=1`
Expected: All pass (except pre-existing v2adapter integration test failures that need live ES)

**Step 3: Run staticcheck**

Run: `make staticcheck`
Expected: No new warnings

**Step 4: Verify no regressions across full test suite**

Run: `make test`
Expected: Same results as before Phase A changes

**Step 5: Commit (if any fixups needed)**

If staticcheck or tests reveal issues, fix and commit with:
```bash
git commit -m "fix(semantic): address lint/test issues from Phase A"
```

---

## Summary

| Task | Description | Files Modified | New Files |
|------|-------------|---------------|-----------|
| 1 | CollapseOptions domain type | query.go | query_test.go |
| 2 | FacetBoolType, Cursor, Hidden, Peek, Debug types | query.go, filter.go | — |
| 3 | Parse collapse from GET | parser.go | — |
| 4 | Parse facetBool, peek, debug, hidden, cursor | parser.go | — |
| 5 | Translate all to v2 adapter | query_translator.go | query_translator_test.go |
| 6 | Hidden filters in breadcrumbs | response.go | response_test.go |
| 7 | Debug section in response | response.go, collection.go | — |
| 8 | Facet cursor in response | store.go, collection.go, response.go | — |
| 9 | Validate new options | config.go | — |
| 10 | Full build verification | — | — |

**Dependency graph:**
```
Task 1,2 (domain types) → Task 3,4 (parsing) → Task 5 (v2 translation)
Task 1,2 (domain types) → Task 6,7,8 (response) → Task 9 (validation)
Task 1-9 → Task 10 (verification)
```

Tasks 3-4 and 6-8 can run in parallel since they touch different files.
