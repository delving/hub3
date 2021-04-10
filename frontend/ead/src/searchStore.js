import {writable} from "svelte/store";
import {createLink, route} from "./router";
import {config} from "./config";
import {Pager} from "./ead/detail/pager";

const {subscribe, update, set} = writable(null)

function createRequestBody(urlDef, context) {
  if (!urlDef.body) return null;
  const body = {}
  for (const property of urlDef.body) {
    if (property.requires && !context[property.requires.substring(1)])
      continue;

    if (typeof property.value === 'string' && property.value.indexOf(':') === 0) {
      const value = context[property.value.substring(1)];
      if (value) {
        body[property.key] = value;
      }
    } else {
      body[property.key] = property.value
    }
  }
  return JSON.stringify(body);
}

function createContext(urlDef, baseContext) {
  return {
    baseUrl: config.baseUrl,
    ...route.values,
    ...baseContext,
  };
}

async function prepare(baseContext) {
  const urlDef = route.config.requestUrls.search
  const context = createContext(urlDef, baseContext)

  const result = await search(urlDef, context);
  if (result.hitCount >= 0) {
    const hitPager = new Pager(baseContext.q, result)
    set({
      hitPager,
      match: hitPager.firstMatch(),
      ...result,
    })
  } else {
    update(currValue => ({
      hitPager: (currValue || {}).hitPager,
      ...result,
    }))
  }
}

async function search(urlDef, context) {
  console.log(urlDef, context)
  const url = createLink(urlDef, context)
  const method = urlDef.method || 'get'
  const body = createRequestBody(urlDef, context);

  const request = new Request(url, {
    headers: {
      'Content-Type': 'application/json;charset=utf-8'
    },
    method,
    body
  })
  const response = await fetch(request)
  return await response.json()
}

async function appendPage(page) {
  const urlDef = route.config.requestUrls.search
  const result = await search(urlDef, createContext(urlDef, {page}))
  update(currValue => ({
    pages: [...currValue.pages.slice(1), ...result.pages]
  }))
}

async function prependPage(page) {
  const urlDef = route.config.requestUrls.search
  const result = await search(urlDef, createContext(urlDef, {page}))
  update(currValue => ({
    pages: [...result.pages, ...currValue.pages.slice(0, currValue.pages.length - 1)]
  }))
}

function setMatch(match) {
  update(currValue => ({
    ...currValue,
    match
  }))
}

export const searchStore = {
  subscribe,
  setMatch,
  prepare,
  appendPage,
  prependPage
};