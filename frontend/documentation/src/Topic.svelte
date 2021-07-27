<script>
  import Markdown from "./Markdown.svelte";

  export let topics
  export let topicId

  let markdown;
  $: markdown = getMarkdown(topicId)

  function getMarkdown(topicId) {
    for (const topic of topics) {
      for (const link of topic.links) {
        if (link.id === topicId) {
          return link.markdown
        }
      }
    }
  }
</script>

<Markdown {markdown}/>