package fragments

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// The mapping's latLong is a geo_point: OpenSearch refuses the whole record
// over a value it cannot parse (GeoDjango's EWKT "SRID=4326;POINT (…)"), and
// reads a bare "x,y" as lat,lon. normalizeLatLong gives the "lat,lon" form.

var (
	srid      = regexp.MustCompile(`(?i)^SRID=\d+;`)
	wktPoint  = regexp.MustCompile(`(?i)^POINT\s*\(\s*(\S+)\s+(\S+)\s*\)$`)
	plainPair = regexp.MustCompile(`^\s*(\S+?)\s*,\s*(\S+?)\s*$`)
)

type geoJSONPoint struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// normalizeLatLong reads a WKT or EWKT POINT (lon lat), a GeoJSON Point (or a
// list of one) or a "lat,lon" pair, and gives "lat,lon"; ok is false for
// anything else or out of range.
func normalizeLatLong(v string) (string, bool) {
	v = srid.ReplaceAllString(strings.TrimSpace(v), "")
	var lat, lon string
	switch {
	case wktPoint.MatchString(v):
		m := wktPoint.FindStringSubmatch(v)
		lon, lat = m[1], m[2]
	case strings.HasPrefix(v, "{") || strings.HasPrefix(v, "["):
		var p geoJSONPoint
		if strings.HasPrefix(v, "[") {
			var ps []geoJSONPoint
			if json.Unmarshal([]byte(v), &ps) != nil || len(ps) != 1 {
				return "", false
			}
			p = ps[0]
		} else if json.Unmarshal([]byte(v), &p) != nil {
			return "", false
		}
		if p.Type != "Point" || len(p.Coordinates) != 2 {
			return "", false
		}
		lon = strconv.FormatFloat(p.Coordinates[0], 'f', -1, 64)
		lat = strconv.FormatFloat(p.Coordinates[1], 'f', -1, 64)
	case plainPair.MatchString(v):
		m := plainPair.FindStringSubmatch(v)
		lat, lon = m[1], m[2]
	default:
		return "", false
	}
	la, errLat := strconv.ParseFloat(lat, 64)
	lo, errLon := strconv.ParseFloat(lon, 64)
	if errLat != nil || errLon != nil || la < -90 || la > 90 || lo < -180 || lo > 180 {
		return "", false
	}
	return lat + "," + lon, true
}

// Dates the year parser does not read on its own.
var (
	// "1980-1985", "ca. 1850 - 1870": a hyphen between two years is a period, not a month
	yearHyphenYear = regexp.MustCompile(`(\d{4})\s*-\s*(\d{4})`)
	// "XIX", "XIXd", "ca. XVII", "19e eeuw", "20a": a century, a-d its quarters, A-B its halves (MIP)
	romanCentury = regexp.MustCompile(`^(?:ca\.?\s*)?([IVX]{1,5})\s?([a-dAB])?$`)
	// "4e kwart 19e eeuw", "2e helft 19e eeuw"
	centuryPart = regexp.MustCompile(`(?i)^(?:ca\.?\s*)?(\d)e\s+(kwart|helft)\s+(\d{1,2})e\s+eeuw$`)
	// "XVIII-XIX", "XVIII;XIX", "XVIII of XIX", "18e-19e eeuw": from the first century to the second
	centurySpan   = regexp.MustCompile(`^(.+?)\s*(?:-|;|\bof\b)\s*(.+)$`)
	arabicCentury = regexp.MustCompile(`^(?:ca\.?\s*)?(\d{1,2})(?:e(?:\s*eeuw)?|([a-dAB]))$`)
	romanValues   = map[byte]int{'I': 1, 'V': 5, 'X': 10}
	// "26-10-1920", "6-5-1925", "31/12/1938", "01-1997": day (and month) first, as Dutch archives write dates
	// (not inside an ISO date: "1793-05/1794" stays as it is)
	dayFirst = regexp.MustCompile(`(^|[^-/\d])(?:(\d{1,2})[-/])?(\d{1,2})[-/](\d{4})\b`)
)

// isoDayFirst rewrites day-first dates in s as yyyy-MM-dd, or yyyy-MM without a
// day, so the year parser reads them. Whether the date exists is checked later.
func isoDayFirst(s string) string {
	return dayFirst.ReplaceAllStringFunc(s, func(d string) string {
		m := dayFirst.FindStringSubmatch(d)
		if m[2] == "" {
			return fmt.Sprintf("%s%s-%02s", m[1], m[4], m[3])
		}
		return fmt.Sprintf("%s%s-%02s-%02s", m[1], m[4], m[3], m[2])
	})
}

// centuryRange reads s as a century, a quarter or half of one, or a span of two.
func centuryRange(s string) (IndexRange, bool) {
	s = strings.TrimSpace(s)
	if m := centuryPart.FindStringSubmatch(s); m != nil {
		i, _ := strconv.Atoi(m[1])
		parts := "abcd"
		if strings.EqualFold(m[2], "helft") {
			parts = "AB"
		}
		if i < 1 || i > len(parts) {
			return IndexRange{}, false
		}
		return centuryOne(m[3] + string(parts[i-1]))
	}
	if ir, ok := centuryOne(s); ok {
		return ir, true
	}
	if m := centurySpan.FindStringSubmatch(s); m != nil {
		first, ok1 := centuryOne(strings.TrimSuffix(strings.TrimSpace(m[1]), " eeuw"))
		last, ok2 := centuryOne(m[2])
		if ok1 && ok2 && first.Greater <= last.Greater {
			return IndexRange{Greater: first.Greater, Less: last.Less}, true
		}
	}
	return IndexRange{}, false
}

// centuryOne reads s as one century, or a quarter or half of it.
func centuryOne(s string) (IndexRange, bool) {
	s = strings.TrimSpace(s)
	var n int
	var part string
	if m := romanCentury.FindStringSubmatch(s); m != nil {
		n, part = roman(m[1]), m[2]
	} else if m := arabicCentury.FindStringSubmatch(s); m != nil {
		n, _ = strconv.Atoi(m[1])
		part = m[2]
	}
	if n < 1 || n > 21 {
		return IndexRange{}, false
	}
	from, span := (n-1)*100, 100
	switch {
	case part >= "a" && part <= "d":
		from, span = from+25*int(part[0]-'a'), 25
	case part == "A" || part == "B":
		from, span = from+50*int(part[0]-'A'), 50
	}
	return IndexRange{Greater: fmt.Sprintf("%04d-01-01", from), Less: fmt.Sprintf("%04d-12-31", from+span-1)}, true
}

// roman is the value of a numeral of I, V and X; 0 when it is not well formed.
func roman(s string) int {
	total := 0
	for i := 0; i < len(s); i++ {
		v := romanValues[s[i]]
		if i+1 < len(s) && v < romanValues[s[i+1]] {
			total -= v
		} else {
			total += v
		}
	}
	if toRoman(total) != s {
		return 0
	}
	return total
}

func toRoman(n int) string {
	var b strings.Builder
	for _, d := range []struct {
		v int
		s string
	}{{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"}} {
		for n >= d.v {
			b.WriteString(d.s)
			n -= d.v
		}
	}
	return b.String()
}
