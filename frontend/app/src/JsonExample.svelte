<script>
  import highlight from 'highlight.js'
  import {onMount} from 'svelte'

  export let objects;
  export let typeID;

  let code;

  function toJson(type, result) {
    const fields = type.fields;
    for (const field of fields) {
      if (field.type.typeID in objects) {
        const child = {}
        result[field.name] = child
        toJson(objects[field.type.typeID], child)
      } else {
        result[field.name] = field.example
      }
    }
  }

  const builder = {}
  toJson(objects[typeID], builder)
  let json = JSON.stringify(builder, null, 2)

  onMount(() => highlight.highlightBlock(code))
</script>

<pre><code bind:this={code} class="language-json">{json}</code></pre>