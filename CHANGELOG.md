# Changelog

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
