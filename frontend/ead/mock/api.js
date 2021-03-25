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

function fetchDescription(ead, params) {
  console.log(params)
  let response = {
    sections: ead.description.sections
  };
  return pages(ead.description.pages, params, response);
}

function fetchTree(ead, params) {
  console.log(params)
  let response = {};
  if (params.navigationTree) {
    response = {...response, navigationTree: ead.navigationTree}
  }

  return pages(ead.tree, params, response);
}

function pages(eadPages, params, response) {
  response = {
    ...response,
    pageCount: eadPages.length,
    pages: []
  };

  if (params.search === true) {
    let count = 0
    let matches = [];
    for (let i = 0; i < eadPages.length; i++) {
      const page = eadPages[i]
      let hits = occurrences(page, params.query)
      count += hits
      if (page.indexOf(params.query) >= 0) {
        matches.push({page: i, hitCount: occurrences(page, params.query)})
        if (response.pages.length === 0) {
          if (i !== 0) {
            response.pages.push({
              index: i - 1,
              html: eadPages[i - 1]
            })
          }
          response.pages.push({
            index: i,
            html: page
          })
          if (i < eadPages.length - 2) {
            response.pages.push({
              index: i + 1,
              html: eadPages[i + 1]
            })
          }
        }
      }
    }
    response = {...response, matches, hitCount: count}
  } else if (params.cLevelId) {
    const needle = `data-identifier="${params.cLevelId.substring(1)}"`;
    for (let i = 0; i < eadPages.length; i++) {
      const page = eadPages[i];
      if (page.indexOf(needle) >= 0) {
        if (i > 0) response.pages.push({
          index: i - 1,
          html: eadPages[i - 1]
        });
        response.pages.push({
          index: i,
          html: page
        });
        if (i < eadPages.length - 2) response.pages.push({
          index: i + 1,
          html: eadPages[i + 1]
        });
        break;
      }
    }
  } else if (params.page >= 0) {
    delete response.pageCount;
    if (params.query && params.page > 0) {
      response.pages.push({
        index: params.page - 1,
        html: eadPages[params.page - 1]
      })
    }
    response.pages.push({
      index: params.page,
      html: eadPages[params.page]
    })
    if (params.query && params.page < eadPages.length - 1) {
      response.pages.push({
        index: params.page + 1,
        html: eadPages[params.page + 1]
      })
    }
  } else {
    response.pages.push({
      index: 0,
      html: eadPages[0]
    })
    if (eadPages.length > 1) {
      response.pages.push({
        index: 1,
        html: eadPages[1]
      })
    }
    if (eadPages.length > 2) {
      response.pages.push({
        index: 2,
        html: eadPages[2]
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
  return response;
}

module.exports = {
  fetchTree,
  fetchDescription
};
