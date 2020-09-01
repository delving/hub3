# Changelog

## unreleased 

- history of changes: see https://github.com/delving/hub3/compare/v0.1.11...master

### Added 

- Support for [test-containers](https://golang.testcontainers.org/) for ikuzo service and storage tests [[GH-27]](https://github.com/delving/hub3/pull/27)
- GitHub Action configurations and quality control [[GH-29]](https://github.com/delving/hub3/pull/29)
- Add logging when EAD dataset is deleted [[GH-46]](https://github.com/delving/hub3/pull/46)
- Config option `maxTreeSize`for setting the maximum size of the navigation tree API [[GH-48]](https://github.com/delving/hub3/pull/48)
- Config opion `processDigital` to enable processing of digital object links in EAD upload [[GH-49]](https://github.com/delving/hub3/pull/49)

### Fixed

- Code quality improvements reported by SonarCloud and golangci-lint [[GH-35]](https://github.com/delving/hub3/pull/35)
- Concurrent retrieval of RDF for webresources now uses errgroup [[GH-44]](https://github.com/delving/hub3/pull/44)
- Orphan Control is now excecuted when all domainpb.IndexMessage have been processed [[GH-47]](https://github.com/delving/hub3/pull/47)
- Search highlighting now adds style-class to all hits when surrounded by custom tags [[GH-50]](https://github.com/delving/hub3/pull/47)

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

- history of changes: see https://github.com/delving/hub3/compare/v0.1.6...master

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
