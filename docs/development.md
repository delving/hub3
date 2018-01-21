# Developing RAPID

RAPID is a [golang] application. In this document we describe the development choices.

## Dependency management

We have decided to use [glide] for vendoring. All pinned dependencies are stored in the `./vendor` directory.

Here are some basic commands to work with glide.

Install the dependencies and revisions listed in the lock file into the vendor directory. If no lock file exists an update is run.

    $ glide install

Install the latest dependencies into the vendor directory matching the version resolution information. The complete dependency tree is installed, importing Glide, Godep, GB, and GPM configuration along the way. A lock file is created from the final output.

    $ glide update

Add a new dependency to the glide.yaml, install the dependency, and re-resolve the dependency tree. Optionally, put a version after an anchor.

    $ glide get github.com/foo/bar

or 

    $ glide get github.com/foo/bar#^1.2.3

## System Dependencies

Currently, RAPID depends on a few external dependencies.

- A search engine: We currently support the [Apache Lucene]-based search-engine [ElasticSearch] 5.
- A triple store: We currently support [Apache Fuseki] out of the box. But any Triple store with a [SPARQL 1.1] and [SPARQL Update 1.1] endpoint should work.
- DeepZoom generation library: [LibVips] is low-memory, high-performance image processing library.

Current development and production uses the following versions:

- ElasticSearch 5.6 (configured port: 9200)
- Apache Fuseki 3.5.0 (configured port: 3030)
- LibVips 8.5.9

All of these dependencies are available on the major platforms via package-managers or direct download: Linux, MacOS and Windows. Please follow the respective installation instructions to get them up and running. RAPID will complain at startup when these are not available.



[golang]: https://golang.org/
[glide]: https://glide.sh 
[Apache Lucene]: https://lucene.apache.org/ 
[ElasticSearch]: https://www.elastic.co/guide/en/elasticsearch/reference/5.6/getting-started.html
[SPARQL 1.1]: https://www.w3.org/TR/sparql11-query/
[SPARQL Update 1.1]: https://www.w3.org/TR/sparql11-update/
[LibVips]: http://jcupitt.github.io/libvips/
