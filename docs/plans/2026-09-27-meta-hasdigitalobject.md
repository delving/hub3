# A media flag in the meta block

**Date:** 2026-09-27
**Status:** Plan — nothing built yet
**Tickets:** [#3052](https://delving.plan.io/issues/3052) (the request),
[#3598](https://delving.plan.io/issues/3598) (`meta.sourceModified`, the same
pattern in the same message)
**Related:** [parity findings](../../tools/parity/bevindingen-2026-09-25.md)

## What is being asked

Erfgoed Brabant Verhalen had a "met media" filter on their old site and wants
it back in the Instant Website. Other institutions on the Instant Website
would get it too.

Measured on the production index: 1,709,224 records carry a digital object and
338,967 do not — 83.4% against 16.6%. So the distinction is worth filtering
on, and the data to make it already exists.

What is **not** being built here is a filter per media type. Cris's ticket says
"afbeelding, pdf, audio, video"; we record only *that* there is an object.
`mimeType` exists in the protobuf (`Header`, field 12) but is not populated for
EDM records, so type-level filtering is a separate and larger job. The reply on
#3052 says so and asks whether met/zonder is enough.

## Where the flag lives today

v1 computes it while indexing, as a presence check —
`hub3/fragments/v1.go:122`:

```go
_, ok = indexDoc["edm_isShownBy"]
l.HasDigitalObject = strconv.FormatBool(ok)
```

with three siblings on the same pattern: `HasGeoHash` (`nave_geoHash`),
`HasDeepZoom` (`nave_deepZoomUrl`), `HasLandingPage` (`edm_isShownAt`). It
lands in the legacy block as a **string**, which is why Django's
`query_string` against `legacy.delving_hasDigitalObject` with `"true"` works.

v2 does not compute it. The name does appear in the v2 mapping —
`ikuzo/driver/elasticsearch/internal/mapping/v2.go:149` — but inside the tree
block, among `daoLink`, `manifestLink` and `genreform`: that is the EAD archive
tree's own flag, not a record-level one. Useful as precedent that a boolean
belongs there, not as something to reuse.

## Why the meta block

Measured, both sides, for `qf[]=delving_hasDigitalObject:true`:

```
django  post_filter  query_string on  legacy.delving_hasDigitalObject   1,709,224
go      post_filter  nested on resources.entries
                       (searchLabel=delving_hasDigitalObject
                        AND @value.keyword="true")                              0
```

Go looks for it among the nested RDF triples, and a field we derive is never
there — it exists in `legacy.*` and `fields.*` only. Hence zero hits, in v1 and
in v2 alike. The Instant Website's filter works today solely because it goes
through Django.

The meta block is where our own conclusions about a record already live —
`docType`, `spec`, `revision`, `modified` — and Go queries those **flat**:
`{"term": {"meta.orgID": "brabantcloud"}}` sits in every base query it builds.
Putting the flag there means the existing filter path carries it with no
changes. That is the whole argument: not tidiness, but that it lands in a
pattern that already works.

## Steps

1. **Protobuf.** Add `hasDigitalObject` as a bool on the `Header` message. That
   message runs to field 15, so 16 and 17 are free; #3598 wants one too, so
   allocate both in one pass rather than racing for 16. `make protobuffer`
   regenerates. Backwards compatible: an older reader ignores it.

2. **Compute it during v2 indexing**, next to where the rest of the meta block
   is filled. Same presence check as v1, so the two agree during the period
   both are served — but as a real boolean, not a `FormatBool` string.

   Take the assumption out of the copy while moving it: v1 checks
   `edm_isShownBy` alone, so a record exposing its object only through
   `edm_object` or `edm_hasView` counts as having no media. In a 400-record
   sample across 8 collections those three moved together (all at 77%), so it
   is probably harmless — but write the condition to say what it means rather
   than inheriting a line whose reasoning is lost.

3. **v2 mapping.** One line under `meta`: `"hasDigitalObject": {"type":
   "boolean"}`. Adding a field to an existing mapping needs no reindex by
   itself; step 5 does.

4. **Filtering needs nothing.** `qf=meta.hasDigitalObject:true` should work as
   soon as documents carry it, because `meta.*` is already queried flat. Verify
   rather than assume — that is one request with `echo=searchService`.

5. **Reindex.** The flag appears only on records indexed after the change, so
   the whole of Brabant Cloud has to be rebuilt before the filter is
   trustworthy. This, not the code, is the timeline for the Deelnemersdag at
   the end of November.

6. **Facet.** A terms aggregation over a boolean works, unlike the current
   situation: asking v1 for `facet.field=delving_hasDigitalObject` returns the
   facet in the list with no values at all, on both sides. Check that the
   Instant Website can render a two-value facet before promising the UI.

## Decisions that are not ours to make alone

- **Do the three siblings come along?** `hasDeepZoom`, `hasLandingPage` and
  `hasGeoHash` are the same shape and the same one-line cost. `hasLandingPage`
  is true on effectively every record, so as a filter it is useless and as a
  trap it is live: it looks like the media flag and is not. Either bring them
  over deliberately or leave them in v1 and say why.

- **Does `delving_hasDigitalObject` stay too?** Keeping both means two sources
  of the same truth. The v1 legacy field cannot go while consumers read it, so
  the honest answer is probably "both, and the v1 one is deprecated on the day
  the Go v1 serves traffic".

## Verifying it

The parity corpus is the check: `tools/parity/replay.py` already has cases with
`qf[]` filters, and a run before and after should move the numFound layer
without touching the others. Add a handful of explicit
`meta.hasDigitalObject` cases to the corpus, since no real traffic contains
them yet.
