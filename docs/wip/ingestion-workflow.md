# Ikuzo ingestion workflow

What is the wizard flow for this.

Create organization
Create dataset (dataset predetermined ingestion pipe-lines: mapping, ead so dataset is marked as such)
    - datasetID (contains org + dataset ID) => together must be unique
- determine source.Synchronizer
    - drop down from list
    - if remote register with source.Scheduler (part of the source service)
- receive data (upload or from Synchronizer/Harvester/Crawler)
    - store in source (implicit opening of work tree in time revision store)
- analyze data => get tree with unique content 
    - support graphs as well
- create extractor (unique ID + record separator)
- extract source records
    - store in source_records (trs prefix)
- run transformer to resourceGraph
    - external sources that are retrieved are stored in sub-dir of source_record (think mets files); external lod data is stored in the cache graph
- store resourceGraph in TRS under 'resources'
    - when all are stored we can
        - either commit all records
        - get a git status on what has changed
    - post-commit we create a diff of which files have been changed (new, updated, deleted)
        - send them to the registered publishers for the dataset
    - rollback to previous state
- publish
    - Publish domain.resourceGraph
    - DropResources()
    - DropDataset()
    - RollBack()
    - Synchronize()
    (publisher register mount points for API) => API discovery and automatic rendering for RAML API
- Generators are registered via dataset to source.Service
    - generators convert domain.ResourceGraph to other formats ( for example to EDM, EDM-strict, RDF formats, Linked Data etc )


# known Resource / TRS prefix paths

- source: the raw source as uploaded or remotely synchronized. This is a single source that records need to be extracted from
- source-records: the individual records in raw format stored by their local identifier as retrieved from extractor.
- mapping: mapping files
- resource: the resource graphs stored in the Internal Resource Model json format
- es-record: the resolved resource graph that it stored in elasticsearch


## extractor

takes io.Reader and splits the source into source.Record (localId string, body io.Reader)

## organization

- list of publishers 
    - can be overridden at the dataset level
- RDF base-url
    - needed for transformer

## dataset

- orgID
- datasetID
- source.Syncer
- source.Extractor ( idPath, recordPath )
- resource.Transformer (trs Path)
- []resource.Publishers (how do they interact, register by mime-type or functionality)
- statistics from 
    - publishers
    - source.Stats
    - syncer.Stats

    
## Free form


DatasetList returns status updates on datasets 


