<script>
  export let timeline;

  const display = [
    {
      label: 'Naam',
      value: 'nk_name',
    },
    {
      label: 'Locatie',
      value: 'nk_location',
    },
    {
      label: 'Rol',
      value: 'nk_role',
    },
    {
      label: 'Start datum',
      value: 'nk_dateStart',
    },
    {
      label: 'Eind datum',
      value: 'nk_dateEnd',
    },
    {
      label: 'Bron',
      value: 'nk_archiveSource',
    }
  ]

  function open(e) {
    let parent = e.target.parentNode;
    while (parent.tagName !== 'ARTICLE') {
      parent = parent.parentNode;
    }
    const metadata = parent.querySelector('section');
    metadata.classList.toggle('is-open');
  }
</script>

<section class="timeline">
  {#each timeline as event}
    <article>
      <header>
        <h2>{event.nk_name}</h2>
        <img aria-hidden="true" src="/circle.svg" alt="Event">
        <div class="dates">
          <time>{event.nk_dateStart ? event.nk_dateStart[0] : 'Onbekend'}</time>
          <span aria-hidden="true">|</span>
          <time itemprop="End date">{event.nk_dateEnd ? event.nk_dateEnd[0] : 'Onbekend'}</time>
        </div>

        <button on:click={open}><span
          class="visually-hidden">Laat meer informatie zien</span>Meer
          informatie
        </button>
      </header>
      <section>
        <ul>
          {#each display as property}
            {#if property.value in event}
              <li>
                <label>
                  {property.label}
                  <p>{event[property.value]}</p>
                </label>
              </li>
            {/if}
          {/each}
        </ul>
      </section>
    </article>
  {/each}
</section>

<style type="text/scss">
  :global(.is-open) {
    display: block !important;
  }

  .timeline {
    padding-top: 4rem;
    border-left: 0.33rem solid #4CC9AA;
    max-width: 34rem;
    margin: 0 auto;
  }

  article {
    margin-bottom: 2rem;
    background: #FFFFFF;
  }

  header {
    position: relative;
    padding: 2rem 2rem 2rem 3rem;
  }

  img {
    position: absolute;
    transform: translate(-50%, -50%);
    top: 50%;
    left: -.16rem;
  }

  .dates {
    margin-bottom: 1rem;
    display: flex;
    font-size: .8rem;
    width: 12rem;
    padding-right: 3rem;
    flex-direction: column;
    position: absolute;
    top: 50%;
    left: -12rem;
    text-align: right;
    transform: translateY(-50%);
  }

  article section {
    display: none;
  }

  label {
    display: inline;
    font-weight: bold;

    p {
      font-weight: normal;
    }
  }
</style>