package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/BurntSushi/toml"
)

// virtualCollection is a nave virtual collection in the form nave stores it:
// query and each orquery are `"field:value";;;"field:value"` strings.
type virtualCollection struct {
	Slug      string   `toml:"slug"`
	Title     string   `toml:"title"`
	Query     string   `toml:"query"`
	OrQueries []string `toml:"orqueries"`
}

func loadVC(path string) (*virtualCollection, error) {
	var vc virtualCollection
	if _, err := toml.DecodeFile(path, &vc); err != nil {
		return nil, fmt.Errorf("unable to read virtual collection %s: %w", path, err)
	}

	return &vc, nil
}

// filterGroups converts the stored queries into groups of v2 qf filters.
// Within a group nave ORs the values of one field and ANDs the fields, which v2 qf does too.
// The query and the orquery are ORed, which v2 cannot express, so each group is harvested on its own.
// Nave only applies the first orquery (_filters_as_dict flattens the list with filters[0]).
func (vc *virtualCollection) filterGroups() ([][]string, error) {
	raws := []string{vc.Query}
	if len(vc.OrQueries) > 0 {
		raws = append(raws, vc.OrQueries[0])
	}
	if len(vc.OrQueries) > 1 {
		slog.Warn("nave ignores all but the first orquery; so does this harvest", "vc", vc.Slug, "orqueries", len(vc.OrQueries))
	}

	groups := make([][]string, 0, len(raws))
	for _, raw := range raws {
		group, err := naveToV2Filters(raw)
		if err != nil {
			return nil, fmt.Errorf("virtual collection %s: %w", vc.Slug, err)
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// naveToV2Filters converts one stored nave query into v2 qf filters, normalising
// the field names the way nave's _filters_as_dict and query_to_facet_key do.
func naveToV2Filters(raw string) ([]string, error) {
	var qfs []string
	for _, f := range strings.Split(raw, ";;;") {
		f = strings.Trim(strings.TrimSpace(f), `"`)
		if f == "" {
			continue
		}

		key, value, ok := strings.Cut(f, ":")
		if !ok || strings.HasPrefix(f, "(") || strings.Contains(f, " OR ") {
			return nil, fmt.Errorf("query string filter %q is not supported", f)
		}

		exclude := strings.HasPrefix(key, "-")
		key = strings.NewReplacer("_facet", "", "_string", "", "_text", "").Replace(strings.TrimLeft(key, "-+"))

		switch {
		case key == "delving_spec":
			key = "meta.spec"
		case strings.HasPrefix(key, "delving_"):
			return nil, fmt.Errorf("legacy filter field %q is not supported", key)
		}

		if exclude {
			key = "-" + key
		}

		qfs = append(qfs, key+":"+strings.ReplaceAll(value, `"`, "'"))
	}

	if len(qfs) == 0 {
		return nil, fmt.Errorf("query %q has no filters", raw)
	}

	return qfs, nil
}

// groupURLs adds each filter group as qf[] parameters to the v2 search URL.
func groupURLs(apiURL string, groups [][]string) ([]string, error) {
	urls := make([]string, 0, len(groups))
	for _, group := range groups {
		u, err := url.Parse(apiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}

		params := u.Query()
		for _, qf := range group {
			params.Add("qf[]", qf)
		}
		u.RawQuery = params.Encode()
		urls = append(urls, u.String())
	}

	return urls, nil
}
