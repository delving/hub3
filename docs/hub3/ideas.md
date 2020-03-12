# Development Ideas

## EAD description search

- Goals: 
    - implement fast, accurate and robust search in EAD description
    - search for words with prefix wildcards
    - do the heavy lifting at index time

- steps:
    - integrate description counter at index time
    - store index counter as json
    - map object is count + list of where it is found (deduplicated list). Append at the end of each DataItem (maybe separate)
    - retrieve count via separate API (always reload summary):
        - get keys with hits
        - create Replacer with keys
        - create deduplicated or map with DataItem order
        - sort DataItem order Keys
        - retrieve all DataItems
        - append matched to new list (fixed size)
        - run replacer on each item for highlights
        - return to the APIs


- next:
    - save counter to file
    - read counter from file
    - filter and update output in order (maybe return map )
    - return API

