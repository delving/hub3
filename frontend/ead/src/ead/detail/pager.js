export class Pager {

  matchIndex = 0;
  hitCount = 0;
  currentContainer;

  constructor(query, searchResult, fetcher) {
    this.query = query;
    this.fetcher = fetcher;

    this.matches = Array(searchResult.hitCount)
    let n = 0;
    for (const match of searchResult.matches) {
      for (let i = 0; i < match.hitCount; i++) {
        this.matches[n] = {
          page: match.page,
          index: i,
          displayString: `${n + 1}`,
          isFirst: n === 0,
          isLast: n === searchResult.hitCount - 1,
        };
        n++;
      }
    }
  }

  firstMatch() {
    return this.matches[0]
  }

  async searchPage(offset) {
    const currentMatch = this.matches[this.matchIndex];
    this.matchIndex += offset;
    const nextMatch = this.matches[this.matchIndex];

    if (currentMatch.page !== nextMatch.page) {
      await this.fetcher({
        page: nextMatch.page,
        query: this.query
      })
    }

    return nextMatch;
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