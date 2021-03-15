const parser = new DOMParser();

function toSpan(node, copyAttributes, depth, builder) {
  const tagName = node.tagName.toLowerCase();
  const classNames = [tagName]
  const attributes = []
  if (tagName === 'c') {
    classNames.push(`c${depth}`);
    attributes.push(`data-identifier=${builder.cLevelCount}`)
  }
  if (copyAttributes) {
    for (let i = 0; i < node.attributes.length; i++) {
      const attr = node.attributes[i];
      attributes.push(`${attr.name}=${attr.value}`);
    }
  }
  if (isText(node)) {
    classNames.push('text');
  }
  return {
    toHtml: function (extraClasses) {
      return `<span ${attributes.join(' ')} class="${classNames.concat(extraClasses).join(' ')}">`
    }
  };
}

function isText(node) {
  const children = node.childNodes;
  return children && children.length === 1 && children[0].nodeType === Node.TEXT_NODE;
}

const COPY_ATTRIBUTES_OF = {
  'LIST': true,
  'EXTPTR': true,
  'C': true
};

function nodeToHtml(node, builder, depth = 0) {

  if (node.nodeType !== Node.TEXT_NODE) {
    const isOpen = builder.openTag(node, depth);
    for (const child of [...node.childNodes]) {
      nodeToHtml(child, builder, depth + 1);
    }
    if(isOpen) builder.closeTag();
  } else {
    builder.addText(node);
  }
}


class Builder {
  html = [];
  open = [];
  pages = [];
  closedTagCount = 0;
  cLevelCount = 0;

  constructor(limit, accept) {
    this.limit = limit;
    this.accept = accept;
  }

  openTag(node, depth) {
    const label = node.getAttribute('label')
    if (label) {
      this.html.push(`<span class="inline-label">${label}</span>`);
    }

    const spanBuilder = toSpan(node, node.tagName.toUpperCase() in COPY_ATTRIBUTES_OF, depth, this);
    if (node.tagName.toLowerCase() === 'c') this.cLevelCount++;
    if (this.accept(node)) {
      this.html.push(spanBuilder.toHtml([]));
      this.open.push(spanBuilder);
      return true;
    }
    return false;
  }

  addText(textNode) {
    if (!this.accept(textNode)) return;
    const text = textNode.textContent.trim();
    if (text) {
      if (isText(textNode.parentNode)) {
        this.html.push(text);
      } else {
        this.html.push(`<span class="text">${text}</span>`);
      }
    }
  }

  closeTag() {
    this.html.push('</span>');
    this.closedTagCount++;
    this.open.pop();
    if (this.closedTagCount === this.limit) {
      const currentHtml = this.html;
      this.html = [];
      for (let i = 0; i < this.open.length; i++) {
        currentHtml.push('</span>');
        this.html.push(this.open[i].toHtml(['continued']));
      }
      this.pages.push(currentHtml.join(''));
      this.closedTagCount = 0;
    }
  }

  build() {
    if (this.html.length > 0) {
      this.pages.push(this.html.join(''));
      this.html = [];
    }
    return this.pages;
  }
}

function toHtml(node, limit = Number.MAX_SAFE_INTEGER, skipFilter = () => true) {
  const builder = new Builder(limit, skipFilter);
  nodeToHtml(node, builder)
  return builder.build()
}

function navTreeFilter(node) {
  if (!node.tagName) {
    return navTreeFilter(node.parentNode);
  }
  const tagName = node.tagName.toLowerCase();
  if (tagName === 'c') {
    const level = node.getAttribute('level')
    return level === 'series' || level === 'subseries';
  } else if (tagName === 'unittitle') {
    let target = node;
    while(target && target.tagName.toLowerCase() !== 'c') {
      target = target.parentNode;
    }
    return navTreeFilter(target);
  }
  return false;
}

export function parseEad(eadXml) {
  const xmlDoc = parser.parseFromString(eadXml, "text/xml");
  let sections = xmlDoc.querySelectorAll('archdesc > did, archdesc > descgrp');
  sections = [...sections].map(section => ({
    title: section.firstChild.textContent.trim() || section.childNodes[1].textContent.trim(),
    html: toHtml(section)[0]
  }))
  const tree = xmlDoc.querySelector('archdesc > dsc[type="combined"]');
  console.log(tree);
  const treePages = toHtml(tree, 250);
  console.log(treePages.length)
  const navigationTree = toHtml(tree, Number.MAX_SAFE_INTEGER, navTreeFilter);
  return {
    descriptions: sections,
    tree: treePages,
    navigationTree: navigationTree[0]
  };
}