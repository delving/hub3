# v2 conformance cases

One bulk request in, the v2 document hub3 sends to Elasticsearch out.

## Run

    go test ./ikuzo/service/x/bulk/ -run TestConformance           # check
    go test ./ikuzo/service/x/bulk/ -run TestConformance -update   # rewrite goldens

`-update` rewrites `NNNN.json` / `NNNN.error.json` only. Review the diff:
every golden change is a behaviour change.

## A case

- `NNNN.<nt|ttl|rdf|jsonld>`: the record graph; the extension is its mime type.
- `NNNN.json`: the expected v2 `_source`, `meta.modified` set to 0, 2-space indent.
- `NNNN.error.json`: `{"kind":"<stage>"}` instead, when hub3 refuses the record.
- `NNNN.bulk.json` (optional): envelope fields over `envelope.json`.
- `formats/NNNN.<fmt>`: the same graph in another format; must give the same
  document apart from `order`, blank-node labels and array order. No goldens.

## Ranges

| Range | Contents |
|---|---|
| 00xx | literals |
| 01xx | resource structure |
| 02xx | types and the root |
| 03xx | envelope |
| 05xx | rich corpus, schema.org |
| 06xx | rich corpus, CIDOC-CRM (mirrors 05xx) |
| 08xx | errors and limits |
| 09xx | real records |

## Cases

| # | Title | What it tests |
|---|---|---|
| 0001 | plain literal | `@value`, `entrytype`, `fields` projection |
| 0002 | language-tagged literals | `@language`, two languages on one predicate |
| 0003 | xsd:dateTime / date / gYear | `@type`, `isoDate`, `dateRange` |
| 0004 | xsd:integer / decimal | `integer`, `float` (invalid lexical form: see FINDINGS F7) |
| 0005 | escapes and unicode | `\"`, `\n`, `é`, emoji, combining character |
| 0006 | duplicate values | repeated triple; dedup in `fields` across resources |
| 0007 | long literal | over 256 characters; whitespace kept |
| 0101 | link to an IRI resource in the graph | `@id`, `level` 2, `context` |
| 0102 | link to an external IRI | entry without a resource; label resolution |
| 0103 | blank node, one level | bnode id, `entrytype` |
| 0104 | blank nodes, three levels | `level`, `context` path |
| 0105 | resource reached by two paths | `containsContext` dedup |
| 0106 | `rdf:List` | `rdf:first`/`rdf:rest` kept or flattened (no `formats/0106.rdf`: FINDINGS F8) |
| 0107 | cycle A → B → A | terminates; `level` |
| 0108 | interleaved predicates | `order` follows the source |
| 0201 | root typed `schema:CreativeWork` only | `entryURI`, `aboutTypeURI` |
| 0202 | root with several types | `types[]` order |
| 0203 | two CreativeWork subjects | which becomes `entryURI` |
| 0204 | `http://schema.org/` | rewritten, or not matched |
| 0205 | `edm:isShownBy` / image | `hasDigitalObject`; IRI absent from `fields` |
| 0301 | tags and type | `meta.tags`: `narthex`, `mdr` plus extras |
| 0302 | revision 7 | `meta.revision` |
| 0303 | recDefId `edm` | `aboutTypeURI`, root ore:Aggregation |

## Other formats

    PY="nix shell --impure --expr 'with import (builtins.getFlake \"nixpkgs\") {}; python3.withPackages (p: [p.rdflib])' -c python3 -I"
    eval $PY ikuzo/service/x/bulk/testdata/conformance/formats.py ikuzo/service/x/bulk/testdata/conformance [NNNN ...]

Regenerate after changing an `.nt` in 00xx–02xx or 05xx–06xx and commit the result.
