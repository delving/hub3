import {fetchTree} from "./api";
import {writable} from "svelte/store";
import {Pager} from "./pager";

let navigationTree = true;
const {subscribe, update} = writable({pages: []});

async function patch(params, updateFunc) {
  const response = await fetchTree({
    navigationTree,
    ...params
  })
  navigationTree = false;

  update( tree => ({
    ...tree,
    ...(updateFunc ? updateFunc(tree, response) : response),
    isReady: true
  }));
  return response;
}

async function appendPage(page) {
  await patch({page}, (tree, response) => ({
    pages: [...tree.pages.slice(1), ...response.pages]
  }))
}

async function prependPage(page) {
  await patch({page}, (tree, response) => ({
    pages: [...response.pages, ...tree.pages.slice(0, tree.pages.length - 1)]
  }))
}

async function search(query) {
  const response = await patch({query, search: true});
  return new Pager(query, response, patch)
}

export const treeStore = {
  subscribe,
  search,
  prependPage,
  appendPage,
  prepare: patch
};