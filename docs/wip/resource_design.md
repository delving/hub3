# Immutable resource service

The goal of this service is to store triples and their resources as a series of immutable iterations. No triple is deleted. Inspiration comes from perkeep, restic and git. 

## Goals

* triple resolver
* resources resolver for indexing
* history for resources or datates
* subscribe to changes


## Ingestion workflow

* Syncer: sync external resources
* Extractor: extract source data 
* Transformer: transform into internal model resources
* Storer: store internal model with backend like postgresql or firestore
* Publisher: holds all functionality to sync with store interface and publish data. Examples are ElasticSearch for the search API and a triple store for the SPARQL endpoint

The publishers hold their own state and ask the storer for changes on dataset and then record/resource level.


## Immutable and history

In order to create a simpler API for history, the data we store must be immutable. We have series of indices to keep track of the claims and current versions of the data. Removed data is unlinked from the current version but not removed from the store

The content is stored in a content-addressable way

