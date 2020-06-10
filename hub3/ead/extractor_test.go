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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtractor_NewExtractor(t *testing.T) {
	text := `Teerste Deel Vande Spieghel der Zeevaerdt, vande navigatie der
	<geogname>Westersche Zee</geogname>Extractor Innehoudende alle de Custen van
	<geogname>Vranckrijck</geogname>   <geogname>Spaignen</geogname>  ende
	t&apos;principaelste deel van <geogname>Engelandt</geogname>, in diversche Zee
	Caerten begrepen, met den gebruijcke van dien, nu met grooter naersticheijt
	bij een vergadert ende ghepractizeert, Door <persname>Lucas Iansz Waghenaer</persname>
	Piloot ofte Stuijrman Residerende Inde vermaerde Zeestadt <geogname>Enchuijsen</geogname>.
	Cum Privilegio ad decennium, Reg. Ma.is et Cancellarie Brabantie. 1583.<lb/>Ghedruct tot
	<geogname>Leyden</geogname>  by <persname>Christoffel Plantijn</persname>
	voor <persname>Lucas Janssz Waghenaer van Enckhuysen</persname>. Anno M.D.LXXXIIII.`

	type args struct {
		input []byte
	}

	tests := []struct {
		name    string
		args    args
		want    []NLPToken
		wantErr bool
	}{
		{
			"no extract",
			args{input: []byte("Westersche Zee")},
			[]NLPToken{},
			false,
		},
		{
			"geo extract",
			args{input: []byte("something <geogname>Westersche Zee</geogname>,")},
			[]NLPToken{
				{Text: "Westersche Zee", Type: GeoLocation},
			},
			false,
		},
		{
			"persname extract",
			args{input: []byte("Door <persname>Lucas Iansz Waghenaer</persname>")},
			[]NLPToken{
				{Text: "Lucas Iansz Waghenaer", Type: Person},
			},
			false,
		},
		{
			"date extract",
			args{input: []byte("<date calendar=\"gregorian\" era=\"ce\" normal=\"1581\">1581</date>")},
			[]NLPToken{
				{Text: "1581", Type: DateText},
				{Text: "1581", Type: DateIso},
			},
			false,
		},
		{
			"mixed tags",
			args{input: []byte(text)},
			[]NLPToken{
				{Type: GeoLocation, Text: "Westersche Zee"},
				{Type: GeoLocation, Text: "Vranckrijck"},
				{Type: GeoLocation, Text: "Spaignen"},
				{Type: GeoLocation, Text: "Engelandt"},
				{Type: GeoLocation, Text: "Enchuijsen"},
				{Type: GeoLocation, Text: "Leyden"},
				{Text: "Lucas Iansz Waghenaer", Type: Person},
				{Type: Person, Text: "Christoffel Plantijn"},
				{Type: Person, Text: "Lucas Janssz Waghenaer van Enckhuysen"},
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := NewExtractor(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extractor.Extract() %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got.Tokens()); diff != "" {
				t.Errorf("Extractor.Extract() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
