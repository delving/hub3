package oaipmh

// ListMetadataFormats holds the formats received from server
//
// OAI-PMH supports the dissemination of records in multiple metadata
// formats from a repository.
// The ListMetadataFormats request returns the list of all metadata
// formats available from a repository
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#MetadataNamespaces
//
type ListMetadataFormats struct {
	MetadataFormat []MetadataFormat `xml:"metadataFormat"`
}

// ListIdentifiers is an abbreviated verb form of ListRecords, retrieving only
// headers rather than records.
//
// Optional arguments permit selective harvesting of headers based on set
// membership and/or datestamp.
//
// Depending on the repository's support for deletions, a returned header
// may have a status attribute of "deleted" if a record matching the arguments
//
// specified in the request has been deleted.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#ListIdentifiers
//
type ListIdentifiers struct {
	Headers         []Header        `xml:"header"`
	ResumptionToken ResumptionToken `xml:"resumptionToken"`
}

// GetRecord is used to retrieve an individual metadata record from a repository.
//
// Required arguments specify the identifier of the item from which the record
// is requested and the format of the metadata that should be included in the
// record. Depending on the level at which a repository tracks deletions, a
// header with a "deleted" value for the status attribute may be returned,
// in case the metadata format specified by the metadataPrefix is no longer
// available from the repository or from the specified item.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#GetRecord
//
type GetRecord struct {
	Record Record `xml:"record"`
}

// ListRecords is a verb used to harvest records from a repository.
//
// Optional arguments permit selective harvesting of records based on set
// membership and/or datestamp. Depending on the repository's support for
// deletions, a returned header may have a status attribute of "deleted"
// if a record matching the arguments specified in the request has been deleted.
// No metadata will be present for records with deleted status.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#FlowControl
type ListRecords struct {
	Records         []Record        `xml:"record"`
	ResumptionToken ResumptionToken `xml:"resumptionToken"`
}

// ListSets represents a list of Sets
//
//  http://www.openarchives.org/OAI/openarchivesprotocol.html#Set
//
type ListSets struct {
	Set []Set `xml:"set"`
}

// Response encapsulates the information from a harvest request
//
// All responses to OAI-PMH requests must be well-formed XML instance documents.
// Encoding of the XML must use the UTF-8 representation of Unicode. Character
// references, rather than entity references, must be used. Character references
// allow XML responses to be treated as stand-alone documents that can be
// manipulated without dependency on entity declarations external to the document.
//
// The XML data for all responses to OAI-PMH requests must validate against the
// XML Schema shown at the end of this section . As can be seen from that schema,
// responses to OAI-PMH requests have the following common markup:
//
// (ignored) The first tag output is an XML declaration where the version is
// always 1.0 and the encoding is always UTF-8, eg: <?xml version="1.0" encoding="UTF-8" ?>
// (ignored) The remaining content is enclosed in a root element with the name OAI-PMH.
//
// For all responses, the first two children of the root element are:
// - responseDate -- a UTCdatetime indicating the time and date that the
// response was sent. This must be expressed in UTC
// - request -- indicating the protocol request that generated this response.
//
// The third child of the root element is either:
// - error -- an element that must be used in case of an error or exception condition;
// - (Identify, ListMetadataFormats, ListSets, GetRecord, ListIdentifiers,
// ListRecords) an element with the same name as the verb of the respective
// OAI-PMH request.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#XMLResponse
//
type Response struct {
	ResponseDate string      `xml:"responseDate"`
	Request      RequestNode `xml:"request"`
	Error        Error       `xml:"error"`

	Identify            Identify            `xml:"Identify"`
	ListMetadataFormats ListMetadataFormats `xml:"ListMetadataFormats"`
	ListSets            ListSets            `xml:"ListSets"`
	GetRecord           GetRecord           `xml:"GetRecord"`
	ListIdentifiers     ListIdentifiers     `xml:"ListIdentifiers"`
	ListRecords         ListRecords         `xml:"ListRecords"`
}

// RequestNode is indicating the protocol request that generated this response.
//
// The rules for generating the request element are as follows:
// 1. The content of the request element must always be the base URL of the protocol request;
// 2. The only valid attributes for the request element are the keys of the key=value pairs of protocol request. The attribute values must be the corresponding values of those key=value pairs;
// 3. In cases where the request that generated this response did not result in an error or exception condition, the attributes and attribute values of the request element must match the key=value pairs of the protocol request;
// 4. In cases where the request that generated this response resulted in a badVerb or badArgument error condition, the repository must return the base URL of the protocol request only. Attributes must not be provided in these cases.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#XMLResponse
//
type RequestNode struct {
	Verb           string `xml:"verb,attr"`
	Set            string `xml:"set,attr"`
	MetadataPrefix string `xml:"metadataPrefix,attr"`
}

// HasResumptionToken determines if the request has or not a ResumptionToken
func (resp *Response) HasResumptionToken() bool {
	return resp.ListIdentifiers.ResumptionToken.Token != "" || resp.ListRecords.ResumptionToken.Token != ""
}

// ResumptionToken determine the resumption token in this Response
func (resp *Response) GetResumptionToken() (hasResumptionToken bool, resumptionToken string, completeListSize int) {
	if resp == nil {
		return
	}

	// First attempt to obtain a resumption token from a ListIdentifiers response
	resumptionToken = resp.ListIdentifiers.ResumptionToken.Token

	if resumptionToken != "" {
		completeListSize = resp.ListIdentifiers.ResumptionToken.CompleteListSize
	}

	// Then attempt to obtain a resumption token from a ListRecords response
	if resumptionToken == "" {
		resumptionToken = resp.ListRecords.ResumptionToken.Token
		completeListSize = resp.ListRecords.ResumptionToken.CompleteListSize
	}

	// If a non-empty resumption token turned up it can safely inferred that...
	if resumptionToken != "" {
		hasResumptionToken = true
	}

	return
}
