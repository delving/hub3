<script>
  import highlight from 'highlight.js'

  export let type;

  let code;
  $: if (code) {
    const builder = {}
    toJson(type, builder)
    code.textContent = JSON.stringify(builder, null, 2)
    highlight.highlightBlock(code)
  }

  function toJson(type, result) {
    const fields = type.fields;
    for (const field of fields) {
      if (field.typeDef) {
        const child = {}
        result[field.name] = child
        toJson(field.typeDef, child)
      } else {
        result[field.name] = field.example
      }
    }
  }
</script>

<pre><code bind:this={code} class="language-json">No content</code></pre>