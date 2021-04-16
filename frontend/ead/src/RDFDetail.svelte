<script>
  import Timeline from "./Timeline.svelte";
  import Sources from "./Sources.svelte";
  import Viewer from "./Viewer.svelte";
  import Series from "./Series.svelte";

  export let new_record;
  export let config;

  function parse(entries, output) {
    for (const entry of entries) {
      if(entry.entrytype === 'Literal') {
        let literals = output[entry.searchLabel]
        const value = entry['@value']
        if(!Array.isArray(literals) && value) {
          literals = []
          output[entry.searchLabel] = literals
        }
        if (value) {
          literals.push(value)
        }
      } else if(entry.entrytype === 'Bnode' || entry.entrytype === 'Resource') {
        let resource = output[entry.searchLabel]
        if(!Array.isArray(resource)) {
          resource = []
          output[entry.searchLabel] = resource
        }

        if(entry.inline && Array.isArray(entry.inline.entries)) {
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
  new_record = result
</script>

<div class="detail-page">
  <Viewer views={new_record['edm_hasView']}/>
  <section class="summary">
    {#each config.summary as property}
      {#if property.value in new_record}
        <p>{new_record[property.value]}</p>
      {/if}
    {/each}
  </section>

  <h1>Bezitsgeschiedenis</h1>
  <h2>Restitutie Status</h2>
  <p>{new_record.nk_restitutionState[0]}</p>

  <h2>Herkomst conclusie</h2>
  <ul>
    {#each new_record.nk_herkomstConclusion as conclusion}
      <li>
        {conclusion}
      </li>
    {/each}
  </ul>
  <Timeline timeline={new_record['nk_herkomst']}/>

  <Series config={config} series={new_record["nk_cho"]}></Series>
  <Sources sources={new_record['nk_hasSources']}/>
</div>

<style type="text/scss">
  //header {
  //  width: 25%;
  //}
  //
  //.info {
  //  width: 75%;
  //  display: flex;
  //  flex-direction: column;
  //}
  //
  //.summary {
  //  color: white;
  //  background-color: black;
  //  font-size: 140%;
  //  text-align: center;
  //  padding: 1em;
  //}
  //
  //.metadata {
  //  display: flex;
  //  flex-direction: row;
  //  border-bottom: 1px solid black;
  //
  //  p {
  //    display: flex;
  //    gap: 1em;
  //
  //    label {
  //      flex-basis: 15%;
  //    }
  //  }
  //}
  //
  //label {
  //  font-weight: bold;
  //}
</style>