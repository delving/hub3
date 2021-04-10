<div class="border-top bg-light">
  <p>
    Resultaten per pagina
    <select class="form-select">
      <option>10</option>
      <option>20</option>
      <option>50</option>
    </select>
  </p>

  <nav aria-label="Pages">
    <ul class="pagination">
      <li class="page-item"><a class="page-link" href="#">Previous</a></li>
      <li class="page-item"><a class="page-link" href="#">1</a></li>
      <li class="page-item"><a class="page-link" href="#">2</a></li>
      <li class="page-item"><a class="page-link" href="#">3</a></li>
      <li class="page-item"><a class="page-link" href="#">Next</a></li>
    </ul>
  </nav>
</div>

<style>
  select {
    display: inline;
    width: auto;
  }

  div {
    padding-top: 0.4rem;
    display: flex;
    flex-direction: row;
    gap: 1rem;
  }
</style>