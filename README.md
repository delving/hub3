# Hub3 

Hub3 is an RDF publication and discovery platform written in Golang. Before the 1.0 release packages that can be of individual use will be split into stand-alone packages.

The goal of Hub3 is to provide *an API Framework that makes it easy and predictable for webdevelopers to work with arbitrarily structured RDF and leverage semantic network technology*.

The core functionality that it aims to provide can be summarised by the acronym *SILAS*:

* **S**PARQL proxy
* **I**ndex RDF
* **L**inked Open Data Resolver
* **A**ggregate and transform RDF
* **S**earch RDF

Part of the design is to require as little external dependencies outside the compiled *Golang* binary as possible. 

**NOTE:** this is currently a work in progress and APIs can change between releases.

## Install

Hub3 is written in Golang, so you have to setup your Golang environment first, see [Golang Installation].

After that you can glone it from github:

    $ git clone git@github.com:delving/hub3.git $GOPATH/src/github.com/delving

Or use `go get`

    $ go get github.com/delving/hub3

Start the server with the default configuration.

    $ hub3 http

For development setup, see [Develop](./docs/development.md).

For deployment instructions, see [Deployment](./docs/deployment.md).

## Changelog

### Master

### 0.1

* First fully-functional public version

## License

Copyright (c) 2017-present Delving B.V.

Licensed under [Apache 2.0](./License)

[Golang Installation]: https://golang.org/doc/install
[glide]: https://glide.sh 












