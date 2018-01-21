// Copyright © 2017 Delving B.V. <info@delving.eu>
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

package index

import "time"

// RdfObject holds all the fields for the RDF triples that will be indexed by ElasticSearch
type RdfObject struct {
	// URI of the subject
	Subject string `json:"subject"`

	// RDF types of the subject
	SubjectClass []string `json:"class"`

	// URI of the predicate
	Predicate string `json:"predicate"`

	// Label of predicate. Used for user facing searching
	SearchLabel string `json:"searchLabel"`

	// URI or Literal value of the objject the object
	Object string `json:"object"`

	// The RDF language of the object
	ObjectLang string `json:"language,omitempty"`

	// The XSD:type of the Object Literal value
	ObjectContentType string `json:"objectContentType,omitempty"`

	// Boolean to determine if the RdfObject is a Literal or URI reference
	IsResource bool `json:"isResource"`

	// RDF label of the resource. This can be a URI, a label or an inlined label from the URI.
	Value string `json:"value"`

	// A field that can be used for searching
	LatLong string `json:"geoHash,omitempty"`

	// A field that contains geo polygons
	Polygon []float64 `json:"polygon,omitempty"`

	// Raw non-analysed field of value that can be used for facetting and aggregations. Will not contain a URI.
	Facet string `json:"facet,omitempty"`

	// The level of the triple compared to the root subject. This is used for relevance ranking in the query.
	Level int `json:"level"`

	// if on level 2 or 3 list the URI of the referring subject.
	ReferrerSubject string `json:"refererSubject,omitempty"`

	// the predicate of the referring label
	ReferrerPredicate string `json:"referrerPredicate,omitempty"`

	// the searchLabel of the predicate. This can be used for searching
	ReferrerSearchLabel string `json:"referrerSearchLabel,omitempty"`

	// the NamedGraph that this object belongs to
	NamedGraph string `json:"namedGraph,omitempty"`

	// the order in which the resource (for the subject) is sorted when inlined from the referer
	ResourceSortOrder int `json:"sortOrder,omitempty"`
}

// RDFSearchRecord holds all the fields for the result of a SPARQL query for a subject.
// The SPARQL query contains data in three levels. The each triple gets assigned a level for additional weighting at search time.
type RDFSearchRecord struct {
	HubID string `json:"hubId"`

	SourceURI string `json:"sourceUri"`

	DataSetName string `json:"spec"`

	DataSetURI string `json:"datasetUri"`

	Tags []string `json:"tags"`

	RecordType string `json:"recordType"`

	Revision int `json:"revision"`

	Modified time.Time `json:"modified"`

	Created time.Time `json:"created"`

	Triples []RdfObject `json:"triples"`
}
