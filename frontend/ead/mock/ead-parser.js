const jsdom = require("jsdom")
const {JSDOM} = jsdom
const dom = new JSDOM()
global.DOMParser = dom.window.DOMParser
const parser = new DOMParser();

const Node = {
  TEXT_NODE: 3
};

const COPY_ATTRIBUTES_OF = {
  'list': () => {},
  'extptr': () => {},
  'c': (jsonParent, node) => {
    jsonParent.level = node.getAttribute('level')
  },
  'unitid': () => {}
};

let counter = 0;
let charCount = 0
const PROPERTY_WEIGHT = 12;
const CHILD_WEIGHT = 4;

function addTextTo(textNode, node) {
  const content = textNode.textContent.trim();
  if (content) {
    charCount += content.length + PROPERTY_WEIGHT;
    node.inline.push({text: content});
  }
}

function Builder(filter, limit) {

  const pages = [];
  const parents = [];

  function createPage(jsonParent) {
    const findPageStart = parents.find(p => !p.continued) || jsonParent;
    const cloneOfPage = JSON.parse(JSON.stringify(findPageStart));
    filter(cloneOfPage);
    pages.push({
      index: pages.length,
      parents: parents
        .filter(p => !p.continued)
        .map(p => p.tagName),
      nodes: cloneOfPage
    });
    parents.forEach(p => p.continued = true);
    charCount = 0;
  }

  function toJson(jsonParent, node) {
    if (charCount > limit) {
      createPage(jsonParent);
    }
    parents.push(jsonParent);

    if (node.nodeType === Node.TEXT_NODE) {
      jsonParent.text = node.textContent;
    } else {
      const tagName = node.tagName.toLowerCase();
      if(tagName === 'c') jsonParent.id = counter++;
      if(tagName in COPY_ATTRIBUTES_OF) {
        COPY_ATTRIBUTES_OF[tagName](jsonParent, node);
      }
      jsonParent.tagName = tagName;
      charCount += tagName.length + PROPERTY_WEIGHT;
      const children = node.childNodes;
      if (children.length > 0) {
        charCount += PROPERTY_WEIGHT;
        const firstChild = children[0];

        if (children.length === 1 && firstChild.nodeType === Node.TEXT_NODE) {
          const text = firstChild.textContent.trim();
          if(text) {
            jsonParent.text = text;
          }
        } else {
          jsonParent.inline = [];
          for (const child of children) {
              charCount += CHILD_WEIGHT;
              if (child.nodeType === Node.TEXT_NODE) {
                addTextTo( child, jsonParent);
              } else {
                jsonParent.inline.push(toJson({}, child));
              }
          }
        }
      }
    }
    parents.pop();
    return jsonParent;
  }

  this.toJson = function (jsonParent, node) {
    counter = 0;
    toJson(jsonParent, node);
    if(charCount > 0) {
      createPage(jsonParent);
    }
    return pages;
  };
}

function isSeries(cNode) {
  return cNode.level === 'series' || cNode.level === 'subseries';
}

function navTreeFilter(parent) {
  if(parent.tagName === 'c' && !isSeries(parent)) {
    parent.inline = null;
    return false;
  }
  if (parent.tagName === 'unittitle') {
    return true;
  }

  if (parent.inline) {
    const acceptedChildren = []
    for (const child of parent.inline) {
      const hasMatch = navTreeFilter(child);
      if (hasMatch) {
        acceptedChildren.push(child);
      }
    }
    parent.inline = acceptedChildren;
  }
  return (parent.inline && parent.inline.length > 0);
}

function xmlToJson(node, limit, filter) {
  const rootJson = {};
  const builder = new Builder((filter || function () {
    return true
  }), limit);
  return builder.toJson(rootJson, node);
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
  const descriptionPages = xmlToJson(dsc, 10000);

  const tree = xmlDoc.querySelector('archdesc > dsc[type="combined"]');
  const treePages = xmlToJson(tree, 10000);
  const navigationTree = xmlToJson(tree, Number.MAX_SAFE_INTEGER, navTreeFilter);
  return {
    description: {
      sections: titles,
      pages: descriptionPages
    },
    tree: treePages,
    navigationTree: navigationTree[0].nodes
  };
}