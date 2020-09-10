// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ead

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var defList = `
	<list type="deflist">
		<listhead>
		<head01>Trefwoord</head01>
		<head02>Verklaring</head02>
		</listhead>
		<defitem>
		<label>Ministerie, personeel:</label>
		<item>zie ook archief van het Kabinet en geheim archief.</item>
		</defitem>
		<defitem>
		<label>Personeel rechterlijke macht enz.:</label>
		<item>zie ook archief van het Kabinet en geheim archief.</item>
		</defitem>
		<defitem>
		<label>Ridderorden, Nederlandse:</label>
		<item>vnl. Kabinetsarchief; toestemming dragen buitenlandse: AS.</item>
		</defitem>
		<defitem>
		<label>Uitgifte Staatsblad:</label>
		<item>tot 1863 verzorgd door het Kabinet des Konings.</item>
		</defitem>
		<defitem>
		<label>Adelszaken:</label>
		<item>tot 1861 bij Binnenl. Zaken 10 e afd.; zie ook geheim archief. in 1937 naar Min. van Algemene Zaken.</item>
		</defitem>
		<defitem>
		<label>Auteursrecht:</label>
		<item>tot 1881 bij Binnenl. Zaken afd. KW (1876-1881) en 1e afd. ASC (voor 1976); vanaf 1913 uitvoering Auteurswet gesplitst: beleidszaken bij IC later 1A, beheer boeken bij AS; in 1941 naar departement van Volksvoorlichting en Kunsten.</item>
		</defitem>
		<defitem>
		<label>Erediensten:</label>
		<item>tot 1871 afzonderlijke administraties; vanaf 1871 ook Miniterie van Financiën; hiertoe hoort ook: toestemming tot aanvaarden van legaten, erfstellingen en makingen door kerkgenootschappen.</item>
		</defitem>
		<defitem>
		<label>Gerechtelijke stukken:</label>
		<item>rogatoire commissiën en buitenlandse dagvaardingen; gerechtelijke stukken in civiele zaken blijven vanaf 1911 bij IC.</item>
		</defitem>
		<defitem>
		<label>Naturalisaties:</label>
		<item>hiertoe ook: Nederlanderschap en burgerschapsrechten.</item>
		</defitem>
		<defitem>
		<label>Privaatrecht:</label>
		<item>verzamelterm voor:<lb></lb>burgerlijke en handelsrecht<lb></lb>burgerlijke stand<lb></lb>consulaire wetgeving (burgerlijke stand en notariaat)<lb></lb>Huwelijk en echtscheiding<lb></lb>krankzinnigen (opneming van Nederlanders in buitenlandse gestichten)<lb></lb>marine en scheepvaart (koopvaardij)<lb></lb>meerderjarig verklaring<lb></lb>minderjarigen<lb></lb>naamsverandering en -aanneming<lb></lb>nalatenschappen<lb></lb>voogdijvoorzieningen<lb></lb>wettiging.</item>
		</defitem>
		<defitem>
		<label>Rechterlijke organisatie:</label>
		<item>hiertoe ook notariaat, rechtswezen en rechtsbijstand onvermogenden.</item>
		</defitem>
		<defitem>
		<label>Rechtspersonen: hiertoe:</label>
		<item>verenigingen en vennootschappen.</item>
		</defitem>
		<defitem>
		<label>Sociale wetgeving: hiertoe:</label>
		<item>Arbeidswet, 1889-1893, vervolgens naar Min. van Waterstaat, Handel en Nijverheid, Afd. Arbeid en Fabriekswezen; Beroepswet 1901.</item>
		</defitem>
		<defitem>
		<label>Staats- en volkenrecht, administratierecht: verzamelterm voor:</label>
		<item>Buitenlandse betrekkingen: vnl. tractaten en buitenlandse wetgeving handels- en fabrieksmerken 1881-1893, vervolgens ook naar Min. van Waterstaat, Handel en Nijverheid, Afd. Handel &amp; Nijverheid I. industieel eigendom 1903-1914, vervolgens naar Min. van Landbuw, Nijverheid en Handel, afd. Handel<lb></lb>loterijen, vanaf 1905.<lb></lb>rechterlijke organisatie in Egypte<lb></lb>uitgeslotenen kiesrecht, vanaf 1923<lb></lb>georganiseerd overleg, vanaf 1919</item>
		</defitem>
		<defitem>
		<label>Strafrecht en strafvordering: verzamelterm voor:</label>
		<item>bertillonnage, 1896-1922<lb></lb>drankwet, 1883-1905, vervolgens naar Min. van Binnenlandse Zaken.<lb></lb>gerechtelijke stukken en buitenlandse dagvaardingen in strafzaken, vanaf 1911.<lb></lb>marine en scheepvaart (desertie en overtredingen zeebrievenwet)<lb></lb>rogatoire commissiëen<lb></lb>statistiek (gerechtelijke)<lb></lb>strafregister, vanaf 1896<lb></lb>alsmede de in 1934 naar de 5e afdeling overgebracht onderwerpen:<lb></lb>politie op de Noordzee, vanaf 1884<lb></lb>vuurwapenwet, vanaf 1890<lb></lb>vrouwenhandel<lb></lb>bestrijding ontuchtige uitgaven<lb></lb>jachtwet, vanaf 1923<lb></lb>motor- en rijwielwet, vanaf 1927<lb></lb>opiumwet, vanaf 1928</item>
		</defitem>
		<defitem>
		<label>Vreemdelingenzaken: verzamelterm voor:</label>
		<item>vreemdelingen (zie ook bij politie)<lb></lb>tractaten van uitlevering<lb></lb>uitleveringen</item>
		</defitem>
		<defitem>
		<label>Gratie</label>
		<item>burgerlijke veroordeelden<lb></lb>jeugdige veroordeelden<lb></lb>militaire veroordeelden<lb></lb>verloven aan gevangenen<lb></lb>af- en ontslag van gevangenen<lb></lb>rehabilitatie, tot 1900<lb></lb>voorwaardelijke invrijheidstelling, vanaf 1915 naar Reclassering</item>
		</defitem>
		<defitem>
		<label>Politie</label>
		<item>algemeen politieblad<lb></lb>marechaussee<lb></lb>personeel: reglementen en verordeningen; toelating en uitzetting van vreemdelingen<lb></lb>rijksveldwacht: detacheringen; gratificaties; inspecties; kleding en wapening; personeel (bezoldigd en onbezoldigd); rijkswoningen; rijwielen; sollicitanten; vergoedingen voor woninghuur; watersurveillance; jacht en visserij, vanaf 1905 naar Min. van Landbouw, Nijverheid en Handel.</item>
		</defitem>
		<defitem>
		<label>Gevangeniswezen</label>
		<item>bevolking<lb></lb>cellulaire rijtuigen<lb></lb>comptabiliteit<lb></lb>gebouwen en meubilair<lb></lb>imprimés<lb></lb>inspectiën en jaarverslagen<lb></lb>landbouw, tuinbouw, veeteelt<lb></lb>maandstaten en andere verantwoordingsstukken<lb></lb>personeel: colleges van regenten; gevangenissen en rijkswerkinrichtingen; verloven<lb></lb>vakonderricht<lb></lb>verkopingen, verhuringen, verpachtingen<lb></lb>voeding, kleding, verpleging</item>
		</defitem>
		<defitem>
		<label>Rijkstucht- en opvoedingswezen</label>
		<item>bevolking<lb></lb>gebouwen en meubilair<lb></lb>imprimés<lb></lb>inspectiën en jaarverslagen<lb></lb>landbouw, tuinbouw, veeteelt<lb></lb>maandstaten<lb></lb>particuliere gestichten<lb></lb>personeel: algemeen college, commissies van toezicht, voogdijraden; rijksopvoedingsgestichten en tuchtscholen; verloven<lb></lb>voeding, kleding, verpleging<lb></lb>voogdijraden</item>
		</defitem>
		<defitem>
		<label>gevangenisarbeid</label>
		<item>boekhouding en administratie arbeidslonen<lb></lb>declaratiën<lb></lb>grondstoffen<lb></lb>leveringen aan departementen<lb></lb>leveringen gevangenissen<lb></lb>plan van arbeid<lb></lb>prijsberekeningen<lb></lb>Rijksinkoopbureau, vanaf 1922</item>
		</defitem>
		<defitem>
		<label>Reclassering</label>
		<item>Centraal college<lb></lb>instellingen<lb></lb>reclassering<lb></lb>lijst van niet-reclassabele personen; in geheim archief</item>
		</defitem>
		<defitem>
		<label>Comptabiliteit</label>
		<item>pensioenen<lb></lb>tractementen<lb></lb>gerechtsgebouwen</item>
		</defitem>
		<defitem>
		<label>Wetgeving</label>
		<item>in 1939 gecentraliseerd bij de 6e afdeling</item>
		</defitem>
	</list>
`

