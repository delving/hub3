# Cloud Functions 

![Go](https://github.com/delving/hub3/workflows/Go/badge.svg)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/delving/hub3)
[![Go Report Card](https://goreportcard.com/badge/github.com/delving/hub3)](https://goreportcard.com/report/github.com/delving/hub3)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=delving_hub3&metric=alert_status)](https://sonarcloud.io/dashboard?id=delving_hub3)
[![codecov](https://codecov.io/gh/delving/hub3/branch/master/graph/badge.svg)](https://codecov.io/gh/delving/hub3)
[![GitHub release](https://img.shields.io/github/release/delving/hub3)](https://github.com/delving/hub3/releases/latest)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)



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












