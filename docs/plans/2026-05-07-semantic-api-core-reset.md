# Semantic API Core Reset

**Date:** 2026-05-07
**Status:** Superseded by [ADR 0001](../adr/0001-semantic-v1-wraps-v2.md) for current delivery scope
**Purpose:** Re-establish the semantic API core before continuing implementation or V2 migration work.

> Current scope update, 2026-06-04: Semantic V1 is delivered as a JSON-LD/Hydra
> wrapper around the existing V2 search implementation. Native Elasticsearch
> backend work remains future development and must preserve the accepted Semantic
> V1 contract before replacing the adapter.

## Problem Statement

The current semantic API work mixes several concerns that should be separate:

- Public API contract: JSON-LD/Hydra search, detail, facets, introspection.
- Query model: fields, paths, operators, facets, sorting, pagination.
- Document model: `FragmentGraph` converted to `SemanticView`.
- Backend implementation: Elasticsearch over the existing V2 `resources.entries` index.
- Migration bridge: V2 feature parity and `v2adapter`.

This makes implementation choices look like core concepts. The result is drift in field naming, response shape, configuration, and backend responsibilities.

The reset goal is to define the smallest stable core, lock each concept deliberately, then derive philosophy, implementation boundaries, and V2 migration mapping from that core.

## Core Philosophy Draft

The semantic API is a self-describing query interface over RDF-shaped records. It exposes stable search and discovery concepts while treating indexed document content as JSON-LD data. Backends may store and query that data differently, but clients should see one coherent API model.

The core API must be:

- **Self-describing:** clients can discover operations, fields, paths, operators, facets, and response shapes from the API itself.
- **Backend-agnostic:** Elasticsearch is the first implementation, not the model.
- **Content-aware but content-decoupled:** the API can query RDF fields and paths, but it does not hardcode one vocabulary as the API structure.
- **Incremental:** simple field search must work first; graph/path depth is added only after the simple model is stable.
- **Migration-friendly, not V2-shaped:** V2 compatibility is an adapter and migration concern, not a design constraint.

## Core Concepts To Lock

### 1. Resource

A resource is the unit returned by search and detail.

Decisions to lock:

- Is a search hit always one top-level RDF resource, or can it be a document containing multiple resources?
- What is the canonical public identifier: `hubID`, `@id` URI, or both with different roles?
- Does detail return the resource document directly, or a `hub3:DetailView` envelope containing `hub3:item`?
- Is `FragmentGraph` part of the resource contract, or only an indexing/input adapter?

Proposed direction:

- Public content is JSON-LD.
- `@id` is the semantic resource identifier.
- `hubID` is an implementation/document identifier exposed as metadata when needed.
- `FragmentGraph` is not core. It is one source representation used by the first ES adapter.

### 2. Field

A field is a queryable RDF predicate or indexed value.

Decisions to lock:

- Canonical internal field syntax: `dc:creator`, full URI, or normalized `dc_creator`.
- Canonical external URL syntax.
- Whether field names identify predicates globally or fields on a specific resource class.
- How labels and documentation are discovered.

Proposed direction:

- Internal semantic field identity should be RDF-aware: compact IRI (`dc:creator`) or full URI.
- URL syntax may use a safe representation, but conversion must happen only at the HTTP edge.
- Introspection should publish available fields, labels, value types, counts, and valid query forms.
- Avoid allowing `dc:creator`, `dc_creator`, and `dc.creator` to coexist past the edge layer.

### 3. Path

A path is a way to query through linked resources.

Decisions to lock:

- Syntax for simple fields vs class-scoped fields vs multi-hop fields.
- Whether path syntax is public in v1 MVP or deferred.
- Maximum path depth and how clients discover valid paths.
- How paths map to ES nested structures and future SPARQL property paths.

Proposed direction:

- MVP supports simple fields only.
- Path syntax becomes a separately locked concept after simple fields are stable.
- Introspection must be the authority for valid paths before path querying is public.

### 4. Query

A query describes what resources to match.

Decisions to lock:

- GET parameter names: `q` vs `query`.
- POST request model and whether it contains request-driven config.
- Text query semantics: default fields, operators, fuzziness, language behavior.
- Whether hidden filters are core or migration-only.

Proposed direction:

- GET is for common query forms.
- POST is for structured queries and later request-driven configuration.
- MVP includes text query plus exact property filters.
- Hidden filters are useful for virtual datasets and scoped views, but should be framed as base constraints, not V2 parity.

### 5. Filter

A filter narrows matching resources by field/path and operator.

Decisions to lock:

- Operator vocabulary and exact semantics.
- Type coercion for dates, booleans, numbers, URIs, language-tagged literals.
- Whether `exists` takes a value.
- Whether invalid field/operator combinations come from static config or introspection.

Proposed direction:

- MVP operators: `eq`, `in`, `exists`, `gte`, `lte`.
- Additional text and geo operators come later.
- Validation should eventually use introspected capabilities, with a small static fallback only for bootstrap.

### 6. Facet

A facet describes value distribution for a field/path in a query scope.

Decisions to lock:

- Whether facets are part of search response only or have a dedicated exploration endpoint.
- Count semantics: affected by active filters, hidden/base filters, and selected facet values.
- Pagination/cursor model for facet values.
- Label resolution for URI values.

Proposed direction:

- MVP supports basic terms facets in search response.
- Deep facet browsing is a separate later endpoint.
- Label resolution is an indexing/introspection concern, not an HTTP response hack.

### 7. Sort

Sort defines result ordering.

Decisions to lock:

