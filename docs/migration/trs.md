# Generic metadata-data object storage (Time Revision Store)

What:

    - store everything as object with changes stored in an eventstream
        - deltas are store on the object level
        - eventstream compatible with for example LDES (Linked Data Event Stream (Gent))
    - single API to:
        - store, introspect and retrieve objects
        - retrieve version of objects
        - multiple identifiers possible for a single object
    - workers run from same binary but on different nodes and get task via Redis
    - workers can subscribe to relevant event-stream for synchronization, for example:
        - elasticsearch
        - triple-store (fuseki)
        - OAI-PMH layer on event-stream
        - Sitemap layer based event-stream filters
        - update of static dumps such as ntriples per dataset
    - object-storage is immutable (so easy to backup and replay)
    - store:
        - source records
        - transformed RDF records
        - configuration
        - prepared downloads

- stage 1:
    - infra:
        - add
            - redis
            - postgresql
            - s3 compatible object storage (file-based)
        - remove
            - nats
            - diskstorage hub3
    - code zvt-hub3:
        - move background go-routines to work tasks interface
        - move custom storage to object storage
        - update service interface
        - update changes to new dataset services (minor)
- stage 2:
    - infra:
        - add
            - mapping-engine grpc service to worker-nodes
        - remove:
            - narthex (datasets managed part of hub3 binary)
    - code:
        - remove:
            - last stateful components from zvt-hub
        - add:
            - IPA authorization to hub3 configuration

# When

- stage 1 implementation 22-08 - 02-09
- stage 2 implementation late september
