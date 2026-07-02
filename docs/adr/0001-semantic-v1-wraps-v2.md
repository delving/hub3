# ADR 0001: Semantic V1 Delivers a JSON-LD Wrapper Around V2 Search

**Date:** 2026-06-04
**Status:** Accepted

## Context

The semantic API work grew in two directions at the same time:

- a JSON-LD/Hydra API that wraps the existing V2 search behavior with a cleaner, self-documenting contract;
- a native Elasticsearch implementation that queries the V2 index structure directly.

The customer scope is now clarified. The current delivery is not a new search engine and not a complete replacement query contract. The agreed scope is to wrap existing V2 search behavior in a better designed JSON-LD API that is easier to extend and easier for clients to discover.

The current V2 API has more than 50 URL parameters. Semantic V1 should not copy that shape. It should expose stable concepts such as query, filters, facets, sorting, pagination, resource detail, includes, and introspection. The API should document those capabilities through JSON-LD/Hydra responses so clients can discover available operations and options without relying only on prose documentation.

## Decision

Semantic V1 is delivered as a wrapper around the existing V2 search implementation.

The V2 adapter is the current production backend for `/api/semantic/v1`. It translates Semantic V1 requests into V2 search requests, always asks V2 for semantic item output, and returns JSON-LD/Hydra responses.

Native Elasticsearch querying is future development. It may remain in the repository as experimental or preparatory code, but it is not part of the current public API contract and must not be exposed through request parameters or migration documentation for the current release.

The public contract is backend-neutral at the response and documentation level, but implementation-neutral does not mean runtime backend switching is part of the API. Clients should not choose `v2` or `es8`; they should consume Semantic V1.

## Design Rules

1. Semantic V1 must expose JSON-LD responses only. V2 format choices such as `itemFormat` and `format` stay behind the adapter.
2. GET requests support common query forms. POST requests carry structured JSON-LD query documents for extensibility.
3. The API is self-documenting. `/api/semantic/v1/`, `/api/semantic/v1/docs`, type documentation, and introspection endpoints are part of the product, not optional extras.
4. Public parameters describe semantic concepts, not V2 implementation names. New public parameters require a contract decision, not another one-off V2 passthrough.
5. V2 compatibility is adapter behavior. It should preserve useful existing behavior without making V2's parameter set the Semantic V1 design.
6. Backend-specific query builders, native ES experiments, and index details stay below the store interface.
7. Future native backend work must implement the same Semantic V1 contract before it can replace the adapter.

## Current Scope

In scope for the current release:

- `GET /api/semantic/v1/search`
- `POST /api/semantic/v1/search`
- `GET /api/semantic/v1/resource/{id}`
- JSON-LD/Hydra collection and resource responses
- text query, property filters, hidden/base filters, facets, sort, pagination
- detail includes that wrap existing V2-backed capabilities, such as related items when supported
- API documentation and introspection endpoints that describe available operations, fields, filters, facets, and types
- migration guidance from V2 parameters to Semantic V1 concepts

Out of scope for the current release:

- public backend switching, including `backend=v2` or `backend=es8`
- native Elasticsearch as a customer-visible backend
- exposing V2-only format switches
- tree/EAD-specific V2 parameters
- new graph path query language
- deep facet exploration unless it is already supported through the wrapper contract
- replacing the V2 index/query implementation

## Consequences

The current implementation should wire the V2 adapter as the only production Semantic V1 backend.

Documents that describe native ES backend work should be treated as future-development plans. They are useful implementation notes, but they do not define the current customer-facing API.

When the native ES backend is revisited, the acceptance bar is contract compatibility with Semantic V1, not feature divergence. If a native backend needs a different public contract, that requires a new ADR.

## Verification Questions

Use these questions when reviewing future changes:

- Does this change improve the Semantic V1 contract, or does it merely expose another V2 parameter?
- Can a client discover the capability from the API documentation or introspection endpoints?
- Is this behavior available through the V2 wrapper now, or is it future native-backend work?
- Does this require clients to know which backend handled the request?
- Would the same request and response shape still make sense if the backend later changes?
