# The Semantic API in One Page

**What it is.** One stable, self-describing search API for hub3 content: JSON-LD in and out,
Hydra hypermedia, a deliberately small request surface. It wraps the production v2 search
today and is designed so the backend (native Elasticsearch, triple store) can be swapped later
without clients noticing. Full contract: `docs/plans/2026-07-02-semantic-v1-greenfield.md`.

## How it works

```
GET  /search?query=rembrandt&filter[dc_creator][eq]=Rembrandt      ─┐
POST /search {"query":{...},"filters":[...]}                        ─┤→  QueryOptions  →  SearchStore  →  Hydra envelope
                                                                     │    (one internal    (v2 today,      (items untouched,
             two encodings, one meaning                             ─┘     query model)     ES/triples      links + facets
                                                                                            tomorrow)       around them)
```

Three rules carry the whole design:

1. **One query model, two encodings.** GET params and the POST JSON body parse into the same
   `QueryOptions`. Every options value has a canonical URL encoding, and *all* links in every
   response (next page, facet apply/remove) are generated from it — so hypermedia links are
   always followable, whichever way you asked.
2. **The API owns the envelope, never the content.** Items are the `semantic` JSON-LD documents
   produced at indexing time, passed through byte-for-byte as `hydra:member`. The API adds
   totals, paging links, facets, and timing *around* them — it never edits them.
3. **The surface is closed.** Twelve operators, ~15 parameters, one filter shape. Anything
   else is rejected with a `hydra:Error` naming the offender. Small enough to memorize,
   honest by construction: if it's accepted, it works.

## Self-describing, in three layers

- **Every response explains itself.** Facet values carry `applyURL`/`removeURL`; active filters
  carry remove links; pagination is `hydra:view` links. A client can navigate the whole dataset
  by following links, without ever constructing a URL.
- **`/docs` cannot lie.** The `hydra:ApiDocumentation` is generated from the same tables the
  parser validates against — parameters and operators are documented if and only if they work.
- **Every term has a meaning.** Responses reference a *versioned* JSON-LD context
  (`/contexts/hub3/1.0/context.jsonld`) that maps every envelope key (`hub3:facets`,
  `hydra:member`, …) to a stable IRI. A JSON-LD-aware client can expand the response into
  plain RDF; a plain-JSON client can ignore it entirely.

## Configuration without parameter growth

The trick is that **field names are data, not API**. `filter[dc_creator][eq]=…` uses the same
generic `filter` parameter whether the field is `dc_creator`, `edm_dataProvider`, or a field
that gets indexed next year. Adding a new vocabulary, dataset, or field to hub3 requires **zero
API changes** — the field flows through to the backend as an opaque name. There is no
server-side field registry gating requests (the old design's mistake): unknown fields simply
match nothing.

Presentation is configured the same way: the item documents carry JSON-LD, so *how the data
reads* is changed by supplying a different `@context` version — not by adding response-shaping
parameters (`detailLevel`, `fields=` and friends are deliberately absent).

## How it grows without breaking anyone

- **Unknown params are rejected today** — so introducing `cursor=` in v1.1 is unambiguous:
  before, it was an error; after, it works. Nothing silently changed meaning. Additions are
  always safe; nothing existing is ever reinterpreted.
- **The context is versioned.** `…/contexts/hub3/1.0/` is frozen forever; new terms arrive in
  `1.1`. Clients pin the version they understand.
- **New backends slot in behind one interface** (`SearchStore`: `Search`, `GetByID`) and must
  pass the contract test suite — the tests *are* the API contract, and they are the acceptance
  gate that lets a triple store replace Elasticsearch without any client migration.

## Deliberately not in v1 (phase 2 candidates)

Field/schema discovery (introspection), typed search (`/type/{t}/…`), per-facet limits,
multi-sort, cursor pagination, result navigation tokens, geo filters, related-item includes.
Each was cut because the current backend can't execute it honestly — and each can return
additively under the rules above.
