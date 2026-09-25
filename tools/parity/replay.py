#!/usr/bin/env python3
"""Replay real v1 search queries against two implementations and diff by layer.

The corpus comes from the access log, so these are queries consumers actually
send rather than ones we imagined. Both sides are read-only and share the same
index, so a run changes nothing.

A single pass/fail would hide what matters: "same results, different order" and
"different results" are different problems with different fixes. So every case
is reported per layer, and the summary counts disagreements per layer.

Usage:
    replay.py --corpus voorbeelden.tsv \
              --a https://data.brabantcloud.nl \
              --b http://localhost:3000 \
              --out rapport.tsv
"""

import argparse
import json
import sys
import time
import urllib.request
from concurrent.futures import ThreadPoolExecutor

PAD = "/api/search/v1/"


def haal(basis, query, timeout):
    url = basis.rstrip("/") + PAD + ("?" + query if query else "")
    req = urllib.request.Request(url, headers={"Accept": "application/json"})
    with urllib.request.urlopen(req, timeout=timeout) as r:
        return json.loads(r.read().decode("utf-8", "replace"))


def uitpakken(doc):
    """The few things worth comparing, pulled out of the v1 envelope."""
    res = (doc or {}).get("result", {})
    pag = res.get("pagination", {}) or {}
    items = res.get("items", []) or []
    ids = [(i.get("item") or {}).get("doc_id") for i in items]
    eerste = (items[0].get("item") or {}).get("fields", {}) if items else {}
    facetten = [(f.get("name"), f.get("total")) for f in (res.get("facets") or [])]
    return {
        "numFound": pag.get("numFound"),
        "ids": ids,
        "velden": sorted(eerste.keys()),
        "eerste_waarden": {k: eerste[k] for k in sorted(eerste)},
        "facetnamen": [n for n, _ in facetten],
        "facettellingen": facetten,
        "paginering": {k: pag.get(k) for k in ("start", "rows", "hasNext", "lastPage")},
    }


LAGEN = [
    # (naam, hoe te vergelijken) -- volgorde van grof naar fijn, want een
    # verschil in numFound verklaart alle lagen eronder en hoeft niet
    # nog eens als "andere resultaten" geteld te worden.
    ("numFound", lambda a, b: a["numFound"] == b["numFound"]),
    ("volgorde", lambda a, b: a["ids"] == b["ids"]),
    ("zelfde_set", lambda a, b: sorted(x or "" for x in a["ids"]) == sorted(x or "" for x in b["ids"])),
    ("facetnamen", lambda a, b: a["facetnamen"] == b["facetnamen"]),
    ("facettellingen", lambda a, b: a["facettellingen"] == b["facettellingen"]),
    ("veldset", lambda a, b: a["velden"] == b["velden"]),
    ("veldwaarden", lambda a, b: a["eerste_waarden"] == b["eerste_waarden"]),
    ("paginering", lambda a, b: a["paginering"] == b["paginering"]),
]


def vergelijk(query, args):
    try:
        a = uitpakken(haal(args.a, query, args.timeout))
    except Exception as e:  # noqa: BLE001 -- an error on either side is a result
        return query, {"fout": f"A: {type(e).__name__}: {e}"}
    try:
        b = uitpakken(haal(args.b, query, args.timeout))
    except Exception as e:  # noqa: BLE001
        return query, {"fout": f"B: {type(e).__name__}: {e}"}

    uit = {}
    for naam, gelijk in LAGEN:
        try:
            uit[naam] = "gelijk" if gelijk(a, b) else "verschilt"
        except Exception:  # noqa: BLE001
            uit[naam] = "onvergelijkbaar"
    uit["numFound_a"] = a["numFound"]
    uit["numFound_b"] = b["numFound"]
    return query, uit


def corpus_lezen(pad, limiet):
    """Accepts a bare query string per line, a full URL, or the analysis TSV
    (count, signature, url) -- the url is always the last tab-separated field."""
    gezien, uit = set(), []
    with open(pad, encoding="utf-8") as f:
        for regel in f:
            regel = regel.strip()
            if not regel or regel.startswith("#"):
                continue
            veld = regel.split("\t")[-1]
            query = veld.split("?", 1)[1] if "?" in veld else veld
            # De cache-buster en de jsonp-callback zeggen niets over het
            # antwoord en zouden het corpus alleen opblazen.
            delen = [p for p in query.split("&") if not p.startswith(("_=", "callback="))]
            query = "&".join(delen)
            if query in gezien:
                continue
            gezien.add(query)
            uit.append(query)
            if limiet and len(uit) >= limiet:
                break
    return uit


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--corpus", required=True)
    p.add_argument("--a", required=True, help="reference implementation (Django)")
    p.add_argument("--b", required=True, help="candidate (Go)")
    p.add_argument("--out", default="rapport.tsv")
    p.add_argument("--limit", type=int, default=0)
    p.add_argument("--timeout", type=float, default=30.0)
    # The index is shared and live: replaying flat out would disturb the site
    # the queries came from, and both sides must be asked close together or
    # re-indexing shows up as a difference that is really just time passing.
    p.add_argument("--workers", type=int, default=4)
    p.add_argument("--rate", type=float, default=0.1, help="seconds between cases")
    args = p.parse_args()

    zaken = corpus_lezen(args.corpus, args.limit)
    print(f"{len(zaken)} unieke queries", file=sys.stderr)

    resultaten = []
    with ThreadPoolExecutor(max_workers=args.workers) as pool:
        futures = []
        for q in zaken:
            futures.append(pool.submit(vergelijk, q, args))
            time.sleep(args.rate)
        for i, fut in enumerate(futures, 1):
            resultaten.append(fut.result())
            if i % 50 == 0:
                print(f"  {i}/{len(zaken)}", file=sys.stderr)

    lagen = [n for n, _ in LAGEN]
    with open(args.out, "w", encoding="utf-8") as f:
        f.write("\t".join(["query", "numFound_a", "numFound_b", *lagen, "fout"]) + "\n")
        for q, r in resultaten:
            rij = [q, str(r.get("numFound_a", "")), str(r.get("numFound_b", ""))]
            rij += [r.get(n, "") for n in lagen]
            rij.append(r.get("fout", ""))
            f.write("\t".join(rij) + "\n")

    print(f"\nrapport: {args.out}\n", file=sys.stderr)
    fouten = sum(1 for _, r in resultaten if "fout" in r)
    print(f"{'laag':<18} {'verschilt':>10} {'van':>8}", file=sys.stderr)
    for n in lagen:
        anders = sum(1 for _, r in resultaten if r.get(n) == "verschilt")
        vergeleken = sum(1 for _, r in resultaten if n in r)
        print(f"{n:<18} {anders:>10} {vergeleken:>8}", file=sys.stderr)
    if fouten:
        print(f"\n{fouten} zaken gaven een fout aan één kant", file=sys.stderr)


if __name__ == "__main__":
    main()
