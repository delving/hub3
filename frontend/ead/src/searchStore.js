import {writable} from "svelte/store";
import {archiveSearch} from "./nav";

const {subscribe, set} = writable(null)

async function prepare() {
  const response = await fetch(archiveSearch())
  const json = await response.json()
  set(json)
}

export const searchStore = {
  subscribe,
  prepare
};