func Test_itemBuilder_parse(t *testing.T) {
	type args struct {
		b []byte
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		length  int
		items   []*DataItem
	}{
		{
			"no lb",
			args{[]byte(`<item>bertillonnage, 1896-1922, drankwet, 1883-1905 opiumwet, vanaf 1928</item>`)},
			false,
			1,
			[]*DataItem{
				{Type: 5, Text: "bertillonnage, 1896-1922, drankwet, 1883-1905 opiumwet, vanaf 1928", Depth: 1, ParentIDS: "", Tag: "item", Order: 1},
			},
		},
		{
			"emph",
			args{[]byte(`<item>bertillonnage, <emph>1896-1922,</emph> drankwet</item>`)},
			false,
			1,
			[]*DataItem{
				{Type: 5, Text: "bertillonnage, <em>1896-1922,</em> drankwet", Depth: 1, ParentIDS: "", Tag: "item", Order: 1},
			},
		},
		{
			"double lb",
			args{[]byte(`<item>bertillonnage, 1896-1922,<lb></lb> drankwet</item>`)},
			false,
			1,
			[]*DataItem{
				{Type: 5, Text: "bertillonnage, 1896-1922,<lb/> drankwet", Depth: 1, ParentIDS: "", Tag: "item", Order: 1},
			},
		},
		{
			"single lb",
			args{[]byte(`<item>bertillonnage, 1896-1922,<lb/> drankwet</item>`)},
			false,
			1,
			[]*DataItem{
				{Type: 5, Text: "bertillonnage, 1896-1922,<lb/> drankwet", Depth: 1, ParentIDS: "", Tag: "item", Order: 1},
			},
		},
		{
			"<lb/>",
			args{[]byte(`<defitem>
		<label>Strafrecht en strafvordering: verzamelterm voor:</label>
		<item>bertillonnage, 1896-1922<lb></lb>drankwet, 1883-1905</item>
		</defitem>`)},
			false,
			3,
			[]*DataItem{
				{Type: 6, Text: "", Depth: 1, ParentIDS: "", Tag: "defitem", Order: 1},
				{Type: 7, Text: "Strafrecht en strafvordering: verzamelterm voor:", Depth: 2, ParentIDS: "1", Tag: "label", Order: 2},
				{Type: 5, Text: "bertillonnage, 1896-1922<lb/> drankwet, 1883-1905", Depth: 2, ParentIDS: "1", Tag: "item", Order: 3},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ib := newItemBuilder(context.TODO())
			if err := ib.parse(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("itemBuilder.parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.length != len(ib.items) {
				t.Errorf("itemBuilder.parse() got items = %d, wantErr %d", len(ib.items), tt.length)
			}

			if diff := cmp.Diff(tt.items, ib.items); diff != "" {
				t.Errorf("itemBuilder.parse() mismatch (-want +got):\n%s", diff)
			}

			t.Logf("item builder: %#v", ib.items[len(ib.items)-1])
			// t.FailNow()
		})
	}
}
