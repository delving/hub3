// Copyright 2020 Delving B.V.
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

package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CustomTransport struct {
	*http.Transport
}

// RoundTrip implements the http.RoundTripper interface
// This is a workaround to support both Elasticsearch and OpenSearch
func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "elasticsearch/8.16.0")
	req.Header.Set("X-Elastic-Client-Meta", "es=8.16.0,go=1.23.3,t=8.6.0")

	if req.Method == "HEAD" && req.URL.Path == "/" {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: http.Header{
				"X-Found":           []string{"true"},
				"Server":            []string{"elasticsearch/8.16.0"},
				"X-Elastic-Product": []string{"Elasticsearch"},
			},
			Body:          io.NopCloser(bytes.NewReader([]byte{})),
			ContentLength: 0,
			Request:       req,
		}, nil
	}

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.Header.Set("X-Elastic-Product", "Elasticsearch")

	if req.Method == "GET" && req.URL.Path == "/" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
		resp.Body.Close()

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %v\nBody: %s", err, string(body))
		}

		modifiedData := map[string]interface{}{
			"name":         "opensearch-node",
			"cluster_name": "opensearch-cluster",
			"version": map[string]interface{}{
				"number":       "7.10.2", // Or match your OpenSearch version
				"build_type":   "docker",
				"build_hash":   "unknown",
				"build_date":   "2023-01-01T00:00:00.000Z",
				"distribution": "elasticsearch",
			},
			"tagline": "You Know, for Search",
		}

		modifiedBody, err := json.Marshal(modifiedData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal modified JSON: %v", err)
		}

		resp.Body = io.NopCloser(bytes.NewReader(modifiedBody))
		resp.ContentLength = int64(len(modifiedBody))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))
	}

	return resp, nil
}
