# Text-fixtures for namespace package

Instead of managing a custom set of namespaces this package, uses the namespaces from [prefix.cc][prefix_cc]. It provides a [JSON-LD](https://www.w3.org/TR/json-ld/) version of all the namespaces. Compared to the previous custom list there are only two differences in prefixes:

* "http://www.w3.org/2003/01/geo/wgs84_pos#" used to be `wgs84_pos` but now defaults to `geo`
* "http://rdvocab.info/ElementsGr2/" used to be `rda` but now defaults to `rda2`


## Files

* `custom_context.jsonld` contains the legacy and custom namespaces
* `prefix_cc_context.jsonld` contains the defaults from [prefix.cc][prefix_cc]


[prefix_cc]: http://prefix.cc
