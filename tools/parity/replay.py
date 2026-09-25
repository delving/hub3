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
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor

PAD = "/api/search/v1/"


def haal(basis, query, timeout):
    url = basis.rstrip("/") + PAD + ("?" + query if query else "")
    # No fixed Accept header: a large share of real traffic asks for
    # format=jsonp, and sending "Accept: application/json" alongside it makes
    # Django answer 406 Not Acceptable. The format parameter decides; let it.
    req = urllib.request.Request(url, headers={"Accept": "*/*"})
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            status, soort, body = r.status, r.headers.get_content_type(), r.read()
    except urllib.error.HTTPError as e:
        # A 4xx/5xx is a result, not a failure to measure: if one side rejects
        # a query and the other answers it, that is exactly what we are here
        # to find.
        status, soort, body = e.code, (e.headers.get_content_type() if e.headers else ""), e.read()
    tekst = body.decode("utf-8", "replace")
    # Without a format parameter the v1 API answers with Django REST
    # Framework's browsable HTML -- three million requests in five years do
    # exactly that -- so a non-JSON body is normal traffic, not an error.
    # Comparing those pages byte for byte says nothing; status and content
    # type are what carry meaning.
    try:
        doc = ontpak(tekst)
    except ValueError:
        doc = None
    return {"status": status, "soort": soort, "doc": doc, "lengte": len(body)}


def ontpak(body):
    """jsonp is not JSON -- it arrives as callback({...}). Unwrap it, so the
    Instant Website's own traffic can be compared like everything else."""
    s = body.strip()
    if s.startswith("{") or s.startswith("["):
        return json.loads(s)
    haak = s.find("(")
    if haak > 0 and s.endswith((")", ");")):
        binnen = s[haak + 1 : s.rindex(")")]
        return json.loads(binnen)
    return json.loads(s)


def uitpakken(antwoord):
    """The few things worth comparing, pulled out of the v1 envelope."""
    doc = antwoord["doc"] or {}
    # An id= lookup answers with a detail envelope rather than a result list,
    # and the two implementations disagree about its shape: Django wraps it in
    # "result" and adds "layout", Go returns item/relatedItems at the top
    # level. So record the shape as its own thing instead of quietly reading
    # past it -- six million requests a year use this path.
    vorm = sorted(doc.keys())
    if "result" in doc:
        vorm = ["result:" + k for k in sorted((doc["result"] or {}).keys())]
    res = doc.get("result", doc) or {}
    pag = res.get("pagination", {}) or {}
    items = res.get("items", []) or []
    if not items and "item" in res:
        items = [{"item": res["item"]}]
    ids = [(i.get("item") or {}).get("doc_id") for i in items]
    # Fields per record, not per position: the two sides often return the same
    # records in a different order, and comparing the first item of each would
    # restate that ordering difference as a field difference.
    velden = {(i.get("item") or {}).get("doc_id"): (i.get("item") or {}).get("fields") or {} for i in items}
    facetten = [(f.get("name"), f.get("total")) for f in (res.get("facets") or [])]
    return {
        "status": antwoord["status"],
        "soort": antwoord["soort"],
        "json": antwoord["doc"] is not None,
        "vorm": vorm,
        "numFound": pag.get("numFound"),
        "ids": ids,
        "velden_per_id": velden,
        "facetnamen": [n for n, _ in facetten],
        "facettellingen": facetten,
        "paginering": {k: pag.get(k) for k in ("start", "rows", "hasNext", "lastPage")},
    }


def beide_json(f):
    """Layers below only mean something when both sides returned JSON."""
    return lambda a, b: None if not (a["json"] and b["json"]) else f(a, b)


def gedeeld(a, b):
    """The records both sides returned -- the only ones whose contents can be
    compared without the ordering difference getting in the way."""
    return [i for i in a["velden_per_id"] if i in b["velden_per_id"]]


def veldset_gelijk(a, b):
    ids = gedeeld(a, b)
    if not ids:
        return None
    return all(sorted(a["velden_per_id"][i]) == sorted(b["velden_per_id"][i]) for i in ids)


def veldwaarden_gelijk(a, b):
    ids = gedeeld(a, b)
    if not ids:
        return None
    return all(a["velden_per_id"][i] == b["velden_per_id"][i] for i in ids)


LAGEN = [
    # (naam, hoe te vergelijken) -- volgorde van grof naar fijn, want een
    # verschil in numFound verklaart alle lagen eronder en hoeft niet
    # nog eens als "andere resultaten" geteld te worden.
    ("status", lambda a, b: a["status"] == b["status"]),
    ("content_type", lambda a, b: a["soort"] == b["soort"]),
    ("omhulsel", beide_json(lambda a, b: a["vorm"] == b["vorm"])),
    ("numFound", beide_json(lambda a, b: a["numFound"] == b["numFound"])),
    ("volgorde", beide_json(lambda a, b: a["ids"] == b["ids"])),
    ("zelfde_set", beide_json(lambda a, b: sorted(x or "" for x in a["ids"]) == sorted(x or "" for x in b["ids"]))),
    ("facetnamen", beide_json(lambda a, b: a["facetnamen"] == b["facetnamen"])),
    ("facettellingen", beide_json(lambda a, b: a["facettellingen"] == b["facettellingen"])),
    ("veldset", beide_json(veldset_gelijk)),
    ("veldwaarden", beide_json(veldwaarden_gelijk)),
    ("paginering", beide_json(lambda a, b: a["paginering"] == b["paginering"])),
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
            oordeel = gelijk(a, b)
            uit[naam] = "n.v.t." if oordeel is None else ("gelijk" if oordeel else "verschilt")
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
            # The cache-buster says nothing about the answer, so it goes. The
            # callback must NOT: format=jsonp without one is a request no real
            # client sends, and the two implementations answer it differently
            # -- dropping it manufactured a content-type difference in 59 of
            # 72 cases in the first run. Only its value is normalised, so the
            # same query does not appear a hundred times over.
            delen = []
            for p in query.split("&"):
                if p.startswith("_="):
                    continue
                if p.startswith("callback="):
                    p = "callback=cb"
                delen.append(p)
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
        # Only cases the layer actually applied to: counting the HTML answers
        # among them would claim coverage the run does not have.
        vergeleken = sum(1 for _, r in resultaten if r.get(n) in ("gelijk", "verschilt"))
        print(f"{n:<18} {anders:>10} {vergeleken:>8}", file=sys.stderr)
    if fouten:
        print(f"\n{fouten} zaken gaven een fout aan één kant", file=sys.stderr)


if __name__ == "__main__":
    main()
