# Changelog

## unreleased 

- history of changes: see https://github.com/delving/hub3/compare/v0.3.0...main

### Added
- support for default genreform values while processing EADs [[GH-173]](https://github.com/delving/hub3/pull/173)
- Add support for tag based filtering via Bulk API [[GH-176]](https://github.com/delving/hub3/pull/176)

### Changed

- default behaviour OAI-PMH is to allow harvesting List verb without a set spec parameter [[GH-172]](https://github.com/delving/hub3/pull/172)
- lod resolver returing 404 and trying fallback to original URI  [[GH-174]](https://github.com/delving/hub3/pull/174)
- Retrieve info from existing mets file if process digital is disabled [[GH-175]](https://github.com/delving/hub3/pull/175)
- LOD resolver configuration [[GH-177]](https://github.com/delving/hub3/pull/177)

### Fixed

- fixed user-input in logging warnings from github code scanning [[GH-170]](https://github.com/delving/hub3/pull/170)

### Deprecated

### Removed

## v0.3.0 (2022-08-31)

- history of changes: see https://github.com/delving/hu3/compare/v0.2.1...v0.3.0

### Added 
 
- [10f521cf](https://github.com/delving/hub3/commit/10f521cf) hub3: added dedicated esproxy service. 
- [80c4d3be](https://github.com/delving/hub3/commit/80c4d3be) hub3: added sitemap package 
- [41309cb0](https://github.com/delving/hub3/commit/41309cb0) hub3: initial version of postgresql service. 
- [14c8be956](https://github.com/delving/hub3/commit/14c8be956) hub3: added basic support for extracting text from alto files.
- [7b33f66fc](https://github.com/delving/hub3/commit/7b33f66fc) hub3: added 'render' helper package. 
- [98ac874a5](https://github.com/delving/hub3/commit/98ac874a5) hub3: added allowed ports filter to imageproxy
- [3b5c13128](https://github.com/delving/hub3/commit/3b5c13128) hub3: changed resource package name to rdf. 
- [22374e307](https://github.com/delving/hub3/commit/22374e307) hub3: added support for resource based indexing to rdf.Graph
- [b56c798b7](https://github.com/delving/hub3/commit/b56c798b7) hub3: added support for sentry.io error aggregation.
- [497639871](https://github.com/delving/hub3/commit/497639871) hub3: bulk sparql updates support filtered by urn:private.
- [6855f2c3e](https://github.com/delving/hub3/commit/6855f2c3e) hub3: added more default namespaces to namespace package.
- [e47a6604c](https://github.com/delving/hub3/commit/e47a6604c) hub3: initial implementation of sparql service
- [1088b1265](https://github.com/delving/hub3/commit/1088b1265) hub3: initial support for file package.
- [6596f5554](https://github.com/delving/hub3/commit/6596f5554) hub3: added initial support for harvest.Syncer interface.
- [44971706a](https://github.com/delving/hub3/commit/44971706a) hub3: initial version of generic lod resolver package.
- [58bd58a16](https://github.com/delving/hub3/commit/58bd58a16) hub3: Initial support for mappingxml  RDF serialization format.
- [abf71ee8a](https://github.com/delving/hub3/commit/abf71ee8a) hub3: added support for rdf.Graph generation from FragmentResources.
- Add /def/ to allowed LOD redirect resources [[GH-155]](https://github.com/delving/hub3/pull/155)
- OAI-PMH service and elasticsearch driver  [[GH-160]](https://github.com/delving/hub3/pull/160)
- Add config option to ignore certain paths from request logging. [[GH-161]](https://github.com/delving/hub3/pull/161)
- `GINGER_LOG` runtime variable to log request send to the Ginger posthook endpoint  [[GH-167]](https://github.com/delving/hub3/pull/167)
- default image support for imageproxy service  [[GH-164]](https://github.com/delving/hub3/pull/164)

### Changed

- [2cf6dc75](https://github.com/delving/hub3/commit/2cf6dc75) hub3: update ikuzoctl configuration objects for domain.Service. 
- [e59c880b](https://github.com/delving/hub3/commit/e59c880b) hub3: refactor for multitenancy based on OrgID across all packages 
- [3a10b33f](https://github.com/delving/hub3/commit/3a10b33f) hub3: added multi-tenant configuration support to domain.Organization 
- [e373e132](https://github.com/delving/hub3/commit/e373e132) hub3: move memory storage package to x 
- [9949b8a7](https://github.com/delving/hub3/commit/9949b8a7) hub3: refactor elasticsearch storage into driver package 
- [82f2e6cf](https://github.com/delving/hub3/commit/82f2e6cf) hub3: LDF always returns 'text/turtle' 
- updated github configuration  [[GH-133]](https://github.com/delving/hub3/pull/133)
- [b6e9e4cd1](https://github.com/delving/hub3/commit/b6e9e4cd1) hub3: use ikuzo/render package in legacy handlers
- [af997e3cd](https://github.com/delving/hub3/commit/af997e3cd) hub3: update organization_config configuration
- [9c4e90946](https://github.com/delving/hub3/commit/9c4e90946) hub3: changed oai-pmh service to support stores.
- [0140f7eee](https://github.com/delving/hub3/commit/0140f7eee) hub3: changed sitemap package to support stores.
- [303eb85a1](https://github.com/delving/hub3/commit/303eb85a1) hub3: migrate es search statistics to own sub-package.
- `/version` endpoint to include orgID and better wrapper for data. [[GH-156]](https://github.com/delving/hub3/pull/156)
- Enable default orgID as fallback [[GH-157]](https://github.com/delving/hub3/pull/157)
- Always set modified when indexing fragments.FragmentGraph [[GH-158]](https://github.com/delving/hub3/pull/158)
- ignore 404 logging fix + updated protobuf definition for scans [[GH-163]](https://github.com/delving/hub3/pull/163)
- Enable harvest all datasets option in OAI-PMH service  [[GH-165]](https://github.com/delving/hub3/pull/165)
- how wikibase SPARQL-endpoints are harvested [[GH-168]](https://github.com/delving/hub3/pull/168)
- orphan-control for mets-records  [[GH-164]](https://github.com/delving/hub3/pull/164)

## Removed
 
- [1df50338](https://github.com/delving/hub3/commit/1df50338) hub3: remove gorm package 

## Fixed
 
- [57a0d8ae9](https://github.com/delving/hub3/commit/57a0d8ae9) hub3: fixed order for processing triples in ead processing.
- [4a3160b27](https://github.com/delving/hub3/commit/4a3160b27) hub3: fixes for unit tests and ikuzo server after refactor.
- [228d44b32](https://github.com/delving/hub3/commit/228d44b32) hub3: fixed predictable ordering for namespace defaults.
- pagination with LDF hypermedia controls  [[GH-154]](https://github.com/delving/hub3/pull/154)
- Fixed adding namespaces of types during graph processing [[GH-159]](https://github.com/delving/hub3/pull/159)
- Typo in `rdf:type` for mappingxml serializer [[GH-165]](https://github.com/delving/hub3/pull/166)

## Deprecated

- [e50ac4aea](https://github.com/delving/hub3/commit/e50ac4aea) hub3: deprecate index.Client in favor of elasticsearch driver Client.

### Security

- fixed dependabot alerts and fixed gitea upgrade errors [[GH-169]](https://github.com/delving/hub3/pull/169)

## v0.2.1 (2022-01-25)

- history of changes: see https://github.com/delving/hu3/compare/v0.2.0...v0.2.1

### Added

- Use cfg.PeriodDesc instead of tree.Periods in daoCfg and FindingAid, fixes GH-118 [[GH-121]](https://github.com/delving/hub3/pull/121)
- ClI subcommand 'bulk' to index bulk-requests that are serialized to disk [[GH-88]](https://github.com/delving/hub3/pull/88)
- Config option to store records generated from METS files in dedicated index [[GH-83]](https://github.com/delving/hub3/pull/83)
- Extended scan.proto with metadata map for free-form content [[GH-108]](https://github.com/delving/hub3/pull/108)
- Resource package to ikuzo for uniform RDF handling [[GH-106]](https://github.com/delving/hub3/pull/106)
- Imageproxy: lrucache, deepzoom and thumbnail transformation support [[GH-115]](https://github.com/delving/hub3/pull/115)
- Added support for indexing map datatypes in scans indexed from METS files [[GH-117]](https://github.com/delving/hub3/pull/117)
- PeriodDesc in daoCfg and FindingAid [[GH-118]](https://github.com/delving/hub3/pull/118)
- Ignore private resources in bulk sparql export [[GH-123]](https://github.com/delving/hub3/pull/123)
- processDigitalIfMissing processes mets if missing [[GH-126]](https://github.com/delving/hub3/pull/126)
- Cache cleaning work for imageproxy service [[GH-119]](https://github.com/delving/hub3/pull/119)
- Support for allowed mimetypes to imageproxy service [[GH-130]](https://github.com/delving/hub3/pull/130)
- Changedetection for reads after the graph has changed [[GH-131]](https://github.com/delving/hub3/pull/131)


## Changed 
- Allow for changes to uploaded EAD file before storing it [[GH-90]](https://github.com/delving/hub3/pull/90)
- Refactored EAD search overview queries into separate filter and collapse queries [[GH-91]](https://github.com/delving/hub3/pull/91)
- Added support for NDE Dataset Register API [[GH-92]](https://github.com/delving/hub3/pull/92)
- Added posthooks support to orphan groups as well  [[GH-99]](https://github.com/delving/hub3/pull/99)
- Added Run() custom function to posthook interface [[GH-107]](https://github.com/delving/hub3/pull/107)
- Return "text/turtle" from LDF endpoint  [345290fca38e](https://github.com/delving/hub3/commit/345290fca38ef9573671da34e11e4fc5f2c20729)
- Added minor changes for EAD support requested by Dutch National Archive [[GH-150]](https://github.com/delving/hub3/pull/150)

### Fixed
 
- Explicit initialisation of config for CLI subcommands instead of global on init [[GH-88]](https://github.com/delving/hub3/pull/88)
- Prevent duplicate files while processing METS files [[GH89]](https://github.com/delving/hub3/pull/89)
- Tree paging was using the old API key [[GH-93]](https://github.com/delving/hub3/pull/93)
- Delete non-EAD datasets would return incorrect error  [[GH-10]](https://github.com/delving/hub3/pull/102)
- Sanitation in EAD unittitle was too strict [[GH-109]](https://github.com/delving/hub3/pull/109)
- Show all EAD unittitles in tree.Label [[GH-110]](https://github.com/delving/hub3/pull/110)
- Prevent redirect loop with invalid 'inventoryID' in tree API [[GH-112]](https://github.com/delving/hub3/pull/112)
- Prevent 404 when call is made to /ead/tree without "q" parameter [[GH-114]](https://github.com/delving/hub3/pull/114)
- Drop orphans with lowercase 'findingaid' tag [[GH-124]](https://github.com/delving/hub3/pull/124)
- Return 404 from imageproxy when remote resource is not found [[GH-127]](https://github.com/delving/hub3/pull/127)
- Bulk index service sparql-update private filter produces invalid RDF [[GH-125]](https://github.com/delving/hub3/pull/125)


### Removed

- Unused embedded assets functionality in favour of `go:embed`  [[GH-99]](https://github.com/delving/hub3/pull/99)

### Security
- Resolved security issues in dependencies [[GH-95]](https://github.com/delving/hub3/pull/95)


## v0.2.0 (2021-03-10)

- history of changes: see https://github.com/delving/hub3/compare/v0.1.10...v0.2.0


### Added 

- Synchronize with remote archives and METS-files via OAI-PMH  [[GH-80]](https://github.com/delving/hub3/pull/80)
- Dedicated dataset service for managing datasets  [[GH-79]](https://github.com/delving/hub3/pull/79)
- Initial service for elasticsearch functionality. Better mapping management.  [[GH-78]](https://github.com/delving/hub3/pull/78)
- Multitenancy support through organisationIDs  [[GH-77]](https://github.com/delving/hub3/pull/77)
- Initial version of Time Revision Store  [[GH-76]](https://github.com/delving/hub3/pull/76)
- Dedicated SPARQL service that can be used as a separate publisher [[GH-75]](https://github.com/delving/hub3/pull/75)
- Support for dropping orphan-groups within a dataset [[GH-74]](https://github.com/delving/hub3/pull/74)
- Support for blacklisting URIs and limiting caching to referrers in the image proxy service [[GH73]](https://github.com/delving/hub3/pull/73)
- Create c level and Node from p element in dsc element and add every Cp field to the Odd in the created Clevel. [[GH-67]](https://github.com/delving/hub3/pull/67)
- Sort fields in the fieldMap resource.NewFields because the map is always unordered. Store protobuf messages in resource and use pointers instead of values for protobuf field in FragmentGraph. [[GH-63]](https://github.com/delving/hub3/pull/63)
- Update to v2 RAML API documentation [[GH-59]](https://github.com/delving/hub3/pull/59)
- Extract extref from p when found and add it to the odd. [[GH-58]](https://github.com/delving/hub3/pull/58)
- Create c level and Node from p element in dsc element. [[GH-56]](https://github.com/delving/hub3/pull/56)
- Posthook callbacks to `ead.Service` and `index.Service` [[GH-55]](https://github.com/delving/hub3/pull/55)
- Support for [test-containers](https://golang.testcontainers.org/) for ikuzo service and storage tests [[GH-27]](https://github.com/delving/hub3/pull/27)
- GitHub Action configurations and quality control [[GH-29]](https://github.com/delving/hub3/pull/29)
- GitHub Action for enforcing Contributor License Agreement [[GH-45]](https://github.com/delving/hub3/pull/45)
- Add logging when EAD dataset is deleted [[GH-46]](https://github.com/delving/hub3/pull/46)
- Config option `maxTreeSize` for setting the maximum size of the navigation tree API [[GH-48]](https://github.com/delving/hub3/pull/48)
- Config option `processDigital` to enable processing of digital object links in EAD upload [[GH-49]](https://github.com/delving/hub3/pull/49)
- Config option `datasetTags` to add custom tags to `meta.tags` based on the datasetID [[GH-52]](https://github.com/delving/hub3/pull/52)

### Fixed
- Fix for graph traversal when creating v1 indexing records [[GH-64]](https://github.com/delving/hub3/pull/64)
- use "<br/>" instead of "<lb/>" in the description API output [[GH-54]](https://github.com/delving/hub3/pull/54)
- Code quality improvements reported by SonarCloud and golangci-lint [[GH-35]](https://github.com/delving/hub3/pull/35)
- Concurrent retrieval of RDF for webresources now uses errgroup [[GH-44]](https://github.com/delving/hub3/pull/44)
- Orphan Control is now excecuted when all domainpb.IndexMessage have been processed [[GH-47]](https://github.com/delving/hub3/pull/47)
- Search highlighting now adds style-class to all hits when surrounded by custom tags [[GH-50]](https://github.com/delving/hub3/pull/47)
- Inline '<lb/>' in description API ead.DataItem [[GH-53]](https://github.com/delving/hub3/pull/53)

### Removed

- Remove RAML api-console [[GH-33]](https://github.com/delving/hub3/pull/33)

### Security

- CVE-2020-14040 (High) detected in github.com/microsoft/hcsshim-fc27c5026e6ff001dc1b171b99bda7bb3dcf6e78 [[GH-34]](https://github.com/delving/hub3/pull/34)

## v0.1.11 (2020-07-21)

- history of changes: see https://github.com/delving/hub3/compare/v0.1.10...v0.1.11

### Added

- Allow for custom []ProxyRoute to be configured when configuring a DataNode proxy [[GH-26]](https://github.com/delving/hub3/pull/26)

### Fixes

- Delete DataSet does not work when when no v1 index is created [[GH-25]](https://github.com/delving/hub3/pull/25)

## v0.1.10 (2020-07-20)

- history of changes: see https://github.com/delving/hub3/compare/v0.1.9...v0.1.10

### Added

- Support for running hub3 in DataNode mode [[GH-24]](https://github.com/delving/hub3/pull/24)
- Support for generic bulk service orgID-aware posthooks
- Implementation of orgID-aware posthook for [Ginger platform](https://github.com/driebit/ginger)

### Fixes

- escape html entities in v1 indexing output

## v0.1.9 (2020-06-15) 

- history of changes: see https://github.com/delving/hub3/compare/v0.1.8...v0.1.9

### Added 

- Tree: make sure fields in Tree API remain sorted by their sortorder key.
- EAD: removed deprecated pre-task EAD processing functions
- EAD: add support for meta.tags in EAD upload form
- License: add license headers to all Go and Protobuf source files

### Fixes 
- Search: add backtick '`' to characters to be ignored or removed during tokenization
- Memory: full-text-index phrase search now docID aware
- EAD: when EAD processing tasks stop with an error the worker tasks should not exit

## v0.1.8 (2020-06-05) 

- EAD: allow treePage calls with search hits
- Revision: add support for Git RPC calls
- Search: caching proxy for ElasticSearch
- Search: add support for index and alias management
- Logging: custom logger based on rs/zerolog
- Index: add index service with support for NATS based queued indexing and official ElasticSearch library
- Server: add TLS support via the configuration
- Server: dependency injection via new config package based on viper configuration
- EAD: task-based processing of uploaded EADs
- Protobuf: update to new golang Protobuf API, see https://blog.golang.org/protobuf-apiv2
- Bulk: Bulk service migrate to Index Service
- EAD: structured logging for all task states

- history of changes: see https://github.com/delving/hub3/compare/v0.1.7...v0.1.8

## v0.1.7 (2020-05-05) 

- bugfix: add c level xml tag for nested clevels when unmarshalling EADs
- bugfix: revert back to ElasticSearch QueryStringQuery because unexpected handling of boolean queries by SimpleQueryStringQuery.
- ead: allow for deeplinking into EAD tree view
- ead: support next and previous scrollIDs in the ScrollPager

- history of changes: see https://github.com/delving/hub3/compare/v0.1.6...v0.1.7

## v0.1.6 (2020-04-22) 

- Protobuf definition for primary domain model for metadata
- Tokenized highlighting is now tag aware.
- EAD: when paging query returns zero redirect to first page.
- EAD: Add eadid, filedesc, 'archdesc>did' to the Description DataItems.
- EAD: Added generator for EAD numbered clevel support.
- EAD: Enable support for numbered clevels again.
- EAD: CLI option to update EAD description indices.
- boyscout: small fixes.

- history of changes: see https://github.com/delving/hub3/compare/v0.1.5...v0.1.6

## v0.1.5 (2020-04-01)

- Set ResponseSize to 1 to properly calculate the scrollID and next cursor position when searching through the tree.
- history of changes: see https://github.com/delving/hub3/compare/v0.1.4...v0.1.5

## v0.1.4 (2020-03-30)

- Remove EAD disk-store on introspect reset
- Use more robust elastic.SimpleQueryStringQuery for full-text searches
- Used config.EAD.SearchFields for building EAD full-text queries
- QueryParser bugfix for wrongly assigned phrase queries.
- search.QueryTerm to elastic.Query converter
- history of changes: see https://github.com/delving/hub3/compare/v0.1.3...v0.1.4


## v0.1.3 (2020-03-29)

- history of changes: see https://github.com/delving/hub3/compare/v0.1.2...v0.1.3

- Remove EAD cache directory on dataset delete.
- New search.Tokenizer that integrates with search.Vector and is tag-aware.
- search.TokenStream supports vector-based highlighting
- TextIndex and TextQuery use search.Vectors and search.Matches
- fragments.Resources NewFields() support vector-based highlighting
- DescriptionIndex and Description now use 'gob' serialisation format for faster retrieval.
- DescriptionIndex is persisted to disk on upload
- all search options that interact with ead.Description use persisted DescriptionIndex
- moved toplevel 'pkg' content to 'hub3' package
- New fast search.Autocomplete that takes search.TokenStream as input.
- New search.SpellCheck for spelling suggestions for historic texts
- Automatic extraction to triples of EAD "<persname>", "<date>", and "<geogname>". These can be used for search or aggregations.
- ElasticSearch mapping update. 'html_strip' char-filter is part of the default analyzer.

## v0.1.2 (2020-03-16)

- minor fixes
    - update to elasticsearch mapping fields
    - fix for access error on EAD genreform.
- history of changes: see https://github.com/delving/hub3/compare/v0.1.1...v0.1.2

## v0.1.1 (2020-03-16)

- history of changes: see https://github.com/delving/hub3/compare/v0.1.0...v0.1.1

### hub3
- refactor hub3 package structure inside the hub3 subpackage.
- hub3 runnable is now inside 'hub3ctl'
- removal of top-level 'middleware' and 'logger' packages in favor of ikuzo packages
- updates to 'ead' package, see EAD section

### ikuzo
- integration of ikuzo proof-of-concept new package organisation in 'ikuzo' sub-package
- integration of RAML-based API console. This is all the scaffolding. The API documentation is still in progress.
- implementation of custom Lucene-like query-parser in 'search.QueryParser'
- implementation of memory-based single document Full-Text indexing in 'memory.TextIndex'
- implementation of memory-based incremental highlighter 'memory.TextQuery'. Unicode search and highlighting is fully supported.
- extended search.Analyser to also transform phrases
- added wrappers to 'logger' package to support integration with stdlib logging implementations such as 'olivere/elasticsearch'.
- initial experimental version of search.QueryTerm builder for ElasticSearch queries. This will replace the elastic.SimpleQueryStringQuery in future versions.

### EAD

- new EAD parser based on the full public set of EADs of the Dutch National Archive. Future parsers can be generated from source EADs.
- implementation of RDF triple generator for EAD clevel and did fields. For search these fields are stored in ElasticSearch in 'tree.rawContent'.
- disabled the support for numbered clevels in the EAD. This functionality can be activated when the generator is completed.
- added support for custom fields in the EAD tree API. Which fields to show can be configured through ead.TreeFields in the TOML configuration.
- tree API handler now supports 'withFields=true' parameters to enable the display of the tree-fields.
- replaced highlighting of EAD description API with memory.TextQuery. 

## v0.1.0 (2020-03-11)

- initial Go Modules complaint version
