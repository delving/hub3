import {parseEad} from "./ead-parser";
import xml from "./4.OSK.xml";

let ead = parseEad(xml);
console.log(ead)

function occurrences(string, subString, allowOverlapping) {

  string += "";
  subString += "";
  if (subString.length <= 0) return (string.length + 1);

  var n = 0,
    pos = 0,
    step = allowOverlapping ? 1 : subString.length;

  while (true) {
    pos = string.indexOf(subString, pos);
    if (pos >= 0) {
      ++n;
      pos += step;
    } else break;
  }
  return n;
}

export function fetchDescription(params) {
  return new Promise((resolve) => {
    const sections = ead.descriptions;
    let response;
    if (typeof params.index !== 'number') {
      response = {
        pageCount: sections.length,
        activeIndex: 0,
        sections: sections.map((section, i) => ({
          title: section.title,
          index: i,
          html: i === 0 ? ead.descriptions[0].html : undefined
        })),
      };
    } else {
      response = {
        index: params.index,
        html: ead.descriptions[params.index].html
      };
    }
    console.log('fetchDescription', params, 'response', response)
    resolve(response)
  })
}

export function fetchTree(params) {
  let response = {
    pageCount: ead.tree.length,
    pages: []
  };
  if (params.navigationTree) {
    response = {...response, navigationTree: ead.navigationTree}
  }

  if (params.search === true) {
    let count = 0
    let matches = [];
    for (let i = 0; i < ead.tree.length; i++) {
      const page = ead.tree[i]
      let hits = occurrences(page, params.query)
      count += hits
      if (page.indexOf(params.query) >= 0) {
        matches.push({page: i, hitCount: occurrences(page, params.query)})
        if (response.pages.length === 0) {
          if (i !== 0) {
            response.pages.push({
              index: i - 1,
              html: ead.tree[i - 1]
            })
          }
          response.pages.push({
            index: i,
            html: page
          })
          if (i < ead.tree.length - 2) {
            response.pages.push({
              index: i + 1,
              html: ead.tree[i + 1]
            })
          }
        }
      }
    }
    response = {...response, matches, hitCount: count}
  } else if (params.cLevelId) {
    const needle = `data-identifier="${params.cLevelId}"`;
    for (let i = 0; i < ead.tree.length; i++) {
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
  } else if (params.page >= 0) {
    delete response.pageCount;
    if (params.query && params.page > 0) {
      response.pages.push({
        index: params.page - 1,
        html: ead.tree[params.page - 1]
      })
    }
    response.pages.push({
      index: params.page,
      html: ead.tree[params.page]
    })
    if (params.query && params.page < ead.tree.length - 1) {
      response.pages.push({
        index: params.page + 1,
        html: ead.tree[params.page + 1]
      })
    }
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

  if (params.query) {
    for (const page of response.pages) {
      page.html = page.html.replace(new RegExp(params.query, 'g'), m => {
        return `<em class="dhcl">${m}</em>`;
      })
    }
  }

  console.log('fetchTree', params, 'response', response)
  return new Promise((resolve, reject) => {
    resolve(response)
  });
}
