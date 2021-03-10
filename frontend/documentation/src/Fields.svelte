<script>
  export let type;

  let fields;
  $: {
    let newFields = []
    collectFields(type, newFields, [])
    fields = newFields
  }

  function collectFields(type, fields, names) {
    for (const field of type.fields) {
      names.push(field.name)
      if (field.type.typeDef) {
        const nestedType = field.type.typeDef
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
</script>

<div>
  {#each fields as field}
    <div class="field">
      <strong class="bright-color" title=".{field.name}"><code>.{field.name}</code></strong>
      ({field.type.typeName})
      {field.comment}
    </div>
  {/each}
</div>

<style>
  .field {
    margin-top: .8em;
  }
</style>