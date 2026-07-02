# Semantic API Current Scope

**Date:** 2026-06-04
**Status:** Current delivery guide

## Summary

Semantic V1 is the public API contract. For the current release it is implemented
as a JSON-LD/Hydra wrapper around the existing V2 search implementation.

The current goal is not to expose every V2 parameter under new names. The goal is
to provide a smaller, self-documenting API with stable semantic concepts:

- query
- filters and hidden/base filters
- facets
- sorting
- pagination
- resource detail
- includes
- introspection
- API documentation

Native Elasticsearch work is future development. It can replace the V2 adapter
only if it preserves the accepted Semantic V1 contract.

## Source Of Truth

| Document | State | Purpose |
|---|---|---|
| [ADR 0001](adr/0001-semantic-v1-wraps-v2.md) | Accepted | Current architecture decision and scope boundary |
| [semantic-api-reference.md](semantic-api-reference.md) | Current reference | Public endpoint and response reference |
| [v2-feature-freeze.md](v2-feature-freeze.md) | Current migration reference | V2 parameter freeze and mapping to Semantic V1 concepts |
| [v2-to-semantic-migration-guide.md](v2-to-semantic-migration-guide.md) | Current migration guide | Frontend migration guidance |
| [plans/2026-05-07-semantic-api-core-reset.md](plans/2026-05-07-semantic-api-core-reset.md) | Superseded design note | Useful concept inventory; current scope is narrowed by ADR 0001 |
| [plans/2026-02-23-phase-c-native-es-backend.md](plans/2026-02-23-phase-c-native-es-backend.md) | Future plan | Native ES backend notes, not current public scope |

## Current Implementation Boundary

The configured Semantic V1 service uses `ikuzo/storage/x/v2adapter`.

The adapter:

- translates Semantic V1 query options to V2 request parameters;
- uses V2 semantic item output for JSON-LD content;
- translates V2 result totals, members, facets, and detail documents into the Semantic V1 response envelope.

The public API does not support selecting `v2` or `es8` per request. Backend choice
is an implementation concern.

## Done Means

For the current release, Semantic V1 is done when:

- documented GET and POST search flows work through the V2 adapter;
- detail lookup by hub ID and URI works;
- available fields, facets, and operations are discoverable through docs and introspection;
- V2 migration mappings identify supported, dropped, and future behavior;
- native ES/backend-switching behavior is not required for client migration.

## Future Work

Future work may include:

- replacing the V2 adapter with a native backend;
- graph path querying;
- deeper facet exploration;
- backend-specific performance improvements;
- removing V2 routes after clients have migrated.

Future work must either preserve the Semantic V1 contract or introduce a new ADR.
