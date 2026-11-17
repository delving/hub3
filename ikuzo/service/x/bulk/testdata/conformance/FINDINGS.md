# Conformance findings

From reviewing the baseline goldens. Status: open, fixed, deferred, wontfix.
A fix is one commit changing code and the affected goldens; it sets the row
to fixed and names the commit.

The harness runs on hub3's default configuration (no `hub3.toml`): `RDFTag.Label`
is `skos:prefLabel`, `dc:title`; `RDFTagMap` is empty. Findings that depend on
configuration say so.

| ID | Cases | Observed | Proposal | Status |
|---|---|---|---|---|
| F1 | — | `index.Graph` declares `json:"_ checksum,omitempty"` (space, `ikuzo/rdf/index/graph.go:26`); mapping declares `_checksum` | fix the tag | open |
| F2 | 0101, 05xx | `sortValue` and `inline` are entry fields (`hub3/fragments/resource.go:1351,1357`) not in the mapping (dynamic) | declare or drop | open |
| F3 | — | `resolvedFrom`, `resolvedLevel`, `mType`, `mRole`, `mFilterID` in mapping but not in `fragments.ResourceEntry` | decide which model writes them | open |
| F4 | 0205 | `tree.hasDigitalObject` vs `meta.hasDigitalObject` share a name, differ in subject | document; no code change | open |
| F5 | 0301 | Confirmed by reading: `Parser.Parse` makes a fresh `Request` per NDJSON line (`parser.go:92`); `req.setTags()` is only called from `setDataSet`, which runs under `p.once` (`parser.go:324`, `:189`). Only the first record processed gets `resolvedTags`; every later record in the upload has `meta.tags: null`, and `fragmentsOnly`/`rdfOnly` on a later line are ignored. With several workers, "first" is whichever worker wins. | call `req.setTags()` per request in `process`; keep only the dataset lookup in `once` | open |
| F6 | 0805 | no limit on nested documents; 0805 (1 resource + 1,001 entries) is indexed, 261 KB; OpenSearch refuses > 50,000 after the round trip | `record_too_large` above 1,000 | open |
| F7 | 0004 (input trimmed) | One invalid typed literal (`"abc"^^xsd:integer`) makes `NewResourceEntry` return an error (`resource.go:1045`); `SetEntries` (`resource.go:983`) iterates `fr.predicates`, a map, and stops at that error; `ResourcesList` logs and swallows it (`builder.go:72`). The resource keeps a **random subset** of its entries: the same `.nt` gave `"entries": []` on one run and width + height on another. | keep the entry without `integer`/`float` (log), never abort the resource; add case 0806 with the invalid literal once fixed | open |
| F8 | 0106 | `rdf:List` is flattened to repeated values (`schema_keywords: ["molen","rivier"]`) from N-Triples, Turtle and JSON-LD, but not from RDF/XML: `formats/0106.rdf` gives `"@id":"_:b","entrytype":"Bnode"` plus `rdf:first`/`rdf:rest` entries. The ntriples and jsonld parsers call `g.Inline()` (`ntriples/parser.go:36`, `jsonld/parser.go:57`); the rdfxml parser does not. `formats/0106.rdf` is not committed. | call `g.Inline()` in the rdfxml parser; regenerate `formats/0106.rdf` | open |
| F9 | 0107 | A cycle A → B → A terminates, but both get `"level": 7` and contexts `[root L1, a L6, b L5]`: the walk runs every path to the depth cap. | follows from F11/F14: a breadth-first walk with a visited set gives A level 2, B level 3 | open |
| F10 | 0204, 0803 | When no subject has the recDef's type, `GetAboutURI` falls back to a structural root (`ikuzo/rdf/graph.go:466`, warns). 0803 (only a `schema:Person`) and 0204 (`http://schema.org/CreativeWork`) are indexed with `meta.aboutTypeURI: ["https://schema.org/CreativeWork"]`, which the root does not have. `http://schema.org/` is not normalised: 0204 gets search label `sorg_name`. | (a) set `aboutTypeURI` to what matched, or refuse; (b) rewrite `http://schema.org/` to `https://` at parse | open |
| F11 | 05xx/06xx `formats/` | `level` and `context` depend on triple order, not on the graph. `SetContextLevels` (`resource.go:1206`) sets `levelN.Level = N` on context objects shared between resources, so the path walked first wins: in 0501 Turtle `place/republiek` is level 5, in N-Triples level 6. All 42 differing `formats/` subtests of 05xx/06xx differ only in `level`/`context` (checked by comparing everything else); they are skipped by name (`knownFormatDiffs`). | rewrite `SetContextLevels` as a breadth-first walk with a visited set: level = shortest distance from the root; contexts per path built fresh | open |
| F12 | 0601–0612, 0910 | `SetContextLevels` is six hand-unrolled loops: a resource deeper than level 7 gets **no context and level 1**, as if it were a property of the root. In 0601 `place/haarlem`, its three appellations and the time-spans of Jacob's birth and death are all `"level": 1`. 0910 (Gouda) already reaches level 7. | fixed by the F11 rewrite (no depth cap, or a configurable one that drops instead of promoting) | open |
| F13 | — | In the same function, from level 3 on `linkedObjects[levelN.ObjectID]` is set to the *parent* resource (`level3Resource` for `level3.ObjectID`, etc.), so the returned map holds wrong values. | fixed by the F11 rewrite | open |
| F14 | 0601, 0910 | No visited set: every path is walked to depth 6, so cost grows with the number of paths, not resources (shared types such as `voorkeursnaam` get 12 contexts in 0601). Likely part of the indexing cost on dense records. | fixed by the F11 rewrite; measure 0601/0910 before and after | open |
| F15 | 0902, 0907, 0908 (and 0906 vs 0903) | `BySortOrder.Less` (`resource.go:894`) returns `true` whenever the first resource has no entries, whatever the second: an inconsistent comparator, so entry-less resources (identifiers, types) come out in map order. 0902/0907/0908 differ between runs; they are compared order-insensitively by name (`knownUnstable`). | order entry-less resources by first appearance in the source (or by id) | open |
| F16 | 0302 | `setDataSet` also sets `req.Revision = ds.Revision` (`parser.go:186`) for the first request only; later records keep the revision from their own envelope. | with F5: resolve per request | open |
| F17 | 0002 | `@language` is kept per entry, but `fields` flattens languages: `schema_name: ["Gezicht op Delft","View of Delft"]`. | decide whether `fields` needs language (e.g. `schema_name_nl`) or stays a plain projection; document | open |
| F18 | 0101, 0102, 05xx | Link entries carry the target's label as `@value` only when the target has a `RDFTag.Label` predicate: 0303 (`dc:title`) has `"@value": "Aggregatie"` on `edm_aggregatedCHO`; 0101 (`schema:name`) has no label on `schema_creator`. With the default config schema.org records get no link labels at all. External IRIs (0102) get none either way. | run the harness with the NDE `hub3.toml` (label = `schema:name`, `skos:prefLabel`, `rdfs:label`, …) or document that goldens are default-config | open |
| F19 | 0003, 0104 | With the default config no entry gets `isoDate`, `dateRange` or `latLong`: these come only from `RDFTagMap`, not from `xsd:date`/`gYear`/`dateTime`. 0003 has `@type` only. | either derive `isoDate`/`dateRange` from the datatype, or load an rdftag config in the harness (with F18) | open |
| F20 | 0004 | `xsd:integer` gives `"integer": 45`; `xsd:decimal` (`38.5`) gives nothing — only `xsd:float` fills `float` (`resource.go:1051`). | also map `xsd:decimal`/`xsd:double` to `float` | open |
| F21 | 0102 | `fields` is `null`, not `{}`, when a record has no literal values. | harmless; document or emit `{}` | open |
| F22 | 01xx, 05xx | `context` objects serialise with Go field names (`Subject`, `SubjectClass`, `Predicate`, `SearchLabel`, `Level`, `ObjectID`): no json tags, unlike every other field. | add camelCase json tags (mapping change) or document | open |
| F23 | 0906 vs 0903 | Same museum-sloten record: JSON-LD relative IRIs (`"isShownBy": "toBeRemoved"`) are dropped silently, so 0906 loses 6 aggregation entries and `hasDigitalObject` is false; 0903 has them (true). Fixture artefact, but nothing warns. | warn on dropped relative IRIs; no golden change | open |

