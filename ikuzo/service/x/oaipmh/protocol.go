package oaipmh

import "fmt"

// About is an optional and repeatable container to hold data about
// the metadata part of the record.
//
// The contents of an about container must conform to an XML Schema.
// Individual implementation communities may create XML Schema that define
// specific uses for the contents of about containers. Two common uses of
// about containers are:
// - rights statements: some repositories may find it desirable to attach
// terms of use to the metadata they make available through the OAI-PMH.
// No specific set of XML tags for rights expression is defined by OAI-PMH,
// but the about container is provided to allow for encapsulating
// community-defined rights tags.
// - provenance statements: One suggested use of the about container is
// to indicate the provenance of a metadata record, e.g. whether it has
// been harvested itself and if so from which repository, and when.
// An XML Schema for such a provenance container, as well as some
// supporting information is available from the accompanying
// Implementation Guidelines document.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#Record
type About struct {
	Body []byte `xml:",innerxml"`
}

// String returns the string representation
func (ab About) String() string {
	return string(ab.Body)
}

// Description is an extensible mechanism for communities to
// describe their repositories.
//
// For example, the description container could be used to include
// collection-level metadata in the response to the Identify request.
// Implementation Guidelines are available to give directions with
// this respect.
// Each description container must be accompanied by the URL of an
// XML schema describing the structure of the description container.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#Identify
//
type Description struct {
	Body []byte `xml:",innerxml"`
}

// String returns the string representation
func (ab Description) String() string {
	return string(ab.Body)
}

// Error represents an OAI error
//
// In event of an error or exception condition, repositories must
// indicate OAI-PMH errors, distinguished from HTTP Status-Codes,
// by including one or more error elements in the response.
// While one error element is sufficient to indicate the presence
// of the error or exception condition, repositories should report
// all errors or exceptions that arise from processing the request.
// Each error element must have a code attribute that must be from the
// following table; each error element may also have a free text string
// value to provide information about the error that is useful to a human
// reader. These strings are not defined by the OAI-PMH.
//
// Error Codes
// - badArgument:	The request includes illegal arguments, is missing
// required arguments, includes a repeated argument, or values for arguments
// have an illegal syntax.
// - badResumptionToken:	The value of the resumptionToken argument is
// invalid or expired.
// - badVerb:	Value of the verb argument is not a legal OAI-PMH verb, the
// verb argument is missing, or the verb argument is repeated.
// - cannotDisseminateFormat: The metadata format identified by the value
// given for the metadataPrefix argument is not supported by the item or by
// the repository.
// - idDoesNotExist:	The value of the identifier argument is unknown or illegal
//  in this repository.
// - noRecordsMatch:	The combination of the values of the from, until, set and
//  metadataPrefix arguments results in an empty list.
// - noMetadataFormats:	There are no metadata formats available for the
//  specified item.
// - noSetHierarchy:	The repository does not support sets.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#ErrorConditions
//
type Error struct {
	Code            string `xml:"code,attr"`
	Message         string `xml:",chardata"`
	applicableVerbs []Verb
}

func (err Error) Error() string {
	return fmt.Sprintf("%s: %s", err.Code, err.Message)
}

func (err Error) Is(target error) bool {
	t, ok := target.(Error)
	if !ok {
		return false
	}

	return t.Code == err.Code
}

// MetadataFormat is a metadata format available from a repository
//
// It contains:
// 1. The metadataPrefix - a string to specify the metadata format in OAI-PMH
// requests issued to the repository. metadataPrefix consists of any valid URI
//  unreserved characters. metadataPrefix arguments are used in ListRecords,
//  ListIdentifiers, and GetRecord requests to retrieve records, or the headers
//   of records that include metadata in the format specified by the
//    metadataPrefix;
// 2. The metadata schema URL - the URL of an XML schema to test validity of
// metadata expressed according to the format;
// 3. The XML namespace URI that is a global identifier of the metadata format.
//  (http://www.w3.org/TR/1999/REC-xml-names-19990114/Overview.html)
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#MetadataNamespaces
//
type MetadataFormat struct {
	MetadataPrefix    string `xml:"metadataPrefix"`
	Schema            string `xml:"schema,omitempty"`
	MetadataNamespace string `xml:"metadataNamespace"`
}

// Header contains the unique identifier of the item and properties necessary
// for selective harvesting.
//
// The header consists of the following parts:
// - the unique identifier -- the unique identifier of an item in a repository;
// - the datestamp -- the date of creation, modification or deletion of the
// record for the purpose of selective harvesting.
// - zero or more setSpec elements -- the set membership of the item for the
// purpose of selective harvesting.
// - an optional status attribute with a value of deleted indicates the withdrawal
// of availability of the specified metadata format for the item, dependent on
// the repository support for deletions.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#Record
type Header struct {
	Identifier string   `xml:"identifier"`
	DateStamp  string   `xml:"datestamp"`
	SetSpec    []string `xml:"setSpec"`
	Status     string   `xml:"status,attr,omitempty"`
}

