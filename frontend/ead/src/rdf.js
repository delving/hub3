export function rdfToHtml(items, config) {

  function findDisplayConfig(entry, searchLabels) {
    const indexOf = config.display.findIndex(item => {
      if (item.searchLabel.length > searchLabels.length) {
        return false;
      }
      let prevIndexOfLabel = -1;
      for (const label of item.searchLabel) {
        const indexOfLabel = searchLabels.indexOf(label);
        if (indexOfLabel === -1) return false;
        if (indexOfLabel <= prevIndexOfLabel) return false;
        prevIndexOfLabel = indexOfLabel;
      }
      return true;
    });
    if (indexOf === -1) return null;

    const display = config.display[indexOf]
    display.section.items.push({
      order: indexOf,
      label: display.label,
      type: display.type,
      value: entry[display.value || '@value'],
      path: display.searchLabel[display.searchLabel.length - 1]
    });
  }

  function entries(html, entry, searchLabels) {
    searchLabels.push(entry.searchLabel);
    findDisplayConfig(entry, searchLabels)
    if (entry.inline) {
      for (const inlineEntry of entry.inline.entries) {
        entries(html, inlineEntry, searchLabels);
      }
    }
    searchLabels.pop();
  }

  for (const item of items) {
    const html = []
    entries(html, {inline: item.resources[0]}, [])
    item.html = html.join('');
  }
}