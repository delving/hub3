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

// nolint:gocritic
package ead_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/google/go-cmp/cmp"
	r "github.com/kiivihal/rdf2go"
	"github.com/matryer/is"
)

func NewSubject(spec, eadType, id string) string {
	identifier := strings.Join([]string{eadType, id}, "/")

	return fmt.Sprintf(
		"%s/%s/archive/%s/%s",
		config.Config.RDF.BaseURL,
		"test",
		spec,
		identifier,
	)
}

func TestDidTriples(t *testing.T) {
	is := is.New(t)

	cc := new(ead.Cc)
	err := parseUtil(cc, "ead.triples.1.xml")
	is.NoErr(err)

	subject := r.NewResource(NewSubject("test", "did", "123"))
	inputSubject := r.NewResource(NewSubject("test", "did", "123") + "/did")

	trip := func(s r.Term, p, o string) *r.Triple {
		return &r.Triple{
			Subject:   s,
			Predicate: r.NewResource(fmt.Sprintf("https://archief.nl/def/ead/%s", p)),
			Object:    r.NewLiteral(o),
		}
	}

	want := []*r.Triple{
		trip(inputSubject, "unitID", "A"),
		trip(inputSubject, "unitTitle",
			"Spieghel der Zeevaerdt, ... (etc.) door <persname>Lucas Jansz Waghenaer</persname>."),
		trip(inputSubject, "unitDate", "1584-1585."),
		trip(inputSubject, "physdescExtent", "1 deel"),
		trip(inputSubject, "physdescPhysfacet", "Folio"),
		trip(inputSubject, "physdescDimension", "39,5 x 28,5 x 4 cm"),
		trip(inputSubject, "physdesc", "1 katern"),
		trip(inputSubject, "physloc", "Ontbreekt"),
		trip(inputSubject, "materialspec", "Normale geschreven, getypte en gedrukte documenten, geen bijzondere handschriften."),
		trip(inputSubject, "origination", "<corpname>Centrale Dienst voor Sibbekunde</corpname>"),
		trip(inputSubject, "abstract", "Het archief bevat o.a."),
		trip(inputSubject, "langmaterial",
			`Het merendeel der stukken is in het <language langcode="dut" scriptcode="Latn">Nederlands</language>`),
		trip(inputSubject, "dao",
			"http://example.com/format/xml/findingaid/1.11.01.01/file/112"),
	}

	got, err := cc.Cdid[0].Triples(subject)
	is.NoErr(err)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Cdid.Triples() mismatch (-want +got):\n%s", diff)
	}
}

func TestClevelTriples(t *testing.T) {
	is := is.New(t)

	cc := new(ead.Cc)
	err := parseUtil(cc, "ead.triples.2.xml")
	is.NoErr(err)

	subject := r.NewResource(NewSubject("test", "c", "123"))
	didSubject := r.NewResource(NewSubject("test", "c", "123/did"))
	trip := func(s r.Term, p, o string) *r.Triple {
		return &r.Triple{
			Subject:   s,
			Predicate: ead.NewResource(p),
			Object:    r.NewLiteral(o),
		}
	}

	want := []*r.Triple{
		r.NewTriple(
			subject,
			ead.NewResource("hasDid"),
			didSubject,
		),
		r.NewTriple(
			didSubject,
			r.NewResource(fragments.RDFType),
			ead.NewResource("Did"),
		),
		r.NewTriple(
			subject,
			r.NewResource(fragments.RDFType),
			ead.NewResource("Clevel"),
		),
		trip(subject, "persname", "R.W. v. Pabst"),
		trip(subject, "accessrestrict",
			"<head>Openbaarheidsbeperkingen</head> <legalstatus type=\"ABS\">Volledig openbaar.</legalstatus>"),
		trip(subject, "controlaccess",
			"<genreform type=\"filtering\">kaart</genreform>"),
		trip(subject, "odd",
			"<p>Op de titelpagina in pen de handtekening ' <persname>R.W. v. Pabst</persname> '</p>"),
		trip(subject, "persname", "R.W. v. Pabst"),
		trip(subject, "scopecontent",
			"<p>Eerste en tweede deel in &#xE9;&#xE9;n band<lb/>Bevat 44 kaarten in koperdruk</p>"),
		trip(subject, "phystech", "<p>Niet raadpleegbaar</p>"),
		trip(subject, "custodhist", "<p>Niet bekend</p>"),
		trip(subject, "altformavail", "<p>Zie facsimile studiezaal: GEO F 515</p>"),
		trip(subject, "altformavail", "<p>Kaart 1: gefacsimileerd in </p>"),
		trip(subject, "acqinfo", "<p>Aanwinsten: 1866 A IV</p>"),
		trip(subject, "userestrict", "<p>Dit archief wordt gedigitaliseerd.</p>"),
		trip(subject, "accruals", "<p>Deze doos is nog niet overgedragen</p>"),
		trip(subject, "appraisal", "<p> <num type=\"Handeling\">020.2-0017-02</num> </p>"),
		trip(subject, "bioghist",
			"<list type=\"simple\"> <item>Grondslag: Wet overheidsaansprakelijkheid Bezettingshandelingen, art. 7, 8. </item> </list>"),
		trip(subject, "relatedmaterial", "<p>Uit dit deel werd in 1989 blad AKF 16A1 afgezonderd.</p>"),
		trip(subject, "arrangement", "<p>Geordend op jaar van verschijnen.</p>"),
		trip(subject, "separatedmaterial", "<p>Gedeeltelijk (minuut-commissie) overgebracht</p>"),
		trip(subject, "processinfo", "<p>Details van deze beschrijvingen zijn in de index op zaaknamen verwerkt.</p>"),
		trip(subject, "otherfindaid", "<p>Lijsten van Aangenomen Manschappen<lb/>1814-1829 [1904-1906] (oud nr. 2.12.09).</p>"),
		trip(subject, "originalsloc", "<p>TNA CO 116/67</p>"),
		trip(subject, "fileplan", "<p>example</p>"),
		trip(subject, "descgrp", "<p>example</p>"),
	}

	got, err := cc.Triples(subject)
	is.NoErr(err)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Clevel.Triples() mismatch (-want +got):\n%s", diff)
	}
}
