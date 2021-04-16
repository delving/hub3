<script>
  import Metadata from "./Metadata.svelte";
  export let sources
  export let display

</script>

<section>
  <h1>Bronnen</h1>
    {#each sources as source}
      <Metadata context={source} {display}/>
<!--      <li>-->
<!--        <a href={source['nk_providerURL'][0]}>{source['rdf_label'][0]}</a>-->
<!--      </li>-->
    {/each}
</section>
