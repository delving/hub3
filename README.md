# Delving RAPID

RAPID is a recursive acronym that stands for *Rapid API Delving*. 

The goal of RAPID is to provide *an API Framework that makes it easy and predictable for webdevelopers to work with arbitrarily structured RDF and leverage semantic network technology*.

The core functionality that it aims to provide can be summarised by the acronym *SILAS*:

* **S**PARQL proxy
* **I**ndex RDF
* **L**inked Open Data Resolver
* **A**ggregate and transform RDF
* **S**earch RDF

Part of the design is to require as little external dependencies outside the compiled *Golang* binary as possible. 

## Install

RAPID is written in Golang, so you have to setup your Golang environment first, see [Golang Installation].

After that you can glone it from bitbucket:

    $ git clone git@bitbucket.org:delving/rapid.git $GOPATH/src/bitbucket.org/delving

Or use `go get`

    $ go get bitbucket.org/delving/rapid

In order to get the dependencies, install [glide]:

    $ curl https://glide.sh/get | sh
    $ glide install

Start the server with the default configuration.

    $ rapid http

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
