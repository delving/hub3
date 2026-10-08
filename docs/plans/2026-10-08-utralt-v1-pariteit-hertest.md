# Utralt: v1-pariteit hertesten tegen de Go-instantie

**Datum:** 2026-10-08
**Status:** klaar om te activeren zodra de tweede instantie staat
**Oude stack:** `prod.utralt.hubs.delving.org` (207.154.244.3), nave (Django), `nave.service`
**Nieuwe stack:** _nog in te vullen — basis-URL van de Go-instantie_

## Waarom dit licht mag

`verhaalvanutrecht.nl` gaat binnenkort over op een v2-site. De v1-API hoeft
daarna niet meer te werken, dus de vraag is niet "is v1 in Go volledig" maar
"houdt v1 het uit tot de overstap". Dat verschil rechtvaardigt een
boodschappenlijstje in plaats van een pariteitsproject: de unieke queries
opnieuw sturen naar beide kanten en de uitkomsten vergelijken.

Dat de hertest zo klein kan, komt niet uit een aanname maar uit meting op de
access log (2026-10-08). Beide vensters, want ze zeggen iets anders:

| | utralt archief (3,7 jaar) | utralt nu (13 dagen) | BrabantCloud (5 jaar) |
|---|---|---|---|
| requests op `/api/search` | 131.434.453 | 1.151 (~88/dag) | 161.074.820 |
| unieke parametercombinaties | **199** | 10 | **3.474** |
| unieke parameternamen | **163** | 15 | **1.057** |
| combinaties voor 99% van het verkeer | **9** | — | 27 |
| consumenten (referrer) | — | 1 — `verhaalvanutrecht.nl` | meerdere |

Het **volume** is dus vergelijkbaar met BrabantCloud; de **diversiteit** is
zeventien keer kleiner. Negen combinaties dekken 99% van bijna vier jaar
verkeer. Dat is wat utralt een goede kandidaat maakt: niet dat er niemand komt,
maar dat het oppervlak smal is.

Alles in het huidige venster draagt `callback` en `diw-version`: de
DIW-widget, via JSONP.

### Het verkeer is voor vier vijfde more-like-this

82,4% van het archief (108,2 van 131,3 miljoen requests) draagt `mlt`. De twee
grootste combinaties samen zijn 106 miljoen aanbevelingsverzoeken. Zoeken is de
minderheid: 17,6%, in 88 combinaties.

In het huidige venster is dat gekanteld — daar is mlt 43 van 1.151 requests,
3,7%. Die 106 miljoen komen dus vrijwel zeker van een oudere versie van de site
die er niet meer is.

Veertien van de vijftien parameters komen letterlijk voor in de
Go-implementatie: `facet.field`, `qf[]`, `qf`, `lang`, `format`, `callback`,
`sortBy`, `rows`, `query`, `page`, `facetBoolType`, `mlt`, `mlt.count`, `id`.
De vijftiende is `diw-version`, die Go niet kent — ongevaarlijk, want de
parameterlus in `hub3/fragments/api.go:188` is een `switch p` zónder `default`,
dus een onbekende naam wordt stil genegeerd. Dat is nagekeken omdat het de enige
manier was waarop juist die parameter de hele API had kunnen breken.

Het archief was de moeite van het doorrekenen waard: het venster van 13 dagen
liet 10 combinaties zien, het archief 199. "Tien" was dus een momentopname. De
replay-set is op het archief gebaseerd en niet op die tien.

## Voorwaarde, en de valkuil die erin zit

De nieuwe stack is een eigen server met eigen Elasticsearch, los geïndexeerd,
die v1 en v2 bedient. Dat is de opzet, en die is prima — maar hij verandert
code en data tegelijk. Een verschil in `numFound` kan dan twee dingen betekenen:
de API antwoordt anders, of de index is anders gevuld. Zonder dat te scheiden
is het rapport niet te lezen.

Dat is niet weg te nemen door anders te meten; het moet gemeten worden. Vandaar
stap 0: eerst de twee indexen tegen elkaar, dán de queries. Geen zoekresultaat
betekent iets zolang de onderliggende verzamelingen niet gelijk zijn.