// Identify is a verb used to retrieve information about a repository.
//
// Some of the information returned is required as part of the OAI-PMH.
// Repositories may also employ the Identify verb to return additional
// descriptive information.
// The response must include one instance of the following elements:
// - repositoryName : a human readable name for the repository;
// - baseURL : the base URL of the repository;
// - protocolVersion : the version of the OAI-PMH supported by the repository;
// - earliestDatestamp : a UTCdatetime that is the guaranteed lower limit
// of all datestamps recording changes, modifications, or deletions in the
// repository. A repository must not use datestamps lower than the one
// specified by the content of the earliestDatestamp element. earliestDatestamp
// must be expressed at the finest granularity supported by the repository.
// - deletedRecord : the manner in which the repository supports the notion
// of deleted records. Legitimate values are no ; transient ; persistent
// with meanings defined in the section on deletion.
// - granularity: the finest harvesting granularity supported by the repository.
// The legitimate values are YYYY-MM-DD and YYYY-MM-DDThh:mm:ssZ with meanings
// as defined in ISO8601.
//
// The response must include one or more instances of the following element:
// - adminEmail : the e-mail address of an administrator of the repository.
//
// The response may include multiple instances of the following optional
// elements:
//
// - compression : a compression encoding supported by the repository.
// The recommended values are those defined for the Content-Encoding header
// in Section 14.11 of RFC 2616 describing HTTP 1.1. A compression element
// should not be included for the identity encoding, which is implied.
// - description : an extensible mechanism for communities to describe their
// repositories. For example, the description container could be used to
// include collection-level metadata in the response to the Identify request.
// Implementation Guidelines are available to give directions with this respect.
// Each description container must be accompanied by the URL of an XML schema
// describing the structure of the description container.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#Identify
type Identify struct {
	// must
	RepositoryName    string   `xml:"repositoryName" json:"repositoryName"`
	BaseURL           string   `xml:"baseURL" json:"baseURL"`
	ProtocolVersion   string   `xml:"protocolVersion" json:"protocolVersion"`
	EarliestDatestamp string   `xml:"earliestDatestamp" json:"earliestDatestamp"`
	DeletedRecord     string   `xml:"deletedRecord" json:"deletedRecord"`
	Granularity       string   `xml:"granularity" json:"granularity"`
	AdminEmail        []string `xml:"adminEmail" json:"adminEmail"`

	// may
	Description []Description `xml:"description" json:"description"`
}

// Metadata is a single manifestation of the metadata from an item.
//
// The OAI-PMH supports items with multiple manifestations (formats)
// of metadata.
// At a minimum, repositories must be able to return records with metadata
// expressed in the Dublin Core format, without any qualification. Optionally,
// a repository may also disseminate other formats of metadata.
// The specific metadata format of the record to be disseminated is specified
// by means of an argument
// -- the metadataPrefix
// -- in the GetRecord or ListRecords request that produces the record.
// The ListMetadataFormats request returns the list of all metadata formats
// available from a repository, or for a specific item (which can be specified
// as an argument to the ListMetadataFormats request).
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#Record
type Metadata struct {
	Body []byte `xml:",innerxml"`
}

// GoString returns the body as a string
func (md Metadata) String() string {
	return string(md.Body)
}

// Record is metadata expressed in a single format.
//
// A record is returned in an XML-encoded byte stream in response to an OAI-PMH
// request for metadata from an item.
// A record is identified unambiguously by the combination of the unique
// identifier of the item from which the record is available, the metadataPrefix
// identifying the metadata format of the record, and the datestamp of
// the record.
// The XML-encoding of records is organized into the following parts:
// header, metadata, about
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#Record
type Record struct {
	Header   Header   `xml:"header"`
	Metadata Metadata `xml:"metadata"`
	About    About    `xml:"about"`
}

// ResumptionToken is a token that manages the flow control during harvesting.
//
// The following optional attributes may be included as part of the
// resumptionToken element along with the resumptionToken itself:
//
// - expirationDate -- a UTCdatetime indicating when the resumptionToken ceases
// to be valid.
// - completeListSize -- an integer indicating the cardinality of the complete
// list (i.e., the sum of the cardinalities of the incomplete lists). Because
// there may be changes in a repository during a list request sequence, as
// described under Idempotency of resumptionTokens, the value of completeListSize
// may be only an estimate of the actual cardinality of the complete list and
// may be revised during the list request sequence.
// - cursor -- a count of the number of elements of the complete list thus far
// returned (i.e. cursor starts at 0).

// https://www.openarchives.org/OAI/openarchivesprotocol.html#FlowControl
type ResumptionToken struct {
	Token            string `xml:",chardata" json:"token"`
	CompleteListSize int    `xml:"completeListSize,attr" json:"completeListSize"`
	ExperationDate   string `xml:"experationDate,attr,omitempty" json:"experationDate"`
	Cursor           int    `xml:"cursor,attr"`
}

// Set is an optional construct for grouping items for the purpose of
// selective harvesting.
//
// Repositories may organize items into sets. Set organization may be
// flat, i.e. a simple list, or hierarchical.
// Multiple hierarchies with distinct, independent top-level nodes are allowed.
// Hierarchical organization of sets is expressed in the syntax of the setSpec
// parameter as described below.
// When a repository defines a set organization it must include set membership
// information in the headers of items returned in response to the
// ListIdentifiers, ListRecords and GetRecord requests.
//
// A Set has:
// - setSpec -- a colon [:] separated list indicating the path from the root
// of the set hierarchy to the respective node. Each element in the list is a
// string consisting of any valid URI unreserved characters, which must not
// contain any colons [:]. Since a setSpec forms a unique identifier for the
// set within the repository, it must be unique for each set. Flat set
// organizations have only sets with setSpec that do not contain any colons [:].
// - setName -- a short human-readable string naming the set.
// - setDescription -- an optional and repeatable container that may hold
// community-specific XML-encoded data about the set; the accompanying
// Implementation Guidelines document provides suggestions regarding the
// usage of this container.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#Set
//
type Set struct {
	SetSpec        string      `xml:"setSpec"`
	SetName        string      `xml:"setName,omitempty"`
	SetDescription Description `xml:"setDescription,omitempty"`
}
