// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package server imageproxy proxies requests for remote images
package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/die-net/lrucache"
	"github.com/die-net/lrucache/twotier"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/peterbourgon/diskv"
	"willnorris.com/go/imageproxy"

	c "bitbucket.org/delving/rapid/config"
)

const defaultMemorySize = 100

var p *imageproxy.Proxy

// setup proxy in init function
func init() {
	u, err := url.Parse("memory:200:1h")
	if err != nil {
		log.Fatal(err)
	}
	memCache, err := lruCache(u.Opaque)
	if err != nil {
		log.Fatal(err)
	}
	fileCache := diskCache(c.Config.ImageProxy.CacheDir)
	cache := twotier.New(memCache, fileCache)
	p = imageproxy.NewProxy(nil, cache)
	p.Whitelist = c.Config.ImageProxy.Whitelist
	p.ScaleUp = c.Config.ImageProxy.ScaleUp
}

// create handler fuction to serve the proxied images
func serveProxyImage(w http.ResponseWriter, r *http.Request) {
	req, err := imageproxy.NewRequest(r, nil)
	if err != nil {
		msg := fmt.Sprintf("invalid request URL: %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	resp, err := p.Client.Get(req.String())
	if err != nil {
		msg := fmt.Sprintf("error fetching remote image: %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	//cached := resp.Header.Get(httpcache.XFromCache)
	//if p.Verbose {
	//log.Printf("request: %v (served from cache: %v)", *req, cached == "1")
	//}
	copyHeader(w.Header(), resp.Header, "Cache-Control", "Last-Modified", "Expires", "Etag", "Link")

	if should304(r, resp) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	copyHeader(w.Header(), resp.Header, "Content-Length", "Content-Type")

	//Enable CORS for 3rd party applications
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		msg := fmt.Sprintf("error copying remote image: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	return
}

// copyHeader copies header values from src to dst, adding to any existing
// values with the same header name.  If keys is not empty, only those header
// keys will be copied.
func copyHeader(dst, src http.Header, keys ...string) {
	if len(keys) == 0 {
		for k := range src {
			keys = append(keys, k)
		}
	}
	for _, key := range keys {
		k := http.CanonicalHeaderKey(key)
		for _, v := range src[k] {
			dst.Add(k, v)
		}
	}
}

// should304 returns whether we should send a 304 Not Modified in response to
// req, based on the response resp.  This is determined using the last modified
// time and the entity tag of resp.
func should304(req *http.Request, resp *http.Response) bool {
	// TODO(willnorris): if-none-match header can be a comma separated list
	// of multiple tags to be matched, or the special value "*" which
	// matches all etags
	etag := resp.Header.Get("Etag")
	if etag != "" && etag == req.Header.Get("If-None-Match") {
		return true
	}

	lastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		return false
	}
	ifModSince, err := time.Parse(time.RFC1123, req.Header.Get("If-Modified-Since"))
	if err != nil {
		return false
	}
	if lastModified.Before(ifModSince) || lastModified.Equal(ifModSince) {
		return true
	}

	return false
}

// lruCache creates an LRU Cache with the specified options of the form
// "maxSize:maxAge".  maxSize is specified in megabytes, maxAge is a duration.
func lruCache(options string) (*lrucache.LruCache, error) {
	parts := strings.SplitN(options, ":", 2)
	size, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}

	var age time.Duration
	if len(parts) > 1 {
		age, err = time.ParseDuration(parts[1])
		if err != nil {
			return nil, err
		}
	}

	return lrucache.New(size*1e6, int64(age.Seconds())), nil
}

func diskCache(path string) *diskcache.Cache {
	d := diskv.New(diskv.Options{
		BasePath: path,

		// For file "c0ffee", store file as "c0/ff/c0ffee"
		Transform: func(s string) []string { return []string{s[0:2], s[2:4]} },
	})
	return diskcache.NewWithDiskv(d)
}

type tieredCache struct {
	imageproxy.Cache
}

func (tc *tieredCache) String() string {
	return fmt.Sprint(*tc)
}

func (tc *tieredCache) Set(value string) error {
	c, err := parseCache(value)
	if err != nil {
		return err
	}

	if tc.Cache == nil {
		tc.Cache = c
	} else {
		tc.Cache = twotier.New(tc.Cache, c)
	}
	return nil
}

// parseCache parses c returns the specified Cache implementation.
func parseCache(c string) (imageproxy.Cache, error) {
	if c == "" {
		return nil, nil
	}

	if c == "memory" {
		c = fmt.Sprintf("memory:%d", defaultMemorySize)
	}

	u, err := url.Parse(c)
	if err != nil {
		return nil, fmt.Errorf("error parsing cache flag: %v", err)
	}

	switch u.Scheme {
	case "memory":
		return lruCache(u.Opaque)
	case "file":
		fallthrough
	default:
		return diskCache(u.Path), nil
	}
}