- Syntax and field validation.
- Multi-sort support.
- Whether sort fields must be introspected as sortable.
- How sorting behaves for multi-valued RDF fields.

Proposed direction:

- MVP supports relevance and one explicit field sort.
- Sortability is a field capability.
- Multi-valued sort semantics must be explicit before broad exposure.

### 8. Pagination And Query Context

Pagination controls result windows. Query context preserves search state for follow-up operations.

Decisions to lock:

- Page/size vs cursor as the public default.
- Whether query context is mandatory for navigation.
- Context storage ownership: search backend, cache service, Redis, database.
- Expiration and URL shape.

Proposed direction:

- MVP starts with page/size.
- Query context is a separate component, not part of `SearchStore`.
- Detail navigation uses query context once the search/detail contract is stable.

### 9. Introspection

Introspection describes what the data and backend can support.

Decisions to lock:

- Required MVP endpoints.
- Whether introspection is scoped by query context.
- Difference between data introspection and API capability documentation.
- How record definitions enrich introspection.

Proposed direction:

- Introspection is core, not an optional extra.
- MVP endpoints: classes, fields/properties, field details.
- Paths and schema overlays can follow after basic fields are reliable.

### 10. Response Envelope

The response envelope describes API structure; members contain content.

Decisions to lock:

- Use compact JSON-LD keys (`member`) or explicit Hydra keys (`hydra:member`) in raw JSON.
- Inline context vs URL context.
- Search collection structure.
- Detail response structure.
- Error structure and machine-readable error codes.

Proposed direction:

- Search response is a Hydra collection.
- Detail response choice must be locked early: direct JSON-LD resource vs `hub3:DetailView`.
- The API envelope and content document should not be mixed accidentally.

### 11. Backend Interface

Backend interfaces implement core query and discovery behavior.

Decisions to lock:

- Minimal interfaces and method ownership.
- Whether contexts/cache/MLT belong in the main search interface.
- How backend capabilities are exposed.
- What a future triplestore must implement to be first-class.

Proposed direction:

- Keep `SearchStore` small: `Search`, `Get`, `Aggregate`.
- Split `IntrospectionStore`, `SimilarStore`, and `QueryContextStore`.
- Keep backend-specific query builders and result parsers below the interface.

### 12. Profile / Vocabulary Configuration

Profiles describe domain-specific defaults such as EDM fields.

Decisions to lock:

- Whether server-side profiles exist in the core.
- How profiles relate to introspection.
- How clients request or discover profiles.

Proposed direction:

- EDM is a profile, not the semantic API core.
- Profiles can seed labels, preferred fields, and examples.
- Introspection should still work without profiles.

## Explicitly Not Core

These are implementation or migration concerns:

- `FragmentGraph` storage shape.
- `resources.entries.searchLabel` query mechanics.
- `GenerateSemantic()` implementation details.
- V2 `qf`, `hqf`, `facet.field`, `peek`, `collapseOn`, `mlt` parameter names.
- `olivere/elastic` vs `go-elasticsearch/v8`.
- Frontend-specific layout blocks.
- EAD/tree navigation.

## Lock-Down Sequence

The concepts should be finalized in this order:

1. **Resource identity and response shape:** decide what a search member and detail response are.
2. **Field identity:** settle compact IRI/full URI/URL-safe translation once.
3. **Simple query/filter/facet/sort model:** define the MVP query language without paths.
4. **Introspection model:** define how clients discover available fields and capabilities.
5. **Backend interfaces:** trim and split interfaces around the locked model.
6. **First ES adapter contract:** map the model to the existing FragmentGraph index explicitly as an adapter.
7. **Path model:** add graph traversal only after simple fields and introspection are coherent.
8. **Query context:** add stable pagination/detail navigation state.
9. **Profiles:** define EDM/profile behavior as optional enrichment.
10. **V2 migration table:** map old V2 features to the locked semantic concepts.

## Minimal Viable Semantic API

The MVP should prove the core without chasing V2 parity:

- `GET /api/semantic/v1/search`
- `GET /api/semantic/v1/resource/{id}`
- `GET /api/semantic/v1/introspect/classes`
- `GET /api/semantic/v1/introspect/fields`
- `GET /api/semantic/v1/docs`

MVP query capabilities:

- Text query.
- Exact field filter.
- Basic terms facet.
- Page/size pagination.
- One field sort.

MVP exclusions:

- Multi-hop paths.
- Collapse/grouping.
- More-like-this.
- Cross-index search.
- Deep facet cursoring.
- Geo.
- Full V2 parameter coverage.

## Chronology To Reconcile

The existing documents appear to represent these design eras:

- **October 2025 clean-break design:** self-documenting Hydra API, declarative server-side config, clean architecture.
- **October 2025 V2 adapter phase:** reuse V2 search and `itemFormat=semantic` as a bridge.
- **October 2025 technical expansion:** path navigation, Redis cursors, PIT, facet exploration.
- **February 2026 core refinement:** structure vs content, introspection as bridge, request-driven config, backend agnostic design.
- **February 2026 migration phase:** semantic API becomes public successor, V2 parity and ES8 backend work begin.
- **March 2026 migration docs:** V2 feature freeze and mapping table based on the implementation state.

The February 2026 core refinement should be treated as the strongest conceptual base. The migration and ES8 documents are useful implementation plans, but they should not redefine the core.

## Next Work Items

1. Review each core concept above in order and write a locked decision section for it.
2. Update or replace older docs once decisions are locked.
3. Derive a fresh, smaller architecture diagram from the locked concepts.
4. Only then return to V2 docs and create a migration table from V2 behavior to semantic concepts.
