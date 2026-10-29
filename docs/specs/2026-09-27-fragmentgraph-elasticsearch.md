# The FragmentGraph in Elasticsearch, as it stands

**Date:** 2026-09-27
**Status:** Reverse-engineered description. No judgement, no proposals.
**Read from:** `hub3/fragments/` (graph.go, resource.go, api.go, v1.go),
`ikuzo/driver/elasticsearch/internal/mapping/v2.go`, `ikuzo/rdf/index/`

This records what the design is, so that the next person changing it does not
have to rediscover it from the code. Where the code and the mapping disagree,
that is written down too — not fixed here.

The design's own goal, in the words of `ikuzo/rdf/index/docs.go`: *"a
high-fidelity means to index, search and process rdf information."* Everything
below is in service of two things that pull against each other — keeping a
graph's precision, and still being able to search and aggregate it.

## The unit

One Elasticsearch document per RDF named graph. `meta.hubID` is the document
id; `meta.namedGraphURI` and `meta.entryURI` name the graph and its subject.

A graph is not stored as triples. It is **grouped by subject**: each distinct
subject in the graph becomes one entry in `resources`, and each triple with
that subject becomes one entry in that resource's `entries`. So the document is
a graph flattened one level, subject-major.

## The document

| property | type | what it is |
|---|---|---|
| `meta` | object, flat | what we know *about* the record — see below |
| `resources` | **nested** | the graph, grouped by subject |
| `fields` | object, dynamic | flat projection for search and aggregation |
| `full_text` | text | everything searchable, filled by `copy_to` |
| `tree` | object | hierarchy for EAD/archival records |
| `protobuf` | object | the whole graph, stored, not indexed |
| `recordType` | short | |
| `_checksum` | keyword | hash of the graph, excluding modified/revision |

`meta` is deliberately root-level and flat, which is why it is queried with a
plain term query while everything in `resources` needs a nested query. Its
fields: `orgID`, `spec`, `hubID`, `revision`, `tags`, `docType`,
`namedGraphURI`, `entryURI`, `modified`, `sourceID`, `sourcePath`, `groupID`,
`recDefID`, `aboutTypeURI`, `hasDigitalObject`.

## resources — the graph, grouped by subject

`resources` is `nested`, so each subject can be queried as a unit without its
values bleeding into a sibling's. Per resource:

- `id` — the subject IRI, or the blank node's identifier
- `types` — the subject's `rdf:type` values, as IRIs
- `entries` — nested, one per triple
- `context` — nested, how this resource is reached (below)
- `graphExternalContext` — referrers from outside this graph

## entries — where the precision lives

Each entry is one triple's object, with everything needed to reconstruct it:

| field | holds |
|---|---|
| `@id` | the object's IRI, when the object is a resource or blank node |
| `@value` | the literal, or the resolved label of a linked resource |
| `@language` | the literal's language tag |
| `@type` | the literal's datatype |
| `entrytype` | literal, resource or bnode — which of the two above to trust |
| `predicate` | the predicate IRI, unabbreviated |
| `searchLabel` | the predicate in namespaced form (`dc_title`), from the namespace manager |
| `level` | depth from the root resource |
| `order` | the triple's position in the source graph |

`entrytype` is the field that makes the rest unambiguous: `@id` and `@value`
can both be set — a linked resource that also carries a label — and the type
says which one is the object.

Two of these matter more than they look:

**`order`** exists because JSON and Elasticsearch do not preserve triple order
and RDF does not guarantee it, yet a record's fields have a presentation order
the source intended. It survives serialisation so the original sequence can be
restored.

**`searchLabel`** is the same abbreviation the v1 indexer derives through
`GetFieldKey` → `rdf.DefaultNamespaceManager.GetSearchLabel`. Both indexers
consult the same predicate table, which is what lets a v1 and a v2 document be
asked the same question.

### Typed projections

Alongside the string value, an entry can carry the same object parsed into a
type Elasticsearch can range-query: `integer`, `float`, `isoDate`,
`dateRange`, `intRange`, `latLong`. These are additions, not replacements —
`@value` keeps the original lexical form. In `ikuzo/rdf/index` they are grouped
in an embedded `TypeIndexField`.

### Label resolution

`resolvedFrom` and `resolvedLevel` record that an entry's `@value` is not the
triple's own object but a label fetched from the resource it points at, and how
many hops away it was found. `resolvedLevel` 0 means not resolved, 1 a direct
label, 2 or more a label from a nested resource.

### Facet grouping across predicates

`mFilterID`, `mType` and `mRole` (embedded as `CustomFilterField`) let entries
with different search labels be aggregated together — a facet over "creator"
regardless of which predicate expressed it.

### sortValue

An alternate sort key for facet buckets: the bucket keeps `@value` as its
visible label while ordering on `sortValue` (surname-first, for instance). It
is set by the dlod indexer for search labels that opted in, and is carried next
to `@value` so that copying a parent's entries into a child record keeps it.

## context — the paths

`context` is what makes this more than a flat projection. Each `ContextRef`
records one hop by which a resource is reached:

| field | holds |
|---|---|
| `Subject` | the referring subject |
| `SubjectClass` | that subject's rdf types |
| `Predicate` / `SearchLabel` | the predicate that points here |
| `Level` | the depth of that hop |
| `ObjectID` | the object reached |
| `SortKey`, `Label` | ordering and display |

