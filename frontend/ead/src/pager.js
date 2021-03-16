import {fetchTree} from "./api";

export class Pager {

  matches = [];
  matchIndex = 0;
  pageNumbers = [];
  hitCount = 0;
  currentContainer;

  constructor(container) {
    this.container = container;
  }

  async search(params) {
    this.query = params.query;
    const result = await fetchTree({
      ...params,
      search: true,
    })

    for (const match of result.matches) {
      for (let i = 0; i < match.hitCount; i++) {
        this.matches.push({
          page: match.page,
          index: i
        });
      }
    }
    console.log(this.matches)

    this.hitCount = result.hitCount
    return result
  }

  async jump() {
    const match = this.matches[this.matchIndex];
    const matchContainers = this.container.querySelectorAll(`.page[data-index="${match.page}"] .dhcl`)
    console.log(match, matchContainers);
    const container = matchContainers[match.index];
    container.classList.add('active');
    container.scrollIntoView();
    this.currentContainer = container;
  }

  async searchPage(offset) {
    const currentMatch = this.matches[this.matchIndex];
    this.matchIndex += offset;
    const nextMatch = this.matches[this.matchIndex];
    this.currentContainer.classList.remove('active');
    if (currentMatch.page === nextMatch.page) {
      await this.jump();
      return null;
    }

    return await fetchTree({
      page: nextMatch.page,
      query: this.query
    })
  }

  async previous() {
    if(this.matchIndex === 0) return;
    return await this.searchPage(-1)
  }

  async next() {
    if(this.matchIndex === this.matches.length - 1) return;
    return await this.searchPage(1)
  }
}