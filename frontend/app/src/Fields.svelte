<script>
  export let objects;
  export let typeID;

  function collectFields(type, fields, names) {
    for (const field of type.fields) {
      names.push(field.name)
      if (field.type.typeID in objects) {
        const nestedType = objects[field.type.typeID]
        fields.push({
          name: names.join('.'),
          comment: field.comment,
          type: {
            typeID: nestedType.typeID,
            typeName: nestedType.name,
          }
        })
        collectFields(nestedType, fields, names)
      } else {
        fields.push({
          ...field,
          name: names.join('.')
        })
      }
      names.pop()
    }
  }

  function hasLink(field) {
    return field.type.typeID in objects
  }

  let fields = []
  collectFields(objects[typeID], fields, [])
</script>

<div>
  {#each fields as field}
    <div class="field">
      <strong class="bright-color" title=".{field.name}"><code>.{field.name}</code></strong>
      {#if hasLink(field)}
        <a href="#object:{field.type.typeID}">({field.type.typeName})</a>
      {:else}
        ({field.type.typeName})
      {/if}
      {field.comment}
    </div>
  {/each}
</div>

<style>
  .field {
    margin-top: .8em;
  }
</style>