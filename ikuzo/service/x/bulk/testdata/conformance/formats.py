"""Write formats/NNNN.{ttl,rdf,jsonld} from NNNN.nt.

usage: formats.py <conformance-dir> [NNNN ...]
Without numbers: every NNNN.nt in 00xx-02xx and 05xx-06xx.
The output is committed; the Go test needs no converter.
"""
import pathlib
import sys

from rdflib import Graph

d = pathlib.Path(sys.argv[1])
wanted = set(sys.argv[2:])
out = d / "formats"
out.mkdir(exist_ok=True)
for nt in sorted(d.glob("[0-9][0-9][0-9][0-9].nt")):
    n = nt.stem
    if wanted and n not in wanted:
        continue
    if not wanted and not (n < "0300" or "0500" <= n < "0700"):
        continue
    g = Graph()
    g.parse(nt, format="nt")
    for fmt, ext in (("turtle", "ttl"), ("xml", "rdf"), ("json-ld", "jsonld")):
        g.serialize(out / f"{n}.{ext}", format=fmt, encoding="utf-8")
    print(n)