Twee dingen om bij het opzetten vast te leggen, want achteraf zijn ze niet meer
te reconstrueren:

- **Vul beide indexen uit dezelfde bron-momentopname.** Wordt de nieuwe index
  gevuld uit de huidige brondata terwijl de oude maanden oud is, dan is elk
  verschil ruis en meet je niets over de API. Lukt dat niet, noteer dan welke
  datasets sindsdien zijn gewijzigd en houd die buiten de vergelijking.
- **Noteer of de oude index tijdens de test nog gevuld wordt.** Draait narthex
  nog richting de oude stack, dan beweegt je referentie onder de meting door.

Eén voordeel van een eigen Elasticsearch: de nieuwe kant mag voluit bevraagd
worden. Alleen de oude kant heeft de rem nodig, want daar hangt de echte site
aan.

## Stap 0 — de indexen gelijkleggen

Doe dit vóór stap 3; de uitkomst bepaalt of een queryverschil überhaupt iets
zegt.

```bash
OUD=https://prod.utralt.hubs.delving.org
NIEUW=<BASIS-URL VAN DE NIEUWE STACK>

# aantal documenten per dataset, aan beide kanten
for B in "$OUD" "$NIEUW"; do
  echo "== $B"
  curl -s "$B/api/search/v1/?format=json&rows=0&facet.field=delving_spec&facetBoolType=and" \
    | python3 -c 'import json,sys
d=json.load(sys.stdin)["result"]
for f in d.get("facets",[]):
    for l in f.get("links",[]):
        print(l["value"], l["count"])' | sort
done
```

Zet die twee lijsten naast elkaar met `diff`. Wat je wil zien is een lege diff.
Wat je waarschijnlijk ziet is een handvol datasets die schelen, en dán is de
vraag per geval: is dat drift in de brondata of doet de indexer iets anders?
Alleen het tweede is een bevinding.

Scheelt het totaal meer dan een procent of twee, stop hier en zoek dat eerst
uit. Queries vergelijken op verzamelingen die niet gelijk zijn levert een
rapport vol verschillen die allemaal dezelfde oorzaak hebben, en dat kost meer
tijd om uit te pluizen dan het nu kost om het recht te zetten.

Neem daarnaast een steekproef op recordniveau: tien `doc_id`'s, aan beide kanten
opgehaald met `format=json&id=<doc_id>`, en de velden vergeleken. Dat vangt een
indexer die alles indexeert maar iets anders *in* de documenten zet — en dat
zou je in aggregatiecijfers nooit zien.

## Stap 1 — corpus verversen uit de logs

Het analysescript staat in `tools/parity/corpus/v1-analyse.awk` en is hetzelfde
dat de BrabantCloud-set maakte, zodat de cijfers vergelijkbaar blijven.

```bash
H=prod.utralt.hubs.delving.org
scp tools/parity/corpus/v1-analyse.awk root@$H:/tmp/

# recent venster: wat de site nu stuurt
ssh root@$H 'mkdir -p /tmp/utralt-recent &&
  nice -n 19 awk -v UIT=/tmp/utralt-recent -f /tmp/v1-analyse.awk \
    /var/log/hub3-utralt/nginx.log'

# vier-jaars archief: 1,8 GB gz, ~20 min op deze machine (2 cores, 3 GB).
# nice + ionice, want de productiedienst draait ernaast.
ssh root@$H 'mkdir -p /tmp/utralt-archief && nohup sh -c "
  nice -n 19 ionice -c3 zcat /var/log/hub3-utralt/nginx-access-20230109-20260925.log.gz |
  nice -n 19 awk -v UIT=/tmp/utralt-archief -f /tmp/v1-analyse.awk
  echo KLAAR > /tmp/utralt-archief/.klaar" > /tmp/utralt-archief/run.log 2>&1 &'

# wachten tot .klaar bestaat, dan:
ssh root@$H 'cat /tmp/utralt-archief/samenvatting.txt
  echo "signatures: $(wc -l < /tmp/utralt-archief/signatures.txt)"
  echo "parameters: $(wc -l < /tmp/utralt-archief/parameters.txt)"'
```

