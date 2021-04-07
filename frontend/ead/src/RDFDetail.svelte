<script>
  import new_record from './NK3189.json'
  import Timeline from "./Timeline.svelte";
  import Sources from "./Sources.svelte";
  import Viewer from "./Viewer.svelte";
  import Series from "./Series.svelte";
  import Metadata from "./Metadata.svelte";

  export let route;

  const viewerSection = {name: '', items: []}
  const idSection = {name: 'Identificatie', items: []}
  const naveSection = {name: 'Collection', items: []}
  const edmSection = {name: 'Rights', items: []}
  const sections = [
    {
      label: 'Identificatie',
      items: [
        {label: 'Titel', path: ['dc:title']},
        {label: 'NK nummer', path: ['objectNumber']},
        {label: 'Beschrijving', path: ['dc:description']}
      ],
    },
    {
      label: 'Vervaardiging',
      items: [
        {label: 'Vervaardiger', path: ['creator', 'creatorName']},
        {label: 'Datum start', path: ['productionDate', 'dateStart']},
        {label: 'Datum einde', path: ['productionDate', 'dateEnd']}
      ],
    },
    {
      label: 'Fysieke kenmerken',
      items: [
        {label: 'Materiaal', path: ['nave:material']},
        {label: 'Hoogte/lengte', path: ['dimension', 'heightLenght']},
        {label: 'Diameter', path: ['dimension', 'depthDiameter']}
      ]
    },
    {
      label: 'Onderwerp',
      items: [
        {label: 'Wat', path: ['dc:subject']},
        {label: 'Object categorie', path: ['nave:objectCategory']},
        {label: 'Object naam', path: ['nave:objectName']}
      ]
    }
  ]

  const display = [
    {
      value: 'dc:identifier'
    },
    {
      value: 'dc:title'
    }
  ]

  let image = viewerSection.items[0]
</script>

<div class="detail-page">
  <Viewer views={new_record['edm:hasView']}/>
  <section class="summary">
    {#each display as property}
      {#if property.value in new_record}
        <p>{new_record[property.value]}</p>
      {/if}
    {/each}
  </section>

  <Timeline timeline={new_record['possesionReconstruction']}/>
  <h2>Restitutie Status</h2>
  <p>{new_record.restititutionStatus.status}</p>

  {#each sections as section}
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
  header {
    width: 25%;
  }

  .info {
    width: 75%;
    display: flex;
    flex-direction: column;
  }

  .summary {
    color: white;
    background-color: black;
    font-size: 140%;
    text-align: center;
    padding: 1em;
  }

  .metadata {
    display: flex;
    flex-direction: row;
    border-bottom: 1px solid black;

    p {
      display: flex;
      gap: 1em;

      label {
        flex-basis: 15%;
      }
    }
  }

  label {
    font-weight: bold;
  }
</style>