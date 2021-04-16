<script>
  import {Viewer} from 'openseadragon'
  import {onDestroy, onMount} from "svelte";

  export let views;

  let viewer;

  onMount(() => {
    console.log(views)
    viewer = new Viewer({
      id: "viewer",
      prefixUrl: '/images/openseadragon/',
      tileSources: views.map((v, i) => ({
        tileSource: v['nave_deepZoomUri'][0],
        width: 200,
        x: 210 * i,
        y: 0
      }))
    });
  })

  onDestroy(() => viewer.destroy())
</script>

<div>
  <div id="viewer"></div>
</div>

<style>
  div {
    background-color: black;
  }

  #viewer {
    width: 100%;
    height: 500px;
  }
</style>