Let op bij het lezen van die map: `ls | head` kapte bij mij `nginx.log` eraf en
daardoor dacht ik even dat het huidige verkeer nergens gelogd werd. Lees de map
volledig.

## Stap 2 — replay-set

**De set staat er al:** `tools/parity/corpus/utralt-corpus.tsv`, 469 gevallen.

| | aantal | combinaties |
|---|---|---|
| zoeken en geavanceerd zoeken | **288** | 89 |
| more-like-this (rookproef) | 181 | 110 |

Zoeken is volledig gedekt: alle 89 niet-mlt-combinaties uit bijna vier jaar
verkeer zitten erin, met hun echte URL's. Dat kan omdat 89 klein is — de
BrabantCloud-set moest van 3.475 combinaties naar 3.197 gevallen worden gewogen
omdat één combinatie 61% van het verkeer was; hier is wegen niet nodig.

Opnieuw genereren hoeft alleen als er iets aan de site verandert. Dan:
`voorbeelden.tsv` uit beide analyses samenvoegen, dedupliceren op URL, en
mlt-combinaties tot twee voorbeelden elk beperken.

### Waarom mlt een rookproef is en geen vergelijking

Een mlt-request is in de log altijd een detailweergave met een `id`: hij moet
het gevraagde record opleveren plus een paar aanbevelingen. **Welke
aanbevelingen hoeft niet gelijk te zijn**, en dat is geen gemakzucht maar
natuurkunde. MLT rangschikt op termfrequenties over de hele index, dus welke
termen kenmerkend genoeg zijn hangt af van wat er nog meer in staat. Twee los
opgebouwde indexen geven daardoor legitiem andere aanbevelingen bij identieke
code en identieke documenten. Gelijke uitvoer eisen levert dan fouten die geen
defecten zijn, en de echte bevindingen verdrinken erin.

Dat onderscheid zit in `replay.py` zelf en niet in een afspraak die iemand
vergeet. `is_mlt()` kijkt of de query `mlt`, `mlt.count` of `mlt.filterkey` als
parameter**naam** draagt — niet als waarde, want `query=mltipliek` is geen
mlt-geval. Mlt-zaken gaan door `MLT_LAGEN` (vier lagen), de rest door `LAGEN`
(elf lagen). Het rapport krijgt een `soort`-kolom en de samenvatting is gesplitst
per soort, zodat een zoekregressie zich niet kan verstoppen tussen de verwachte
mlt-verschillen.

## Stap 3 — beide kanten bevragen

```bash
python3 tools/parity/replay.py \
  --corpus tools/parity/corpus/utralt-corpus.tsv \
  --a https://prod.utralt.hubs.delving.org \
  --b <BASIS-URL VAN DE GO-INSTANTIE> \
  --out tools/parity/utralt-rapport-$(date +%Y-%m-%d).tsv
```

`--a` is de referentie (nave), `--b` de kandidaat (Go). Beide kanten zijn
alleen-lezen, dus een run verandert niets. Reken op een kleine twee minuten
voor 469 gevallen.

De standaardwaarden `--workers 4 --rate 0.1` blijven staan, maar om één reden:
aan de oude kant hangt de echte site. De nieuwe stack heeft eigen
Elasticsearch en zou voluit kunnen, maar `replay.py` bevraagt beide kanten per
geval kort na elkaar — en dat is precies de bedoeling, want anders zie je een
indexering die ondertussen doorloopt als verschil terwijl het tijd is.

## Stap 4 — wat het rapport zegt

De samenvatting komt in twee tabellen, `zoek` en `mlt`, en het rapport heeft een
`soort`-kolom. Lees ze apart — ze hebben verschillende eisen.

**Zoeken (288 zaken), elf lagen, van grof naar fijn:**

```
status · content_type · omhulsel · numFound · volgorde · zelfde_set
facetnamen · facettellingen · veldset · veldwaarden · paginering
```

Die laagjes zijn geen opsmuk. "Zelfde resultaten, andere volgorde" en "andere
resultaten" zijn verschillende problemen met verschillende oplossingen, en een
enkel pass/fail verstopt dat. Een verschil in `numFound` verklaart bovendien
alles eronder en hoeft niet nog eens als "andere resultaten" te worden geteld.

