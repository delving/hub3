export function rdfToHtml(items, config) {

  function findDisplayConfig(searchLabels) {
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
    return {
      order: indexOf,
      ...config.display[indexOf]
    }
  }

  function defaultHelper(entry, searchLabels) {
    const display = findDisplayConfig(searchLabels);
    if (display) {
      const style = `style="order: ${display.order};"`;
      let label = '';
      if (display.label) {
        label = `<label>${display.label}</label>`
      }
      const id = `data-label="${entry.searchLabel || 'other'}"`;
      const value = entry[display.value || '@value'];

      if (display.type === 'image') {
        return {
          open: `<p ${style}>${label}<img ${id} src="${value}" />`,
          close: '</p>',
          display
        }
      } else if(display.type === 'link') {
        return {
          open: `<p ${style}>${label}<a ${id} href="${value}">${value}`,
          close: '</a></p>',
          display
        }
      }
      return {
        open: `<p ${style}>${label}<span ${id}>${value}`,
        close: '</span></p>',
        display
      }
    }
    return null;
  }

  function entries(html, entry, searchLabels) {
    searchLabels.push(entry.searchLabel);
    let tag = defaultHelper(entry, searchLabels)
    if (tag)
      tag.display.section ? tag.display.section.html.push(tag.open) : html.push(tag.open);
    if (entry.inline) {
      for (const inlineEntry of entry.inline.entries) {
        entries(html, inlineEntry, searchLabels);
      }
    }
    searchLabels.pop();
    if (tag)
      tag.display.section ? tag.display.section.html.push(tag.close) : html.push(tag.close);
  }

  for (const item of items) {
    const html = []
    entries(html, {inline: item.resources[0]}, [])
    item.html = html.join('');
  }
}