Not findings (behaviour checked and fine):
0005 escapes, emoji and combining characters intact; 0006 a repeated triple gives
one entry, `fields` dedups across resources; 0007 a 300-character value and
double spaces kept; 0108 `order` follows the source; 0105 a resource reached by
two paths gets two contexts; 0203 the first CreativeWork is the root, the second
is nested at level 2; 0202 `types` in source order; 0301/0302 tags and revision
from the envelope (for the first request, see F5/F16); 0801/0802 `rdf_parse`,
0804 `validation`; no `untyped` refusals.

## Query assumptions, as far as the goldens answer them

| A | Answer from the goldens |
|---|---|
| A1 | Broken for CRM: in 0601 the `crm_P32_used_general_technique` entry has only `@id` (no label, F18), and "techniek" sits two hops further (`type/paneelschildering` → `P2_has_type` → `type/techniek` `rdfs_label`, level 4). |
| A2 | Broken for CRM: `crm_P82a_begin_of_the_begin` "1660-01-01T00:00:00" is at level 3 in 0601; birth/death dates of the painter at level 7, Jacob's at level 1 (F12). |
| A3 | Broken for both models with the default config (F18); for CRM the name is on `E41` appellations one hop further. |
| A4 | Not testable from documents alone; 0601 shows `PC14` nodes at level 7 with `P14.1_in_the_role_of` as an `@id` entry. |
| A5 | Holds for schema.org: 0501 `fields.schema_alternateName` contains "Leyden", "Leyde", "Lugdunum Batavorum". CRM: only via `crm_P190_has_symbolic_content`, not tied to the place. |
| A6 | Not answerable from the goldens; needs the query follow-up. |
