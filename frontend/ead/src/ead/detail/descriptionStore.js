import {fetchDescription} from "../../api";
import {writable} from "svelte/store";
import {Pager} from "./pager";

const {subscribe, update} = writable(null);

async function patch(params, updateFunc) {
  const response = await fetchDescription(params)

  update(desc => ({
    ...desc,
    ...(updateFunc ? updateFunc(desc, response) : response),
  }));
  return response;
}

async function appendPage(page) {
  await patch({page}, (desc, response) => ({
    pages: [...desc.pages.slice(1), ...response.pages]
  }))
}

async function prependPage(page) {
  await patch({page}, (desc, response) => ({
    pages: [...response.pages, ...desc.pages.slice(0, desc.pages.length - 1)]
  }))
}

async function search(query) {
  const response = await patch({query, search: true});
  return new Pager(query, response, patch)
}

export const descriptionStore = {
  subscribe,
  search,
  prependPage,
  appendPage,
  prepare: patch
}
