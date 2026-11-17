"""Write the rich conformance corpus: 05xx in schema.org, 06xx the same facts in CIDOC-CRM.

usage: python3 -I corpus.py   (writes next to itself)
Every record is self-contained: the people, places, materials and types it
reaches are copied into it, root first, then in the order they are reached.
"""
import pathlib

HERE = pathlib.Path(__file__).parent
EX = "http://example.org/c/"
S = "https://schema.org/"
RDF = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
RDFS = "http://www.w3.org/2000/01/rdf-schema#"
XSD = "http://www.w3.org/2001/XMLSchema#"
CRM = "http://www.cidoc-crm.org/cidoc-crm/"
DEF = "http://example.org/def/"


def iri(x):
    return f"<{x}>"


def lit(v, lang=None, dt=None):
    s = '"' + v.replace("\\", "\\\\").replace('"', '\\"') + '"'
    if lang:
        return f"{s}@{lang}"
    if dt:
        return f"{s}^^<{XSD}{dt}>"
    return s


A = iri(RDF + "type")


def s_(name):
    return iri(S + name)


def c(name):
    return iri(CRM + name)


def e(kind, key):
    return iri(f"{EX}{kind}/{key}")


# key: (parent, [(name, lang), ...]); the first name is the preferred one
PLACES = {
    "republiek": (None, [("Republiek der Zeven Verenigde Nederlanden", "nl"), ("Dutch Republic", "en"), ("Provinces-Unies", "fr")]),
    "holland": ("republiek", [("Holland", "nl"), ("Holland", "en"), ("Hollande", "fr")]),
    "leiden": ("holland", [("Leiden", "nl"), ("Leyden", "en"), ("Leyde", "fr"), ("Lugdunum Batavorum", "la")]),
    "amsterdam": ("holland", [("Amsterdam", "nl"), ("Amsterdam", "en"), ("Amstelodamum", "la")]),
    "haarlem": ("holland", [("Haarlem", "nl"), ("Harlem", "en"), ("Harlemum", "la")]),
}
# key: (name, (birth place, year) | None, (death place, year) | None, teacher | None)
PEOPLE = {
    "pieter": ("Pieter de Schilder", ("leiden", "1620"), ("amsterdam", "1675"), "jacob"),
    "paulus": ("Paulus de Schilder", ("leiden", "1620"), ("haarlem", "1675"), "jacob"),
    "adriaen": ("Adriaen de Schilder", ("leiden", "1622"), ("amsterdam", "1680"), "claes"),
    "jacob": ("Jacob de Meester", ("haarlem", "1590"), ("leiden", "1655"), None),
    "claes": ("Claes de Leermeester", ("amsterdam", "1595"), ("amsterdam", "1660"), None),
    "maria": ("Maria van Dorp", None, None, None),
    "willem": ("Willem van Dorp", None, None, None),
}
MATERIALS = {
    "eikenpaneel": [("eikenhouten paneel", "nl"), ("oak panel", "en")],
    "doek": [("doek", "nl"), ("canvas", "en")],
}
TECHNIQUE = {"eikenpaneel": ("paneelschildering", "panel painting"), "doek": ("doekschildering", "canvas painting")}

# n, title, painter, date, material, about, shape variant
RECORDS = [
    (1, "Portret van Maria", "pieter", "1660", "eikenpaneel", "maria", None),
    (2, "Portret van Maria, later", "pieter", "1685", "eikenpaneel", "maria", None),
    (3, "Portret van Maria, Haarlem", "paulus", "1660", "eikenpaneel", "maria", None),
    (4, "Portret van Maria op doek", "pieter", "1660", "doek", "maria", None),
    (5, "Portret van Maria, leerling van Claes", "adriaen", "1660", "eikenpaneel", "maria", None),
    (6, "Portret van Willem", "pieter", "1660", "eikenpaneel", "willem", None),
    (7, "Portret door de meester", "jacob", "1650", "eikenpaneel", "maria", None),
    (8, "Portret van Maria, gedateerd", "pieter", "1660-03-12", "eikenpaneel", "maria", "full-date"),
    (9, "Portret van Maria, plaats als tekst", "pieter", "1660", "eikenpaneel", "maria", "place-literal"),
    (10, "Portret van Maria, externe maker", "pieter", "1660", "eikenpaneel", "maria", "external-creator"),
    (11, "Portret van Maria, circa", "pieter", "ca. 1660", "eikenpaneel", "maria", "date-text"),
    (12, "Portret van Maria, materiaal als tekst", "pieter", "1660", "eikenpaneel", "maria", "material-literal"),
]


