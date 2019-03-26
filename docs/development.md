# Developing Hub3

Hub3 is a [golang] application. In this document we describe the development choices.

Currently, Hub3 depends on a few external dependencies.

- A search engine: We currently support the [Apache Lucene]-based search-engine [ElasticSearch] 5.
- A triple store: We currently support [Apache Fuseki] out of the box. But any Triple store with a [SPARQL 1.1] and [SPARQL Update 1.1] endpoint should work.
- DeepZoom generation library: [LibVips] is low-memory, high-performance image processing library.

Current development and production uses the following versions:

- ElasticSearch 6.4 (configured port: 9200)
- Apache Fuseki 3.5.0 (configured port: 3030)
- LibVips 8.5.9

All of these dependencies are available on the major platforms via package-managers or direct download: Linux, MacOS and Windows. Please follow the respective installation instructions to get them up and running. Hub3 will complain at startup when these are not available.

[golang]: https://golang.org/
[Apache Lucene]: https://lucene.apache.org/ 
[ElasticSearch]: https://www.elastic.co/guide/en/elasticsearch/reference/5.6/getting-started.html
[SPARQL 1.1]: https://www.w3.org/TR/sparql11-query/
[SPARQL Update 1.1]: https://www.w3.org/TR/sparql11-update/
[LibVips]: http://jcupitt.github.io/libvips/
