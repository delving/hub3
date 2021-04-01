<script>
  import {rdfToHtml} from "./rdf";
  import * as response from './v2.json'
  import Timeline from "./Timeline.svelte";

  export let route;

  const item = response.items.find(i => i.id === +route.values.id);

  console.log(response.items)

  const viewerSection = {name: '', html: []}
  const dcSection = {name: 'Summary', html: []}
  const naveSection = {name: 'Collection', html: []}
  const edmSection = {name: 'Rights', html: []}
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

  rdfToHtml([item], detailConfig)
</script>

<div class="detail-page">
  {@html viewerSection.html.join('')}

  <Timeline/>

  {#each sections as section}
    <section>
      <header><h1>{section.name}</h1></header>
      <div class="info">
      {@html section.html.join('')}
      </div>
    </section>
  {/each}
</div>

<style>
  .detail-page {
    width: 50%;
    margin: auto;
  }

  header {
    width: 25%;
  }

  .info {
    width: 75%;
    display: flex;
    flex-direction: column;
  }
</style>