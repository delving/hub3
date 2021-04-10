<script>
  import Timeline from "./Timeline.svelte";
  import Sources from "./Sources.svelte";
  import Viewer from "./Viewer.svelte";
  import Series from "./Series.svelte";
  import Metadata from "./Metadata.svelte";

  export let new_record;
  export let config;
</script>

<div class="detail-page">
  <Viewer views={new_record['edm:hasView']}/>
  <section class="summary">
    {#each config.summary as property}
      {#if property.value in new_record}
        <p>{new_record[property.value]}</p>
      {/if}
    {/each}
  </section>

  <Timeline timeline={new_record['possesionReconstruction']}/>
  <h2>Restitutie Status</h2>
  <p>{new_record.restititutionStatus.status}</p>

  {#each config.sections as section}
    <section class="metadata">
      <header><h1>{section.label}</h1></header>
      <div class="info">
        {#each section.items as item}
          <p>
            <label>{item.label}</label>
            <Metadata context={new_record['cho']} path={item.path}/>
          </p>
        {/each}
      </div>
    </section>
  {/each}

  <Series series={new_record["dcterms:hasParts"]}></Series>
  <Sources sources={new_record['hasSources']}/>
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