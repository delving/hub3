# Retiring Nave: what is actually used, and how to prove a replacement matches

**Date:** 2026-09-24
**Status:** Plan — nothing built yet
**Related:**
[ES 9 migration and Nave retirement](2026-09-22-elasticsearch-9-and-nave-retirement.md),
[NixOS host for Brabant Cloud's Nave](../../../delvi.ng.code/nixops/docs/plans/2026-09-23-nave-brabantcloud-nixos.md)

## Why now

Bringing Nave up on NixOS meant reproducing, one by one, everything the 2021
host carried that no repository recorded: a Django patch living in
site-packages, an entry point excluded by .gitignore, a celery schedule that
had rotted into an import error, a compiler that had to be pinned along with
the libraries. Each was fixable. Together they say something the individual
fixes do not, which is that the cost of keeping this application alive is
mostly the cost of keeping 2016 alive.

Then the database made the same point more bluntly.

## What is actually alive

The production database is 1354 MB. Its contents, by age:

| table | rows | size | last written |
|---|---|---|---|
| `lod_rdfsubjectlookup` | 3,132,682 | **1264 MB** | 2016-06-06 |
| `mip_image` | 57,949 | 15 MB | 2017-09-14 |
| `mip_miprecord` | 26,703 | 43 MB | 2017-09-14 |
| `lod_cacheresource` | 4,494 | 7 MB | 2017-04-11 |
| `virtual_collection_virtualcollection` | **28** | 112 kB | 2026-03-05 |
| `diw_diwinstance` | **28** | 64 kB | 2019-09-05 |
| `auth_user` | 10 | 64 kB | login 2026-07-24 |

93% of the database is a lookup table that has not been written to since 2016,
and which the code derives from the records anyway. The MIP tables are a
single import from one day in September 2017 — still served, but frozen.

What is genuinely maintained is 28 virtual collections, 28 Instant Website
instances and 10 users: well under a megabyte, edited through the Django admin
as recently as March.

## What is actually requested

From 400,000 consecutive requests in the access log:

| path | requests | share |
|---|---|---|
| `/api/search` | 378,912 | **94.7%** |
| `/data/`, `/resource/`, `/page/` (LOD) | 8,838 | 2.2% |
| `/search/` (HTML search UI) | 3,413 | |
| `/vc/…` (virtual collections) | 3,403 | |
| `/gebouwen/query` | 1,319 | |
| `/api/oai-pmh` | 1,091 | |
| `/proxy/tm`, `/detail/foldout`, `/api/memorix` | ~1,450 | |
| `/mip-gallery/…` | tens | |
| `/admin/` | a handful | |

The v1 search API is not the main thing this application does. It is very
nearly the only thing.

## The two paths are one order

Two options have been on the table: port Nave to current Django and Python
while discarding what is unused, or replace its read paths with Go services
behind nginx.

They are not alternatives. The Go path covers `/api/search`, OAI-PMH and LOD
resolving — about 97% of requests — and hub3 already serves a v1 API. What
remains for Django is virtual collections and their admin, MIP, the HTML
search page, foldout and the TM proxy: small, and none of it volume-sensitive.

So: **Go first, then port.** Once Go serves the read paths, the Django
remainder no longer has to be fast or highly available. It becomes an
editorial application over well under a megabyte of live state plus MIP, and
porting it is a contained job rather than a rewrite. Doing it the other way
round means modernising the heavy part and then throwing it away.

On how much can go: the estimate was "over 50%", and that is conservative.
requirements pulls in wagtail (a complete CMS), puput (a blog), django-suit,
rosetta, import-export and oauth-toolkit. None of them appear in the traffic.

## The parity suite

This is the part that decides whether any of it is safe, and it exists because
the two v1 implementations already disagree.

### They already disagree

One real query, taken verbatim from the log, against both:

```
/api/search/v1/?format=json&sortBy=tib_notes&rows=3&start=0
  &hqf[]=tib_collection_facet:Museum Helmond
```

| | numFound | first three |
|---|---|---|
| Django | 2581 | `81-277`, `2001-009`, `2008-080-C` |
| Go | 2581 | `2024-066`, `80-501`, `80-608` |

Same shape, same count, different results on page one. The filter agrees; the
ordering does not. `sortBy` appears in 181,344 of 200,000 sampled requests, so
this is not an edge case — it is the common case.

A discrepancy of this kind is the dangerous kind: nothing errors, nothing logs,
and an Instant Website simply shows different objects than it did yesterday.

### Build it from real traffic, not from imagination

The access log holds years of real queries, including every parameter
combination consumers actually send. Those are the test cases. Invented ones
would cover what we think the API does.

Parameters in 200,000 sampled requests:

```
format 186,989   rows 184,622   sortBy 181,344   start 180,155
facet.field 22,368   mlt.filterkey 9,016   callback 6,494   lang 6,492
query 4,859   facetBoolType 4,467   page 4,213   qf 3,960   mlt.count 1,991
```

Note `mlt.*` and `facetBoolType`: more-like-this and facet boolean handling are
in active use and are exactly where two implementations are most likely to part
ways.

### Shape of the suite

1. **Extract.** Pull distinct `/api/search` query strings from the log,
   normalise volatile parameters (`_`, `callback`), and keep the distinct set
   weighted by frequency. Aim for a few thousand cases covering every parameter
   combination that occurs, not a sample of the most common one.

2. **Replay.** Send each case to both implementations and store both responses.
   Run it read-only against production for Django and against hub3 for Go; both
   are reads, and the index is shared.

3. **Compare, in layers.** A single pass/fail hides what matters:
   - `numFound` — does filtering agree?
   - the ordered list of `doc_id` — does ranking agree?
   - the set of `doc_id` — same results, different order?
   - facet names, then facet counts
   - the field set of the first item, then its values
   - pagination fields

   Report per layer. "Same results, different order" and "different results"
   are different problems with different fixes, and a suite that says only
   "mismatch" makes them look alike.

4. **Triage.** Not every difference is a defect. Some will be Django bugs worth
   leaving behind. The output is a list of differences with a decision against
   each, not a number to drive to zero.

5. **Keep it.** Once the Go implementation serves traffic, the same suite runs
   against the old host for as long as it stays up, and then against a recorded
   baseline. It is how a later change is shown not to have moved anything.

### What to be careful about

The index is shared and live. Replay must be read-only, and rate-limited
enough not to disturb the running site — 378,912 requests in the log took
real time; the suite should not replay them at once.

Results shift as records are re-indexed. Both implementations must be queried
close together, or the diff will report changes that are simply the index
moving underneath.

## Steps

1. Route inventory: every URL Django exposes, against what the log shows being
   requested. Turns "what can go" from an estimate into a list. Half a day, and
   the foundation for both tracks.
2. Build the parity suite and run it. Report the differences by layer.
3. Triage the differences into: fix in Go, accept, or leave behind
   deliberately.
4. Put the Go v1 in front of the real traffic for one consumer, with the suite
   watching.
5. Move OAI-PMH and LOD resolving the same way.
6. Only then port what is left of Django, using the route inventory to decide
   what comes along.

## Open questions

The `sortBy` difference needs a cause before it needs a fix. Django may be
sorting on something Go ignores, or the reverse; the answer decides whether
this is one bug or a class of them.

Whether MIP moves, stays, or is archived is a product question, not a
technical one. It serves a 2017 import through a Django app nobody has changed
in years, at a handful of requests a day.
