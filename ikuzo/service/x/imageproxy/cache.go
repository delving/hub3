package imageproxy

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/OneOfOne/xxhash"
	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidCacheKey  = errors.New("invalid cache key; can't be decoded to URL")
	ErrCacheKeyNotFound = errors.New("cache key is not found")
)

type RequestOption func(req *Request) error

func SetRawQueryString(queryString string) RequestOption {
	return func(req *Request) error {
		req.rawQueryString = queryString
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
		req.transformOptions = options
		return nil
	}
}

type Request struct {
	cacheKey         string
	sourceURL        string
	cached           bool
	transformOptions string
	rawQueryString   string
	s                *Service
}

func NewRequest(input string, options ...RequestOption) (*Request, error) {
	req := &Request{
		cacheKey: input,
	}

	// apply options
	for _, option := range options {
		if err := option(req); err != nil {
			return nil, err
		}
	}

	if strings.HasPrefix(req.cacheKey, "http") {
		req.sourceURL = req.cacheKey

		if req.rawQueryString != "" {
			req.sourceURL = fmt.Sprintf("%s?%s", req.sourceURL, req.rawQueryString)
		}

		if strings.Contains(req.sourceURL, "&amp;") {
			req.sourceURL = strings.ReplaceAll(req.sourceURL, "&amp;", "&")
		}

		req.cacheKey = encodeURL(req.sourceURL)
	} else {
		sourceURL, err := decodeURL(req.cacheKey)
		if err != nil {
			log.Warn().Err(err).Msg("unable to decode cache key")
			return nil, ErrInvalidCacheKey
		}

		req.sourceURL = sourceURL
	}

	return req, nil
}

// GET returns a *http.Request
func (req *Request) GET() (*http.Request, error) {
	return http.NewRequest("GET", req.sourceURL, nil)
}

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

func (req *Request) Write(path string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("unable to create directories; %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create file; %w", err)
	}

	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("unable to write to file; %w", err)
	}

	return nil
}

// sourcePath is the relative path to the source
func (req *Request) sourcePath() string {
	hash := fmt.Sprintf("%016x", xxhash.ChecksumString64(req.sourceURL))

	return filepath.Join(
		hash[0:3],
		hash[3:6],
		hash[6:9],
		req.cacheKey,
	)
}

func (req *Request) derivativePath() string {
	if req.transformOptions == "" {
		return ""
	}

	return filepath.Join(
		req.sourcePath()+"#",
		req.transformOptions,
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
