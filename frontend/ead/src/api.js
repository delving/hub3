import {parseEad} from "./ead-parser";
import xml from "./4.OSK.xml";
let ead = parseEad(xml);
console.log(ead)

export function getDescription() {
  return
}

export function getTree(params) {
  let response = {
    pageCount: ead.tree.length,
    pages: []
  };
  if (params.navigationTree) {
    response = {...response, navigationTree: ead.navigationTree}
  }

  if(params.query) {
    response = {
      ...response,
      matches: ead.tree.map((_, i) => i),
      hits: 150
    }
  } else if(params.cLevelId) {
    const needle = `data-identifier="${params.cLevelId}"`;
    for(let i = 0; i < ead.tree.length; i++) {
      const page = ead.tree[i];
      if (page.indexOf(needle) >= 0) {
        if (i > 0) response.pages.push(ead.tree[i - 1]);
        response.pages.push(page);
        if (i < ead.tree.length - 2) response.pages.push(ead.tree[i + 1]);
        break;
      }
    }
  }
  else {
    response.pages.push(ead.tree[params.index || 0])
  }

  return new Promise((resolve, reject) => {
    resolve(response)
  });
}

function getTreePage(index) {
  return new Promise((resolve, reject) => {
    resolve({

      page: ead.tree[index]
    });
  });
}

export function getTreePageForCLevel(cLevelId) {
  return new Promise((resolve, reject) => {

  });
}