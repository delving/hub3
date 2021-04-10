const jsdom = require("jsdom")
const {JSDOM} = jsdom
const dom = new JSDOM()
global.DOMParser = dom.window.DOMParser
const parser = new DOMParser();

const Node = {
  TEXT_NODE: 3
};

function toSpan(node, copyAttributes, depth, builder) {
  const tagName = node.tagName.toLowerCase();
  const classNames = [tagName]
  const attributes = []

  let identifier = null;
  if (tagName === 'c') {
    classNames.push(`c${depth}`);
    for (const child of node.childNodes) {
      if (child.tagName === "did") {
        for (const sibling of child.childNodes) {
          if (sibling.tagName === "unitid") {
            if (sibling.getAttribute('type') !== 'blank' && sibling.getAttribute('type') !== 'handle') {
              identifier = sibling.textContent || sibling.getAttribute('identifier');
            }
          }
        }
      }
    }

    const cId = identifier || '@';
    builder.path.push(cId);
    const path = builder.path.join('~');
    attributes.push(`data-identifier="${path}"`)
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
    isCLevel: tagName === 'c',
    closingTags: '</span>',
    toHtml: function (extraClasses) {
      const attrs = `${attributes.join(' ')} class="${classNames.concat(extraClasses).join(' ')}"`;
      if(tagName === 'c') {
        this.closingTags = '</li></ul>';
        const listType = node.getAttribute('level') === 'file' ? 'list-group-flush' : 'list-group';
        const cAttrs = `${attributes.join(' ')} class="${classNames.concat(extraClasses).concat(listType).join(' ')}"`;
        return `<ul ${cAttrs}><li class="list-group-item">`;
      }
      if (tagName === 'head') {
        this.closingTags = '</h1>';
        return `<h1>`;
      }
      if(tagName === 'p') {
        this.closingTags = '</p>';
        return `<p>`;
      }
      if(tagName === 'tgroup') {
        this.closingTags = '</table>';
        return `<table class="table">`;
      }
      if(tagName === 'thead') {
        this.closingTags = '</thead>';
        return `<thead>`;
      }
      if(tagName === 'tbody') {
        this.closingTags = '</tbody>';
        return `<tbody>`;
      }
      if(tagName === 'row') {
        this.closingTags = '</tr>';
        return `<tr>`;
      }
      if(tagName === 'entry') {
        this.closingTags = '</td>';
        return `<td>`;
      }
      if(tagName === 'list') {
        this.closingTags = '</ul>';
        return `<ul class="list-group" ${attrs}>`;
      }
      if(tagName === 'item') {
        this.closingTags = '</li>';
        return `<li class="list-group-item">`;
      }
      if(tagName === 'descgrp' || tagName === 'did') {
        this.closingTags = '</section>';
        return '<section>';
      }
      if(tagName === 'unitdate') {
        this.closingTags = '</time>';
        return '<time>';
      }
      if(tagName === 'physloc') {
        this.closingTags = '</address>';
        return '<address>';
      }
      return `<span ${attrs}>`;
    },
  };
}

function isText(node) {
  const children = node.childNodes;
  return children && children.length === 1 && children[0].nodeType === Node.TEXT_NODE;
}

const COPY_ATTRIBUTES_OF = {
  'LIST': true,
  'EXTPTR': true,
  'C': true,
  'UNITID': true
};

function nodeToHtml(node, builder, depth = 0) {

  if (node.nodeType !== Node.TEXT_NODE) {
    const isOpen = builder.openTag(node, depth);
    for (const child of [...node.childNodes]) {
      nodeToHtml(child, builder, depth + 1);
    }
    if (isOpen) builder.closeTag();
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
  path = []

  constructor(limit, accept) {
    this.limit = limit;
    this.accept = accept;
  }

  openTag(node, depth) {
    const label = node.getAttribute('label')
    if (label) {
      this.html.push(`<span class="inline-label">${label}</span>`);
    }

    if (this.accept(node)) {
      const spanBuilder = toSpan(node, node.tagName.toUpperCase() in COPY_ATTRIBUTES_OF, depth, this);
      if (node.tagName.toLowerCase() === 'c') this.cLevelCount++;
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
    this.closedTagCount++;
    const pop = this.open.pop();
    this.html.push(pop.closingTags);
    if (pop.isCLevel) {
      this.path.pop();
    }
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
    while (target && target.tagName.toLowerCase() !== 'c') {
      target = target.parentNode;
    }
    return navTreeFilter(target);
  }
  return false;
}

function treeFilter(node) {
  const tagName = node.tagName;
  if (!tagName) return treeFilter(node.parentNode);
  if (tagName.toLowerCase() !== 'unitid') return true;
  const type = node.getAttribute("type");
  return type !== "blank" && type !== "handle";
}

module.exports = function (eadXml) {
  const xmlDoc = parser.parseFromString(eadXml, "text/xml");
  let sections = xmlDoc.querySelectorAll('eadheader > filedesc, archdesc > did, archdesc > descgrp');
  const titles = [...sections].map((section, i) => ({
    index: i,
    title: section.firstChild.textContent.trim() || section.childNodes[1].textContent.trim(),
  }))
  const dsc = dom.window.document.createElement('archdesc');
  sections.forEach(section => dsc.appendChild(section));
  const descriptionPages = toHtml(dsc, 200);

  const tree = xmlDoc.querySelector('archdesc > dsc[type="combined"]');
  const treePages = toHtml(tree, 250, treeFilter);
  const navigationTree = toHtml(tree, Number.MAX_SAFE_INTEGER, navTreeFilter);
  return {
    description: {
      sections: titles,
      pages: descriptionPages
    },
    tree: treePages,
    navigationTree: navigationTree[0]
  };
}