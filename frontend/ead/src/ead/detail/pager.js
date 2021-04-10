import {queryStore} from "../../search/queryStore";
import {searchStore} from "../../searchStore";

export class Pager {

  matchIndex = 0;
  hitCount = 0;

  constructor(query, searchResult) {
    this.query = query;
    this.hitCount = searchResult.hitCount;

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
    searchStore.setMatch(this.firstMatch())
  }

  firstMatch() {
    return this.matches[0]
  }

  searchPage(offset) {
    const currentMatch = this.matches[this.matchIndex];
    this.matchIndex += offset;
    const nextMatch = this.matches[this.matchIndex];

    if (currentMatch.page !== nextMatch.page) {
      queryStore.updateQuery({
        page: nextMatch.page,
        query: this.query
      })
    }

    searchStore.setMatch(nextMatch)
    return nextMatch;
  }

  previous() {
    if(this.matchIndex === 0) return;
    return this.searchPage(-1)
  }

  next() {
    if(this.matchIndex === this.matches.length - 1) return;
    return this.searchPage(1)
  }
}