package imageproxy

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/OneOfOne/xxhash"
	"github.com/rs/zerolog/log"
)

type cacheKey struct{}

type ProxyOption string

const (
	Raw          ProxyOption = "raw"
	DeepZoom     ProxyOption = "deepzoom"
	Explain      ProxyOption = "explain"
	ProxyRequest ProxyOption = "request"
	Thumbnail    ProxyOption = "thumbnail"
)

type CacheType string

const (
	Source             CacheType = "source"
	Cache              CacheType = "cache"
	Lru                CacheType = "lrucache"
	RejectDomain       CacheType = "domain_rejected"
	RejectReferrer     CacheType = "referrer_rejected"
	TargetRequestError CacheType = "remote_request_error"
)

var (
	ErrInvalidCacheKey  = errors.New("invalid cache key; can't be decoded to URL")
	ErrCacheKeyNotFound = errors.New("cache key is not found")
	deepZoomSuffix      = ".dzi"
)

type RequestOption func(req *Request) error

func SetRawQueryString(queryString string) RequestOption {
	return func(req *Request) error {
		req.RawQueryString = queryString
		return nil
	}
}

func SetService(s *Service) RequestOption {
	return func(req *Request) error {
		req.s = s
		return nil
	}
}

func SetTransform(options string) RequestOption {
	return func(req *Request) error {
		req.TransformOptions = options
		return nil
	}
}

func SetEnableTransform(enabled bool) RequestOption {
	return func(req *Request) error {
		req.EnableTransform = enabled
		return nil
	}
}

type Request struct {
	SourceURL        string    // the request remote URL
	CacheKey         string    // the normalised cache key for both storage and retrieval
	TransformOptions string    // options to transform images or trigger actions
	RawQueryString   string    // not sure why this is needed
	CacheType        CacheType // CacheType is how the data is returned
	SubPath          string    // subPath is appended to raw cacheKey to get derivatives
	thumbnailOpts    string
	EnableTransform  bool
	s                *Service
}

func NewRequest(input string, options ...RequestOption) (*Request, error) {
	req := &Request{
		CacheKey: input,
	}

	// apply options
	for _, option := range options {
		if err := option(req); err != nil {
			return nil, err
		}
	}

	if !req.EnableTransform {
		switch {
		case req.TransformOptions == "deepzoom":
			req.TransformOptions = "raw"
		case strings.HasSuffix(req.TransformOptions, ",smartcrop"):
			req.TransformOptions = "raw"
		}
	}

	if strings.HasPrefix(req.CacheKey, "http") {
		if !strings.HasPrefix(req.CacheKey, "https://") && !strings.HasPrefix(req.CacheKey, "http://") {
			req.CacheKey = strings.ReplaceAll(input, ":/", "://")
		}
		req.SourceURL = req.CacheKey

		if req.RawQueryString != "" {
			req.SourceURL = fmt.Sprintf("%s?%s", req.SourceURL, req.RawQueryString)
		}

		if strings.Contains(req.SourceURL, "&amp;") {
			req.SourceURL = strings.ReplaceAll(req.SourceURL, "&amp;", "&")
		}

		if strings.HasSuffix(req.SourceURL, deepZoomSuffix) && req.TransformOptions == "deepzoom" {
			req.SourceURL = strings.TrimSuffix(req.SourceURL, deepZoomSuffix)
			req.SubPath = deepZoomSuffix
		}

		if strings.Contains(req.SourceURL, "_files/") && strings.HasSuffix(req.SourceURL, ".jpeg") {
			parts := strings.Split(req.SourceURL, "_files/")
			req.SourceURL = parts[0]
			req.SubPath = "_files/" + parts[len(parts)-1]
		}

		if strings.HasSuffix(req.TransformOptions, ",smartcrop") {
			req.thumbnailOpts = strings.TrimSuffix(req.TransformOptions, ",smartcrop")
			req.SubPath = "_" + req.TransformOptions + "_tn.jpg"
		}

		req.CacheKey = encodeURL(req.SourceURL)

		if req.SubPath != "" {
			req.CacheKey += req.SubPath
		}
	} else {
		sourceURL, err := decodeURL(req.CacheKey)
		if err != nil {
			log.Warn().
				Err(err).
				Str("cacheKey", input).
				Msg("unable to decode cache key")
			return nil, ErrInvalidCacheKey
		}

		req.SourceURL = sourceURL
	}

	return req, nil
}

// GET returns a *http.Request for the sourceURL
func (req *Request) GET() (*http.Request, error) {
	return http.NewRequest("GET", req.SourceURL, http.NoBody)
}

func (req *Request) Remove() error {
	targetDir := path.Dir(req.downloadedSourcePath())

	files, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	cacheBase := encodeURL(req.SourceURL)

	for _, file := range files {
		if strings.HasPrefix(file.Name(), cacheBase) {
			if err := os.RemoveAll(filepath.Join(targetDir, file.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

// Read returns an io.ReadCloser found at path from the cache.
// ErrCacheKeyNotFound is returned when nothing is found.
func (req *Request) Read(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCacheKeyNotFound
		}

		return nil, err
	}

	return f, nil
}

// Write writes all content of the reader to path in the cache
func (req *Request) Write(path string, r io.Reader) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return 0, fmt.Errorf("unable to create directories; %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("unable to create file; %w", err)
	}

	size, err := io.Copy(f, r)
	if err != nil {
		return 0, fmt.Errorf("unable to write to file; %w", err)
	}

	return size, nil
}

// existsInCache returns whether a path is stored in the cache
func existsInCache(path string) (info os.FileInfo, present bool) {
	file, err := os.Stat(path)
	return file, !errors.Is(err, os.ErrNotExist)
}

// cacheKeyPath is the full path to the cached resource
func (req *Request) cacheKeyPath() string {
	hash := fmt.Sprintf("%016x", xxhash.ChecksumString64(req.SourceURL))

	return filepath.Join(
		req.s.cacheDir,
		hash[0:3],
		hash[3:6],
		hash[6:9],
		req.CacheKey,
	)
}

// sourcePath is the full path to the downloaded source
func (req *Request) downloadedSourcePath() string {
	hash := fmt.Sprintf("%016x", xxhash.ChecksumString64(req.SourceURL))

	return filepath.Join(
		req.s.cacheDir,
		hash[0:3],
		hash[3:6],
		hash[6:9],
		encodeURL(req.SourceURL),
	)
}

func encodeURL(input string) string {
	return base64.URLEncoding.EncodeToString([]byte(input))
}

func decodeURL(input string) (string, error) {
	b, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}

	sourceURL := string(b)
	if !strings.HasPrefix(sourceURL, "http") {
		return "", ErrInvalidCacheKey
	}

	return sourceURL, nil
}
