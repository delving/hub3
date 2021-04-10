import {writable} from "svelte/store";

let isReady = false
const facets = new Set()
const {subscribe, update, set} = writable(null)

function createSearchString(search) {
  const queryBuilder = []
  Object.entries(search)
    .filter(p => p[0] !== 'searchString' && p[1])
    .forEach(p => queryBuilder.push(`${p[0]}=${p[1]}`))
  facets.forEach(url => queryBuilder.push(url))

  const searchString = queryBuilder.join('&');
  return searchString ? `?${searchString}` : searchString;
}

function updateQuery(changes) {
  update(currValue => {
    const updatedValue = {
      ...currValue,
      search: false,
      ...changes
    };
    return {...updatedValue, searchString: createSearchString(updatedValue)}
  })
}

function setFacetLink(facet, link, isSelected) {
  const property = `qf[]=${facet.field}:${link.value}`;
  if (isSelected) {
    facets.add(property);
  } else {
    facets.delete(property);
  }
  updateQuery({})
}

export function parseUrl() {
  const searchString = location.search
  const urlSearchParams = new URLSearchParams()

  isReady = true;
  set({
    q: urlSearchParams.get('q'),
    search: !!urlSearchParams.get('q'),
    sortBy: urlSearchParams.get('sortBy') || '',
    searchString: searchString
  })
}

export const queryStore = {
  subscribe: (callback) => {
    subscribe(currValue => {
      if (isReady)
        callback(currValue)
    })
  },
  updateQuery,
  parseUrl,
  setFacet: setFacetLink
};