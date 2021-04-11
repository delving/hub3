import {writable} from "svelte/store";

const {set, subscribe} = writable({})

export const store = {
  set,
  subscribe
}