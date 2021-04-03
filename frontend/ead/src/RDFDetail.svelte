<script>
  import {rdfToHtml} from "./rdf";
  import response from './v2.json'
  import example from './example.json';
  import new_record from './new_base_record.json'
  import Timeline from "./Timeline.svelte";
  import Sources from "./Sources.svelte";
  import Viewer from "./Viewer.svelte";
  import Series from "./Series.svelte";

  export let route;

  const item = response.items[0]

  console.log(response.items)

  const viewerSection = {name: '', items: []}
  const dcSection = {name: 'Summary', items: []}
  const naveSection = {name: 'Collection', items: []}
  const edmSection = {name: 'Rights', items: []}
  const sections = [dcSection, edmSection]

  const detailConfig = {
    display: [

      {section: viewerSection, type: 'image', searchLabel: ['edm_hasView', 'nave_thumbLarge']},
      {section: dcSection, label: 'Vervaardiger', searchLabel: ['dc_creator']},
      {section: dcSection, label: 'Beschrijving', searchLabel: ['dc_description']},
      {section: dcSection, label: 'Onderwerp', searchLabel: ['dc_subject']},
      {section: dcSection, label: 'Soort object', searchLabel: ['dc_type']},
      {section: dcSection, label: 'Identificatie', searchLabel: ['dc_identifier']},
      {section: naveSection, label: 'Collectie', searchLabel: ['nave_collection']},
      {section: naveSection, label: 'Deelcollectie', searchLabel: ['nave_collectionPart']},
      {section: naveSection, label: 'Collectietype', searchLabel: ['nave_collectionType']},
      {section: edmSection, label: 'Instelling/bron', searchLabel: ['edm_dataProvider']},
      {section: edmSection, type: 'link', value: '@id', label: 'Originele context', searchLabel: ['edm_isShownAt']},
      {section: edmSection, type: 'link', value: '@id', label: 'Rechten', searchLabel: ['edm_rights']},
    ]
  };

  const display = [
    {
      value: 'dc:identifier'
    },
    {
      value: 'dc:title'
    }
  ]

  rdfToHtml([item], detailConfig)

  let image = viewerSection.items[0]
  console.log(example)
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

  <Timeline timeline={example['possesionReconstruction']}/>

  {#each sections as section}
    <section class="metadata">
      <header><h1>{section.name}</h1></header>
      <div class="info">
        {#each section.items as item}
          <p>
            <label>{item.label}</label>
            <span>{item.value}</span>
          </p>
        {/each}
      </div>
    </section>
  {/each}

  <Series></Series>
  <Sources/>
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