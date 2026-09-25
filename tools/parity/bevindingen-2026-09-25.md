# First parity run: Django v1 against Go v1

**Date:** 2026-09-25
**A:** `https://data.brabantcloud.nl` (Nave, Django)
**B:** `https://api.brabantcloud.nl` (hub3, Go)
**Corpus:** 2,827 unique cases from `corpus/corpus.tsv`
**Raw report:** `rapport-2026-09-25.tsv`

Both sides read the same index, so a difference is implementation, not data.

## Result, weighted by real traffic

The case count says how much of the corpus disagrees; the traffic share says
how much of the 161 million real requests sits behind those cases. They differ
sharply, and the second column is the one to plan from.

| layer | cases differ | of | share of traffic |
|---|---|---|---|
| status | 14 | 2827 | 0.1% |
| content_type | 158 | 2827 | 5.0% |
| envelope | 1583 | 2684 | 8.5% |
| numFound | 544 | 2684 | 2.4% |
| **order** | 1642 | 2684 | **57.4%** |
| **same set** | 1504 | 2684 | **52.5%** |
| facet names | 2362 | 2684 | 21.2% |
| **facet counts** | 2588 | 2684 | **92.9%** |
| **field set** | 536 | 536 | **100%** |
| field values | 536 | 536 | 100% |
| pagination | 1452 | 2684 | 2.5% |

Two readings worth keeping apart. `numFound` and `pagination` disagree in a
fifth to a half of the *cases* but almost none of the *traffic*: filtering and
paging are sound for the queries consumers actually send, and break on the
rare ones. Facet counts are the opposite — they disagree on nearly everything.

## Where the differences come from

Breakage by parameter, over cases containing each (percentage of applicable
cases that disagree):

| parameter | cases | numFound | order | facet counts | envelope |
|---|---|---|---|---|---|
| `q` | 67 | 41% | 97% | 100% | 21% |
| `query` | 2515 | 21% | 58% | 99% | 61% |
| `qf[]` | 2191 | 24% | 56% | 100% | 65% |
| `sortBy` | 697 | 14% | 55% | 100% | 57% |
| `facet.field` | 2427 | 22% | 58% | 100% | 61% |
| `mlt` | 124 | 0% | 100% | 39% | 61% |
| `id` | 102 | 0% | 100% | 1% | **99%** |
| `itemFormat` | 44 | 50% | 100% | 100% | 0% |

### 1. Facet counts, everywhere (93% of traffic)

Wherever facets are requested the counts differ. This is the largest gap by
traffic and worth a cause before anything else is touched, because facets
drive every Instant Website's filter UI: wrong counts are visible to end users
on every page.

### 2. Ordering (57%) and the result set itself (52.5%)

The known `sortBy` disagreement, confirmed at scale. Note that "same set" also
differs in half the traffic: this is not only a different order of the same
records but different records on page one. A consumer would show different
objects than yesterday, with nothing in any log to say why.

### 3. The detail path is structurally incompatible (`id=`, 6.8M requests)

Django answers an `id=` lookup with `{"result": {layout, item, relatedItems}}`;
Go answers with `{item, relatedItems}` — no `result` wrapper and no `layout`.
Any consumer reading `result.item` breaks, and anything rendering field labels
from `layout` loses them. Counts and facets on this path agree; it is the
envelope alone.

### 4. `q` is a different language on each side

495 cases return nothing from Go while Django returns results, and 13 the
other way. The Lucene-ish syntax goes through Django and not through Go:

```
format=json&q=delving_spec:museum-helmond-vervaardigers AND dc_title:"Albert Neuhuys"
  django: results        go: 0
```

And the reverse, where Django is the odd one:

```
q=*:*&page=1&rows=15&format=json&itemFormat=semantic&facetBoolType=OR
  django: 0             go: 40323
```

`q` carries 2,766,166 requests, so this needs a decision rather than a fix:
which of the two behaviours is the one we mean.

### 5. One field each way, on every record

On the same `doc_id`, Django returns `delving_collection` and Go returns
`delving_landingpage`; the other 35 fields match. Systematic and small, but it
is why the field-set layer is 100%.

### 6. Malformed input diverges

```
format=json&amp;query="dolor"    django: 6    go: 2048172 (everything)
```

A literal `&amp;` makes the parameter name `amp;query`. Django still narrows
the result; Go ignores it and returns the whole index. Both behaviours are
defensible; they must not be different, because these requests are real.

## Also seen, not yet quantified

Django's `query` block carries `breadCrumbs`; Go's `query` block is empty.
Consumers that render a breadcrumb trail from it have nothing to render.

## What this run does not tell us

Which side is right. Every difference above is a question for triage, not a
defect list: some will be Django behaviour worth leaving behind, and the `q`
and malformed-input cases are almost certainly among them. The next step is a
decision per row, not driving a number to zero.

## Repeating it

```
python3 replay.py --corpus corpus/corpus.tsv \
  --a https://data.brabantcloud.nl --b https://api.brabantcloud.nl \
  --out rapport.tsv --rate 0.15 --workers 3
```

Twelve minutes, read-only, rate-limited so the live site is undisturbed. Run
both sides close together: the index moves, and a slow run reports that
movement as a difference.
