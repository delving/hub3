<script>
  import Timeline from "./Timeline.svelte";
  import Sources from "./Sources.svelte";
  import Viewer from "./Viewer.svelte";
  import Series from "./Series.svelte";
  import Metadata from "./Metadata.svelte";

  export let new_record;
  export let config;

  function parse(entries, output) {
    for (const entry of entries) {
      if (entry.entrytype === 'Literal') {
        let literals = output[entry.searchLabel]
        const value = entry['@value']
        if (!Array.isArray(literals) && value) {
          literals = []
          output[entry.searchLabel] = literals
        }
        if (value) {
          literals.push(value)
        }
      } else if (entry.entrytype === 'Bnode' || entry.entrytype === 'Resource') {
        let resource = output[entry.searchLabel]
        if (!Array.isArray(resource)) {
          resource = []
          output[entry.searchLabel] = resource
        }

        if (entry.inline && Array.isArray(entry.inline.entries)) {
          const item = {}
          resource.push(item)
          parse(entry.inline.entries, item)
        }
      } else {
        console.warn(`Unsupported entry type: ${entry.entrytype}`)
      }
    }
  }

  const result = {}
  console.log(new_record)
  parse(new_record.resources[0].entries, result)
  console.log(result)
  console.log(config)
  new_record = result
</script>

<div class="detail-page">
  <Viewer views={new_record['edm_hasView']}/>

  <Metadata context={new_record} display={config.ownershipHistory}/>

  <Timeline timeline={new_record['nk_herkomst']}/>

  <Series config={config} series={new_record["nk_cho"]}></Series>
  <Sources display={config.source} sources={new_record['nk_hasSources']}/>
</div>
