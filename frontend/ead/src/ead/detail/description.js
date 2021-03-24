import {fetchDescription} from "../../api";
import {writable} from "svelte/store";

let currValue = {};
const {subscribe, set, update} = writable(currValue);
subscribe(description => currValue = description)

async function prepare() {
  const response = await fetchDescription({});
  set({
    ...response,
    activeSection: response.sections[response.activeIndex].html
  });
}

async function patchSection(index) {
  const section = currValue.sections[index];
  if(!section.html) {
    const response = await fetchDescription({index});
    section.html = response.html;
  }
  update(currValue => ({
    ...currValue,
    activeSection: section.html,
    activeIndex: index
  }))
}

export const description = {
  subscribe,
  prepare,
  patchSection
}