<script>
  import {queryStore} from "./queryStore";

  export let query;
  let sortBy = query.sortBy;
  let sortOrder = sortBy && sortBy.indexOf('^') === 0 ? "desc" : "asc";

  function setSortBy(e) {
    sortBy = e.target.value;
    queryStore.updateQuery({sortBy})
  }

  function setSortOrder(e) {
    sortOrder = e.target.value;
    const sortAndOrder = sortOrder === 'asc' ? sortBy.replace('^',  '') : `^${sortBy}`;
    queryStore.updateQuery({sortBy: sortAndOrder})
  }
</script>

Sorteren op
<select on:change={setSortBy}>
  <option value=""></option>
  <option value="_score">Relevantie</option>
  <option value="meta.spec">Nummer toegang</option>
  <option value="ead-rdf_periodDesc">Periode</option>
</select>

Volgorde
<select on:change={setSortOrder}>
  <option value="asc">Oplopend</option>
  <option value="desc">Aflopend</option>
</select>
Resultaten
<select>
  <option>10</option>
  <option>20</option>
  <option>50</option>
</select>