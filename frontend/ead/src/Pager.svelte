<div>
  <nav aria-label="Pages">
    <ul>
      <li><a href="#">« Vorige</a></li>
      <li><a href="#">1</a></li>
      <li><a href="#">2</a></li>
      <li><a href="#">3</a></li>
      <li><a href="#">4</a></li>
      <li><a href="#">5</a></li>
      <li><a href="#">Volgende »</a></li>
    </ul>
  </nav>
</div>

<style>
  select {
    display: inline;
    width: auto;
  }

  nav {
    margin: auto;
  }

  ul {
    list-style-type: none;
    display: flex;
    flex-direction: row;
    gap: 3px;
  }

  li:first-child a, li:last-child a {
    background-color: transparent;
  }

  a {
    text-decoration: none;
    color: black;
    padding: 10px 15px;
    background-color: white;
  }

  a:active {
    color: orange;
  }

  div {
    display: flex;
    flex-direction: row;
    gap: 1rem;
  }
</style>