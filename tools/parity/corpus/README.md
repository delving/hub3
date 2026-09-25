# The v1 parity corpus

Taken from five years of nginx access log on `nave-prod-brabantcloud`
(`/var/log/hub3-brabantcloud/nginx.log` plus the `.bak` nobody rotated):
**161,303,245** requests to `/api/search`, carrying **3,475** distinct
combinations of query parameters.

These are the queries consumers actually send. Invented cases would only cover
what we think the API does.

## Files

| file | what it is |
|---|---|
| `corpus.tsv` | the replay set: request count, parameter signature, real URL |
| `parameters.txt` | every parameter name seen, with its count |
| `signatures.txt` | every parameter combination, with its count |

`corpus.tsv` is weighted rather than sampled flat, because the distribution is
extremely skewed — one combination is 61% of all traffic:

- every combination with **1,000 requests or more**, up to 5 examples each
  (565 cases) — complete coverage of real use
- the **10 to 999** band, 2 examples each (2,572 cases) — breadth
- 60 from the **under-10** tail, which is almost entirely malformed query
  strings from real clients: an unescaped `&` turning object text into a
  parameter name, `amp%3BsortBy`, a stray `*:*`. Both implementations have to
  treat these the same way and neither may fall over, so they belong here.

3,197 cases in total.

## Regenerating it

`v1-analyse.awk` (in this directory) streams the log and writes the three
files. One pass over 76 GB takes about 45 minutes; run it with `nice -n 19
ionice -c3` because that host still serves production traffic.

```
mawk -v UIT=/var/tmp/v1-analyse -f v1-analyse.awk \
  /var/log/hub3-brabantcloud/nginx.log /var/log/hub3-brabantcloud/nginx.log.bak
```

It counts parameter names, groups requests by their sorted set of parameter
names, and keeps five real URLs per group. Grouping rather than collecting
distinct URLs is what keeps it bounded: 161 million URLs will not fit, and
parameter coverage is what the suite needs.

## Two things the corpus already told us

**`sortBy` is on 120,864,742 requests**, three quarters of all traffic. The
known Django/Go ordering disagreement is therefore the common case, not an
edge case.

**Three million requests arrive without a `format` parameter** (signature
`query+rows+start`), and Django answers those with REST Framework's browsable
HTML. Whatever the Go implementation does there, it is a decision someone has
to make deliberately rather than discover.
