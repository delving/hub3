# [Ginger](Ginger) synchronization posthook

This posthook was originally developed for [Erfgoed Brabant's](https://erfgoedbrabant.nl) [Brabant Cloud](https://www.brabantcloud.nl) service to synchronize data with the [ginger](https://github.com/driebit/ginger) CMS that is behind the [Brabants Erfgoed](https://www.brabantserfgoed.nl) discovery website. The posthook runs when data is saved to the *Brabant Cloud* index and is an asynchronous process. The core functionality of this posthook is:


* Each valid record that is saved in the *Brabant Cloud* is send as valid RDF graph to a *Ginger* endpoint in batches of 250. The format is *line-delimited JSON*. So *Ginger* should process each line as a new record. For datasets configured as *excluded* no records are submitted to the *Ginger endpoint*.
* When a dataset is disabled/deleted from the *Brabant Cloud* index a `DELETE` request is send to the *Ginger* endpoint with the *dataset spec* and a *revision*. When the *revision* is 0, all records should be removed.
* Orphan control (removal of records that should no longer be in the index) uses the same `DELETE` request but with a specific *ervision*. All records not matching this *revision* should be removed from *Ginger*.
* When a request fails the error and the submitted *JSON* is logged. The request is not retried. 


## Configuration

The posthook is configured in the main TOML configuration.

```toml
[[posthooks]]
# the name of the posthook
name = "ginger"
# specs to exclude from posthook
excludeSpec = [
    "mip",
]
# target URLs for JSON-LD post to Ginger
url = ''
# API-key that is appended to the url as `api_key=[key]`
apikey = ''
```

*Excluded specs/datasets* are not send to the *Ginger* endpoint. Without the `API key` a *forbidden* response is returned. 

*Ginger* exposes a single API endpoint that performs the POST and DELETE requests.

## Date validation

Because the semantic indexing behind *Ginger* automatically detects dates and invalid dates cause indexing errors, the posthook parses and validates the dates in the following fields.

```golang
var dateFields = map[string]bool{
	ns.dcterms.Get("created").RawValue():       true,
	ns.dcterms.Get("issued").RawValue():        true,
	ns.nave.Get("creatorBirthYear").RawValue(): true,
	ns.nave.Get("creatorDeathYear").RawValue(): true,
	ns.nave.Get("date").RawValue():             true,
	ns.dc.Get("date").RawValue():               true,
	ns.nave.Get("dateOfBurial").RawValue():     true,
	ns.nave.Get("dateOfDeath").RawValue():      true,
	ns.nave.Get("productionEnd").RawValue():    true,
	ns.nave.Get("productionStart").RawValue():  true,
	ns.nave.Get("productionPeriod").RawValue(): true,
	ns.rdagr2.Get("dateOfBirth").RawValue():    true,
	ns.rdagr2.Get("dateOfDeath").RawValue():    true,
}
```

The date-field is always included with a `Raw` suffix. The above fields are only included when the date is valid. 


## PostHook `POST`
The posthook gathers postook jobs until it reaches 250. Then it will generate and postprocess the JSON-LD. Postprocessing steps are:

* adding custom JSON-LD fields for *Ginger* in the *Narthex* namespaces (http://schemas.delving.eu/narthex/terms/)
    * `localID`:the identifier of the record
    * `hubID`: the identifier used internally by the *BrabantCloud*
    * `spec`: the dataset identifier
    * `revision`: the revision of the dataset
    * `belongsTo`: the RDF subject of the record
* validating the JSON-LD
* valdating the configured date-fields.
* creating the line-delimited JSON

Finally, it is submitted to the *Ginger* endpoint with a `POST` and *Content-Type* `application/json-ld; charset=utf-8`. 

When status-code is not `200`, an error is logged with the payload and reasons returned from the *Ginger* endpoint. The API contract assumes that if the PostHook sends over valid JSON-LD that *Ginger* should be able to process them without errors. The errors returned are therefore *internal server errors* that should be monitored and resolved by ginger. The error logging in the PostHook is mainly for debugging and signalling purposes. 

## PostHook `DELETE`

To remove a datasets or orphaned records from the *Ginger endpoint* a `DELETE` request is send with the following parameters:

* `collection`: the dataset identifier (or `spec`) that is supplied in the custom JSON-LD fields `spec`. 
* `revision`: the revision of valid records. All records with other *revisions* should be removed from the *Ginger* index. When `0` or no revision is given all records should be removed.

## Data model

Currently, the JSON-LD send is mostly a [EDM](https://pro.europeana.eu/page/edm-documentation) extension. But in the future other valid RDF models/records could be send over. The posthook treats any valid RDF graph as a record that can be submitted to the *Ginger endpoint*. Currently, the API contract only guarantees that the custom JSON-LD fields are always present. All other triples should be processed generically by the *Ginger endpoint*. 
