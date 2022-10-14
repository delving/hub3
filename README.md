# Hub3  (v2)

![Go](https://github.com/delving/hub3/workflows/Go/badge.svg)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/delving/hub3)
[![Go Report Card](https://goreportcard.com/badge/github.com/delving/hub3)](https://goreportcard.com/report/github.com/delving/hub3)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=delving_hub3&metric=alert_status)](https://sonarcloud.io/dashboard?id=delving_hub3)
[![codecov](https://codecov.io/gh/delving/hub3/branch/master/graph/badge.svg)](https://codecov.io/gh/delving/hub3)
[![GitHub release](https://img.shields.io/github/release/delving/hub3)](https://github.com/delving/hub3/releases/latest)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Hub3 is an [Linked Open Data](https://en.wikipedia.org/wiki/Linked_data) Aggregration and Publication platform.

## History

Since 2010 Delving is providing an [Open Source](https://en.wikipedia.org/wiki/Open_source) Aggregation and Publication platform. The original version was based on the Open Source version of the first [Europeana](https://www.europeana.eu/en) production version. In the following years different iterations of the CultureHub were developed and used in production. In 2014, the internal core was refactored to use [RDF](https://en.wikipedia.org/wiki/Resource_Description_Framework) and [Semantic Web](https://en.wikipedia.org/wiki/Semantic_Web) technologies. 

- CultureHub 0: java-based platform enherited from Europeana
- CultureHub 1: migration to Play1 scala based multi-tentant system
- CultureHub 2: Play2 based refactor
- CultureHub 3 (v1): 
  - [Narthex](https://github.com/delving/narthex): Play2/scala based Dataset management and aggregation platform
  - [Nave](https://github.com/delving/nave): Django/python based front-end and API platform
  - [Sip-Creator](https://github.com/delving/sip-creator/): Java swing based desktop mapping application
- Hub3 (v2): refactor of hub3 v1 to a [Golang](https://go.dev) based platform. (This is the current repository)

The Roadmap of the v2 version is to migrate all diverse services and frameworks used into a single and more maintainable code-base.

## Changelog

[Changelog](./CHANGELOG.md)

## License

Copyright (c) 2017-present Delving B.V.

Licensed under [Apache 2.0](./License)

[Golang Installation]: https://golang.org/doc/install








