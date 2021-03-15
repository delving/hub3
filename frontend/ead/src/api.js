import {parseEad} from "./ead-parser";
import xml from "./4.OSK.xml";
let ead = parseEad(xml);
console.log(ead)

export function getDescription() {
  return {}
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
      pages: ['htkrlthr'],
      matches: ead.tree.map((_, i) => i),
      hits: 150
    }
  } else if(params.cLevelId) {
    const needle = `data-identifier="${params.cLevelId}"`;
    for(let i = 0; i < ead.tree.length; i++) {
      const page = ead.tree[i];
      if (page.indexOf(needle) >= 0) {
        if (i > 0) response.pages.push({
          index: i - 1,
          html: ead.tree[i - 1]
        });
        response.pages.push({
          index: i,
          html: page
        });
        if (i < ead.tree.length - 2) response.pages.push({
          index: i + 1,
          html: ead.tree[i + 1]
        });
        break;
      }
    }
  }
  else {
    if(params.page >= 0) {
      response.pages.push({
        index: params.page,
        html: ead.tree[params.page]
      })
    } else {
      response.pages.push({
        index: 0,
        html: ead.tree[0]
      })
      if (ead.tree.length > 1) {
        response.pages.push({
          index: 1,
          html: ead.tree[1]
        })
      }
      if (ead.tree.length > 2) {
        response.pages.push({
          index: 2,
          html: ead.tree[2]
        })
      }
    }
  }

  return new Promise((resolve, reject) => {
    resolve(response)
  });
}