An ordered list of these is the path from the root resource to this one. Being
`nested` inside a `nested` resource, a query can ask for a resource reached by
a given predicate from a given class at a given level, which is how a query
addresses graph structure rather than just field values:

```go
classQuery := elastic.NewTermQuery("resources.context.SubjectClass", tc)
labelQ := elastic.NewTermQuery("resources.context.SearchLabel", level2.SearchLabel)
lq := elastic.NewNestedQuery("resources.context", levelq.Must(labelQ))
```

`containsContext` deduplicates these, because a resource reachable by two paths
would otherwise be counted twice in an aggregation.

## fields — the flat projection

`fields` is a map of `searchLabel` → values, built by `GenerateFields`. It
exists for speed: aggregating and filtering through a nested query is
expensive, and most searching does not need structure.

It is **lossy on purpose**, and in one way that surprises:

```go
// Skip non-literal entries
if entry.EntryType != literal { continue }
```

Only literals are projected. A predicate whose object is an IRI —
`edm_isShownBy`, `edm_object`, `edm_isShownAt` — is in `resources.entries` and
never in `fields`, however central it is to the record. Values are
deduplicated per search label.

The dynamic template maps `fields.*` as `text` with a `keyword` subfield capped
at 256 characters, which is what makes `fields.<label>.keyword` available for
aggregation without declaring every predicate in advance.

## full_text and protobuf

`resources.entries.@value` is `copy_to: full_text`, so every literal in the
graph lands in one analysed field. The index's `query.default_field` is
`full_text`, which is what a query without a field searches.

`protobuf.data` holds the serialised graph: `store: true`, `index: false`,
`doc_values: false`. Elasticsearch keeps it and never looks inside it. This is
the lossless copy — whatever the projections above drop, this still has.

## Index settings that shape the design

```
mapping.nested_objects.limit   50000
mapping.nested_fields.limit       50
mapping.depth.limit               20
mapping.total_fields.limit      1000
date_detection                 false
query.default_field        full_text
```

The nested-objects limit is the real ceiling on graph size: every resource and
every entry is a nested document, so a large graph is thousands of them in one
Elasticsearch document. `date_detection: false` matters because dynamic
`fields.*` values must not be guessed into dates; dates are explicit, in
`isoDate` and `dateRange`.

The default analyzer strips HTML and applies lowercase and asciifolding.

## Which paths are actually queried

From the query builders in `hub3/fragments/`:

```
resources.entries.searchLabel        resources.entries.@value
resources.entries.@value.keyword     resources.entries.@id
resources.entries.isoDate            resources.entries.dateRange
resources.entries.integer            resources.entries.tags
resources.entries.sortValue.keyword
resources.context.SearchLabel        resources.context.SubjectClass
```

Filtering on a field value is a nested query on `resources.entries` matching
`searchLabel` and then the value. Sorting on a field is the same shape as a
nested sort. A `meta.` field skips all of that and is a plain term query.

## The second attempt: ikuzo/rdf/index

`ikuzo/rdf/index` is a cleaner restatement of the same model, used as the
datamodel of the v3 search API. The shapes correspond:

| hub3/fragments | ikuzo/rdf/index |
|---|---|
| `FragmentGraph` | `Graph` |
| `Header` (protobuf) | `Header` (plain struct) |
| `FragmentResource` | `Resource` |
| `ResourceEntry` | `Entry` |
| `FragmentReferrerContext` | `ContextRef` |

What changed beyond naming:

- The types are plain Go structs with documented fields instead of generated
  protobuf messages. The JSON shape is kept, so documents stay readable by both.
- The typed projections and the facet-grouping fields are grouped into embedded
  structs (`TypeIndexField`, `CustomFilterField`) instead of sitting loose among
  twenty other fields.
- `EntryType` is a named type rather than a bare string.
- `Entry` carries a private `fingerprint` for deduplication.
- `Embed []embed.Raw` allows structs to be carried in the graph, marked
  `json:"-"` so it never reaches the index.
- `Tree`, `predicates` and `objectIDs` are commented out rather than carried —
  the archival hierarchy has not been brought across.

## Where code and mapping disagree

Found while reading. Recorded, not fixed.

**`_checksum` can never arrive.** `index.Graph` declares
`json:"_ checksum,omitempty"` — with a space. The mapping declares
`_checksum`. Whatever the v3 model writes, it does not land in the declared
property.

**`sortValue` and `inline` are not in the mapping.** Both are entry fields, so
they are mapped dynamically. A query already depends on
`resources.entries.sortValue.keyword`, a subfield that exists only because
dynamic mapping gives strings a `.keyword` — not because anything declared it.
`inline` is documented as never indexed but is not excluded from indexing
either; it is a whole nested resource inside an entry, against a nested-objects
limit.

**Entry fields the old struct lacks.** `resolvedFrom`, `resolvedLevel`,
`mType`, `mRole` and `mFilterID` are in the mapping and in
`ikuzo/rdf/index.Entry`, but not in `hub3/fragments.ResourceEntry`. The mapping
was extended for the newer model while the older writer stayed as it was.

**`tree.hasDigitalObject` is not the record's.** The `tree` block has its own
`hasDigitalObject` and `mimeType`, for the archival hierarchy. The record-level
flag is `meta.hasDigitalObject` (#3052). Same name, different subject.

**`meta.tags` and `resources.entries.tags` are different things.** Graph-level
tags against entry-level tags, queried through different paths.
