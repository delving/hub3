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