**More-like-this (181 zaken), vier lagen:** `status`, `content_type`,
`omhulsel`, `mlt_gevuld`. Die laatste vraagt alleen of er aan beide kanten
aanbevelingen terugkomen, niet welke.

**Geslaagd, voor dit doel:**

- *zoeken*: geen afwijkingen op `status`, `omhulsel`, `numFound` en
  `zelfde_set`, en de facetnamen gelijk. Dat is wat de widget gebruikt, en hier
  wil je geen regressie missen — dit is waarom deze hertest bestaat.
- *mlt*: alle vier de lagen gelijk. Dus: hij werkt, hij geeft het record, en er
  komen aanbevelingen uit. Niet welke.

**Geen blokkade voor de overstap:** `volgorde` bij gelijke scores,
`facettellingen` die één of twee schelen, en afwijkingen in de staart van
kapotte query strings. Noteer ze, los ze niet op — v1 verdwijnt.

Reken bij twee los geïndexeerde clusters op meer `volgorde`-verschillen dan je
bij één gedeelde index zou zien. Scores hangen af van termfrequenties over de
hele index, dus een paar documenten verschil verschuift de rangschikking bij
bijna-gelijke relevantie. Dat is geen API-verschil. Verschilt de volgorde bij
een expliciete `sortBy`, dan wél — dat is code.

**Wel blokkerend:** een status die verschilt (de één weigert wat de ander
beantwoordt), een leeg omhulsel aan één kant, of `numFound` die afwijkt op een
dataset waarvan stap 0 zei dat de aantallen gelijk zijn. Die laatste conditie is
de hele reden dat stap 0 bestaat: zonder die vergelijking is "numFound wijkt af"
geen bevinding maar een vraag.

## Wat al eerder is betaald, dus niet opnieuw ontdekken hoeft

Deze drie kostten bij BrabantCloud tijd en zitten al in `replay.py`:

- **Geen vaste `Accept: application/json`.** Een groot deel van het echte
  verkeer vraagt JSONP, en die header ernaast laat Django met 406 antwoorden.
  De `format`-parameter beslist; laat hem.
- **JSONP uitpakken voor je vergelijkt.** Alles van utralt draagt `callback`,
  dus zonder uitpakken vergelijk je twee stukken omhullende JavaScript.
- **Velden per `doc_id` vergelijken, niet per positie.** Mijn eerste versie deed
  het op positie en rapporteerde tien ontbrekende afbeeldingsvelden die er wel
  waren.

En: byte-gelijkheid haal je nooit. Scroll-ID's, tijdstempels en de ordening
binnen gelijke scores verschillen altijd.

## De ene waarborg die ik ook bij dit lichte bereik zou houden

Draai de site zelf naast elkaar in twee browservensters voor je omzet: zoeken,
een facet aanklikken, een detail openen. Eén afnemer maakt dat goedkoop, en het
vangt wat een JSON-diff niet ziet — dat een veld klopt maar de widget er iets
anders mee doet.

Reden om dat niet weg te laten: met 88 requests per dag merkt niemand een
regressie aan het verkeer. Er komen geen klachten die je waarschuwen. De
vergelijking *is* de poort, en houd `nave.service` draaibaar tot de site een
paar dagen goed heeft gelopen.

## Niet in dit plan

**v2-pariteit.** De nieuwe stack bedient v1 en v2, maar `verhaalvanutrecht.nl`
gebruikt v1 — dat is wat de logs zeggen: 1.151 requests op `/api/search/v1`,
en geen enkele op v2. De v2-API is daarmee niet wat deze omzetting riskeert.
Zodra de site op de v2-variant gaat, is dát het moment voor een eigen
vergelijking, en die heeft een ander corpus nodig dan dit plan levert.

**De oude stack opruimen.** Houd `nave.service` draaibaar tot de site een paar
dagen goed heeft gelopen op de nieuwe. Met 88 requests per dag duurt het langer
dan je denkt voor een regressie zich meldt.
