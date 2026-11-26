package fragments

import "testing"

func TestNormalizeLatLong(t *testing.T) {
	ok := map[string]string{
		"SRID=4326;POINT (5.141153 51.72828)":                  "51.72828,5.141153", // MIP (GeoDjango EWKT)
		"Point(5.7497 51.4207)":                                "51.4207,5.7497",    // brabantse-gebouwen (WKT)
		"POINT(-73.98 40.75)":                                  "40.75,-73.98",
		`[{"type":"Point","coordinates":[5.70458356,51.6]}]`:   "51.6,5.70458356", // GeoJSON, a list of one
		`{"type": "Point", "coordinates": [5.70458356, 51.6]}`: "51.6,5.70458356",
		"51.4207,5.7497":                                       "51.4207,5.7497", // already the mapping's lat,lon
		" 51.4207, 5.7497 ":                                    "51.4207,5.7497",
	}
	for in, want := range ok {
		if got, valid := normalizeLatLong(in); !valid || got != want {
			t.Errorf("normalizeLatLong(%q) = %q, %v; want %q", in, got, valid, want)
		}
	}
	for _, in := range []string{"", "POINT EMPTY", "POINT (5.1)", "POINT (200 51)", "POINT (5 95)", "LINESTRING (5 51, 6 52)",
		`{"type":"Polygon","coordinates":[]}`, "x,y", "91,5", "SRID=4326;"} {
		if got, valid := normalizeLatLong(in); valid {
			t.Errorf("normalizeLatLong(%q) = %q, want refused", in, got)
		}
	}
}

func TestCreateDateRangeMore(t *testing.T) {
	ok := map[string]IndexRange{
		// years joined by a hyphen (MIP "1850-1870", "ca. 1850-1870")
		"1980-1985":     {"1980-01-01", "1985-12-31"},
		"1980 - 1985":   {"1980-01-01", "1985-12-31"},
		"ca. 1850-1870": {"1850-01-01", "1870-12-31"},
		"1980 – 1990":   {"1980-01-01", "1990-12-31"},
		// centuries: XIX is 1800-1899; a-d its quarters, A-B its halves (MIP)
		"XIX":               {"1800-01-01", "1899-12-31"},
		"XIXd":              {"1875-01-01", "1899-12-31"},
		"XIXa":              {"1800-01-01", "1824-12-31"},
		"XIXA":              {"1800-01-01", "1849-12-31"},
		"XIXB":              {"1850-01-01", "1899-12-31"},
		"XVIII":             {"1700-01-01", "1799-12-31"},
		"XXa":               {"1900-01-01", "1924-12-31"},
		"ca. XVII":          {"1600-01-01", "1699-12-31"},
		"19e eeuw":          {"1800-01-01", "1899-12-31"},
		"19e":               {"1800-01-01", "1899-12-31"},
		"20a":               {"1900-01-01", "1924-12-31"},
		"XIX B":             {"1850-01-01", "1899-12-31"},
		"XVIII-XIX":         {"1700-01-01", "1899-12-31"}, // a span of centuries
		"XVIII;XIX":         {"1700-01-01", "1899-12-31"},
		"XVIII of XIX":      {"1700-01-01", "1899-12-31"},
		"18e-19e eeuw":      {"1700-01-01", "1899-12-31"},
		"4e kwart 19e eeuw": {"1875-01-01", "1899-12-31"},
		"1e helft 19e eeuw": {"1800-01-01", "1849-12-31"},
		"2e helft 20e eeuw": {"1950-01-01", "1999-12-31"},
		// day first, as Dutch archives write them (birth, death and burial dates)
		"26-10-1920":              {"1920-10-26", "1920-10-26"},
		"6-5-1925":                {"1925-05-06", "1925-05-06"},
		"31/12/1938":              {"1938-12-31", "1938-12-31"},
		"01-1997":                 {"1997-01-01", "1997-01-31"},
		"2-1996":                  {"1996-02-01", "1996-02-28"},
		"09-2013 – 01-2014":       {"2013-09-01", "2014-01-31"},
		"23-03-2012 – 01-04-2012": {"2012-03-23", "2012-04-01"},
		// unchanged
		"1980":                  {"1980-01-01", "1980-12-31"},
		"1793-05-13/1794-01-11": {"1793-05-13", "1794-01-11"},
		"1793-05/1794":          {"1793-05-01", "1794-12-31"},
	}
	for in, want := range ok {
		got, err := CreateDateRange(in)
		if err != nil || got != want || !got.isDates() {
			t.Errorf("CreateDateRange(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, in := range []string{"Z.d.", "ongeveer", "XXX", "XXII", "19", "IIX9", "1985-1980", "XIX-XVIII", "5e kwart 19e eeuw",
		"3e helft 19e eeuw", "eind 19e eeuw", "XIXm", "31-02-1945", "13-1997", "00-1997", "32-01-1945", "01-04-2012 – 23-03-2012"} {
		if got, err := CreateDateRange(in); err == nil && got.isDates() {
			t.Errorf("CreateDateRange(%q) = %v, want refused", in, got)
		}
	}
}

func TestCreateDateRangeEDTF(t *testing.T) {
	ok := map[string]IndexRange{
		// EDTF qualifiers: uncertain (?), approximate (~), both (%) keep their date
		"1850~":       {"1850-01-01", "1850-12-31"},
		"1850?":       {"1850-01-01", "1850-12-31"},
		"1850%":       {"1850-01-01", "1850-12-31"},
		"1850-05~":    {"1850-05-01", "1850-05-31"},
		"1850~/1870?": {"1850-01-01", "1870-12-31"},
		// EDTF unspecified digits: the decade, century or millennium they leave open
		"185X":       {"1850-01-01", "1859-12-31"},
		"18XX":       {"1800-01-01", "1899-12-31"},
		"1XXX":       {"1000-01-01", "1999-12-31"},
		"1850-XX":    {"1850-01-01", "1850-12-31"},
		"1850-05-XX": {"1850-05-01", "1850-05-31"},
		"18XX/19XX":  {"1800-01-01", "1999-12-31"},
		"1850/1870":  {"1850-01-01", "1870-12-31"},
	}
	for in, want := range ok {
		got, err := CreateDateRange(in)
		if err != nil || got != want || !got.isDates() {
			t.Errorf("CreateDateRange(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	// open-ended: no end (or start) to index, so no range rather than the wrong year
	for _, in := range []string{"1850/..", "../1850", "1850/", "/1850", "na 1850", "Na 1850", "voor 1850", "vóór 1850",
		"vanaf 1850", "sinds 1850", "tot 1850", "after 1850", "before 1850", "-0800", "XXXX"} {
		if got, err := CreateDateRange(in); err == nil && got.isDates() {
			t.Errorf("CreateDateRange(%q) = %v, want refused", in, got)
		}
	}
}