class Graph:
    def __init__(self):
        self.lines, self.seen = [], set()

    def add(self, s, p, o):
        line = f"{s} {p} {o} ."
        if line not in self.seen:
            self.seen.add(line)
            self.lines.append(line)

    def text(self):
        return "\n".join(self.lines) + "\n"


def date_lit(v):
    if len(v) == 4 and v.isdigit():
        return lit(v, dt="gYear")
    if len(v) == 10:
        return lit(v, dt="date")
    return lit(v)


def alt_name(place):
    return PLACES[place][1][1][0]  # e.g. "Leyden"


def external(person):
    return iri(f"http://example.org/ext/person/{person}")


# ---------- schema.org ----------

def schema_place(g, key):
    parent, ((n0, l0), *alts) = PLACES[key]
    p = e("place", key)
    g.add(p, A, s_("Place"))
    g.add(p, s_("name"), lit(n0, l0))
    for n, l in alts:
        g.add(p, s_("alternateName"), lit(n, l))
    if parent:
        g.add(p, s_("containedInPlace"), e("place", parent))
        schema_place(g, parent)


def schema_person(g, key, place_literal=False):
    name, born, died, teacher = PEOPLE[key]
    p = e("person", key)
    g.add(p, A, s_("Person"))
    g.add(p, s_("name"), lit(name))
    if born:
        g.add(p, s_("birthPlace"), lit(alt_name(born[0])) if place_literal else e("place", born[0]))
        g.add(p, s_("birthDate"), date_lit(born[1]))
    if died:
        g.add(p, s_("deathPlace"), e("place", died[0]))
        g.add(p, s_("deathDate"), date_lit(died[1]))
    if teacher:
        g.add(p, iri(DEF + "pupilOf"), e("person", teacher))
    if born and not place_literal:
        schema_place(g, born[0])
    if died:
        schema_place(g, died[0])
    if teacher:
        schema_person(g, teacher)


def schema_material(g, key):
    m = e("material", key)
    g.add(m, A, s_("DefinedTerm"))
    for n, l in MATERIALS[key]:
        g.add(m, s_("name"), lit(n, l))


def schema_record(n, title, painter, date, material, about, shape):
    g = Graph()
    w = iri(f"{EX}{500 + n:04d}")
    g.add(w, A, s_("CreativeWork"))
    g.add(w, A, s_("Painting"))
    g.add(w, s_("name"), lit(title, "nl"))
    g.add(w, s_("creator"), external(painter) if shape == "external-creator" else e("person", painter))
    g.add(w, s_("dateCreated"), date_lit(date))
    if shape == "material-literal":
        g.add(w, s_("material"), lit(MATERIALS[material][0][0], "nl"))
    else:
        g.add(w, s_("material"), e("material", material))
    g.add(w, s_("about"), e("person", about))
    if shape != "external-creator":
        schema_person(g, painter, place_literal=shape == "place-literal")
    if shape != "material-literal":
        schema_material(g, material)
    schema_person(g, about)
    return g


# ---------- CIDOC-CRM ----------

def crm_type(g, key, label, meta=None):
    t = e("type", key)
    g.add(t, A, c("E55_Type"))
    g.add(t, iri(RDFS + "label"), lit(label, "nl"))
    if meta:
        g.add(t, c("P2_has_type"), e("type", meta))
        crm_type(g, meta, meta)


def crm_appellations(g, owner, key, names):
    ids = [e("appellation", f"{key}-{i}") for i in range(len(names))]
    for a in ids:
        g.add(owner, c("P1_is_identified_by"), a)
    for i, (a, (n, l)) in enumerate(zip(ids, names)):
        g.add(a, A, c("E41_Appellation"))
        g.add(a, c("P190_has_symbolic_content"), lit(n, l))
        g.add(a, c("P2_has_type"), e("type", "voorkeursnaam" if i == 0 else "alternatieve-naam"))
    crm_type(g, "voorkeursnaam", "voorkeursnaam")
    if len(names) > 1:
        crm_type(g, "alternatieve-naam", "alternatieve naam")


def crm_timespan(g, key, v):
    ts = e("timespan", key)
    g.add(ts, A, c("E52_Time-Span"))
    if len(v) == 4 and v.isdigit():
        begin, end = f"{v}-01-01T00:00:00", f"{v}-12-31T23:59:59"
    elif len(v) == 10:
        begin, end = f"{v}T00:00:00", f"{v}T23:59:59"
    else:
        g.add(ts, c("P3_has_note"), lit(v))
        return
    g.add(ts, c("P82a_begin_of_the_begin"), lit(begin, dt="dateTime"))
    g.add(ts, c("P82b_end_of_the_end"), lit(end, dt="dateTime"))


