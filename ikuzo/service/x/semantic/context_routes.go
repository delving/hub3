package semantic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	defaultContextPrefix = "/contexts"
	ldJSONContentType    = "application/ld+json; charset=utf-8"
)

// ContextRouter returns a chi router function that serves embedded JSON-LD contexts.
func ContextRouter(fsys fs.FS, mount string, filenames []string) func(r chi.Router) {
	if mount == "" {
		mount = defaultContextPrefix
	}

	if !strings.HasPrefix(mount, "/") {
		mount = "/" + mount
	}

	allowed := make(map[string]struct{}, len(filenames))
	names := make([]string, 0, len(filenames))

	for _, name := range filenames {
		cleaned := path.Clean(name)
		if cleaned != name {
			continue
		}
		allowed[name] = struct{}{}
		names = append(names, name)
	}

	sort.Strings(names)

	basePath := strings.TrimSuffix(mount, "/")
	if basePath == "" {
		basePath = "/"
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, basePath)
		name = strings.TrimPrefix(name, "/")
		if name == "" {
			http.NotFound(w, r)
			return
		}

		if _, ok := allowed[name]; !ok {
			http.NotFound(w, r)
			return
		}

		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		host := fmt.Sprintf("%s://%s", getScheme(r), r.Host)
		contextBase := strings.TrimSuffix(host+mount, "/")

		data = bytes.ReplaceAll(data, []byte("{{BASE_URL}}"), []byte(host))
		data = bytes.ReplaceAll(data, []byte("{{CONTEXT_BASE}}"), []byte(contextBase))

		w.Header().Set("Content-Type", ldJSONContentType)
		w.Header().Set("Cache-Control", "public, max-age=3600")

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}

		_, _ = w.Write(data)
	}

	index := func(w http.ResponseWriter, r *http.Request) {
		base := mount
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}

		scheme := getScheme(r)
		host := r.Host
		if host == "" {
			host = "localhost"
		}
		fullBase := fmt.Sprintf("%s://%s%s", scheme, host, base)

		type fileEntry struct {
			Name string
			URL  string
		}

		type versionInfo struct {
			Version string
			Files   []fileEntry
		}

		type contextSummary struct {
			Versions  map[string]*versionInfo
			Resources []fileEntry
		}

		grouped := make(map[string]*contextSummary)

		members := make([]map[string]any, 0, len(names))

		for _, name := range names {
			url := fullBase + name
			parts := strings.Split(name, "/")
			ctxName := parts[0]

			summary, ok := grouped[ctxName]
			if !ok {
				summary = &contextSummary{Versions: make(map[string]*versionInfo)}
				grouped[ctxName] = summary
			}

			if len(parts) >= 3 {
				version := parts[1]
				fileName := strings.Join(parts[2:], "/")

				vi, ok := summary.Versions[version]
				if !ok {
					vi = &versionInfo{Version: version}
					summary.Versions[version] = vi
				}

				vi.Files = append(vi.Files, fileEntry{Name: fileName, URL: url})
			} else if len(parts) >= 2 {
				fileName := strings.Join(parts[1:], "/")
				summary.Resources = append(summary.Resources, fileEntry{Name: fileName, URL: url})
			} else {
				summary.Resources = append(summary.Resources, fileEntry{Name: name, URL: url})
			}

			members = append(members, map[string]any{
				"@id":                   url,
				"@type":                 []string{"hub3:ContextDocument"},
				"schema:path":           name,
				"schema:name":           parts[len(parts)-1],
				"schema:url":            url,
				"schema:encodingFormat": "application/ld+json",
			})
		}

		contexts := make([]map[string]any, 0, len(grouped))

		for ctxName, summary := range grouped {
			versionList := make([]map[string]any, 0, len(summary.Versions))
			for version, info := range summary.Versions {
				files := make([]fileEntry, len(info.Files))
				copy(files, info.Files)
				sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

				fileMaps := make([]map[string]string, 0, len(files))
				canonical := ""
				for _, f := range files {
					fileMaps = append(fileMaps, map[string]string{
						"schema:name": f.Name,
						"schema:url":  f.URL,
					})
					if f.Name == "context.jsonld" {
						canonical = f.URL
					}
				}

				entry := map[string]any{
					"schema:version": version,
					"hub3:files":     fileMaps,
				}
				if canonical != "" {
					entry["schema:url"] = canonical
				}

				versionList = append(versionList, entry)
			}

			sort.Slice(versionList, func(i, j int) bool {
				return versionList[i]["schema:version"].(string) > versionList[j]["schema:version"].(string)
			})

			extra := make([]map[string]string, 0, len(summary.Resources))
			sort.Slice(summary.Resources, func(i, j int) bool { return summary.Resources[i].Name < summary.Resources[j].Name })
			for _, res := range summary.Resources {
				extra = append(extra, map[string]string{
					"schema:name": res.Name,
					"schema:url":  res.URL,
				})
			}

			contexts = append(contexts, map[string]any{
				"schema:identifier": ctxName,
				"hub3:versions":     versionList,
				"hub3:resources":    extra,
			})
		}

		sort.Slice(contexts, func(i, j int) bool {
			return contexts[i]["schema:identifier"].(string) < contexts[j]["schema:identifier"].(string)
		})

		response := map[string]any{
			"@context": map[string]string{
				"hydra":  "http://www.w3.org/ns/hydra/core#",
				"schema": "http://schema.org/",
				"hub3":   "https://hub3.delving.org/vocab/",
			},
			"@id":               fmt.Sprintf("%s://%s%s", scheme, host, mount),
			"@type":             []string{"hydra:Collection"},
			"hydra:title":       "Available JSON-LD Contexts",
			"hydra:description": "Collection of JSON-LD @context documents for use with the semantic API",
			"totalItems":        len(names),
			"member":            members,
			"hub3:contexts":     contexts,
		}

		w.Header().Set("Content-Type", ldJSONContentType)
		json.NewEncoder(w).Encode(response) //nolint:errcheck // response writer
	}

	return func(r chi.Router) {
		r.Route(mount, func(cr chi.Router) {
			cr.Get("/", index)
			cr.Get("/*", handler)
			cr.Head("/*", handler)
		})
	}
}