def crm_place(g, key):
    parent, names = PLACES[key]
    p = e("place", key)
    g.add(p, A, c("E53_Place"))
    if parent:
        g.add(p, c("P89_falls_within"), e("place", parent))
    crm_appellations(g, p, f"place-{key}", names)
    if parent:
        crm_place(g, parent)


def crm_person(g, key, place_literal=False):
    name, born, died, teacher = PEOPLE[key]
    p = e("person", key)
    g.add(p, A, c("E21_Person"))
    if born:
        g.add(p, c("P98i_was_born"), e("birth", key))
    if died:
        g.add(p, c("P100i_died_in"), e("death", key))
    if teacher:
        g.add(p, c("P14i_performed"), e("training", key))
    crm_appellations(g, p, f"person-{key}", [(name, None)])
    for ev, cls, fact in (("birth", "E67_Birth", born), ("death", "E69_Death", died)):
        if not fact:
            continue
        place, year = fact
        x = e(ev, key)
        g.add(x, A, c(cls))
        if place_literal and ev == "birth":
            g.add(x, c("P3_has_note"), lit(alt_name(place)))
        else:
            g.add(x, c("P7_took_place_at"), e("place", place))
        g.add(x, c("P4_has_time-span"), e("timespan", f"{ev}-{key}"))
        crm_timespan(g, f"{ev}-{key}", year)
        if not (place_literal and ev == "birth"):
            crm_place(g, place)
    if teacher:
        t = e("training", key)
        g.add(t, A, c("E7_Activity"))
        g.add(t, c("P2_has_type"), e("type", "opleiding"))
        for role, who in (("leerling", key), ("leermeester", teacher)):
            g.add(t, c("P01i_is_domain_of"), e("pc14", f"{key}-{role}"))
        for role, who in (("leerling", key), ("leermeester", teacher)):
            pc = e("pc14", f"{key}-{role}")
            g.add(pc, A, c("PC14_carried_out_by"))
            g.add(pc, c("P02_has_range"), e("person", who))
            g.add(pc, c("P14.1_in_the_role_of"), e("type", role))
            crm_type(g, role, role)
        crm_type(g, "opleiding", "opleiding")
        crm_person(g, teacher)


def crm_record(n, title, painter, date, material, about, shape):
    g = Graph()
    k = f"{600 + n:04d}"
    w = iri(EX + k)
    g.add(w, A, c("E22_Human-Made_Object"))
    g.add(w, c("P102_has_title"), e("title", k))
    g.add(w, c("P108i_was_produced_by"), e("production", k))
    if shape == "material-literal":
        g.add(w, c("P3_has_note"), lit(MATERIALS[material][0][0], "nl"))
    else:
        g.add(w, c("P45_consists_of"), e("material", material))
    g.add(w, c("P62_depicts"), e("person", about))
    t = e("title", k)
    g.add(t, A, c("E35_Title"))
    g.add(t, c("P190_has_symbolic_content"), lit(title, "nl"))
    pr = e("production", k)
    g.add(pr, A, c("E12_Production"))
    g.add(pr, c("P14_carried_out_by"), external(painter) if shape == "external-creator" else e("person", painter))
    g.add(pr, c("P4_has_time-span"), e("timespan", k))
    tech_nl, tech_en = TECHNIQUE[material]
    g.add(pr, c("P32_used_general_technique"), e("type", tech_nl))
    crm_timespan(g, k, date)
    crm_type(g, tech_nl, tech_nl, meta="techniek")
    if shape != "external-creator":
        crm_person(g, painter, place_literal=shape == "place-literal")
    if shape != "material-literal":
        m = e("material", material)
        g.add(m, A, c("E57_Material"))
        for name, lang in MATERIALS[material]:
            g.add(m, iri(RDFS + "label"), lit(name, lang))
        g.add(m, c("P2_has_type"), e("type", "materiaal"))
        crm_type(g, "materiaal", "materiaal")
    crm_person(g, about)
    return g


for rec in RECORDS:
    n = rec[0]
    (HERE / f"{500 + n:04d}.nt").write_text(schema_record(*rec).text(), encoding="utf-8")
    (HERE / f"{600 + n:04d}.nt").write_text(crm_record(*rec).text(), encoding="utf-8")
    (HERE / f"{600 + n:04d}.bulk.json").write_text('{"recDefId": "crm"}\n', encoding="utf-8")
print(f"{len(RECORDS)} records x 2")
