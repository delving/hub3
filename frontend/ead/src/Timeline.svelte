<script>
function open(e) {
  let parent = e.target.parentNode;
  while(!parent.classList.contains('event')) {
    parent = parent.parentNode;
  }
  const metadata = parent.querySelector('.event__metadata');
  metadata.classList.toggle('is-open');
}
</script>

<section id="timeline" class="timeline-wrapper">
  <div class="timeline__events fadein ">
    <div id="1" class="event fadein" style="animation-delay: 0.3s;">
      <header class="event__header"><span class="icon  event__header__icon"><img
        src="/circle.svg"
        alt="Geboren"></span>
        <time class="date"><span class="date-start day-with-month">21 juni</span><span
          class="date-start year">1888</span>
        </time>
        <h2>Jan Oosterwijk is geboren in Assen</h2>
        <button on:click={open} class="event__metadata-toggle"><span class="visually-hidden">Laat meer informatie zien</span>Meer
          informatie
        </button>
      </header>
      <div class="event__metadata">
        <ul>
          <li class="publisher"><h6>Documentatiegroep 40-45</h6>
            <div class="publisher-event-title">Geboren op 21 juni 1888</div>
            <div class="publisher-event-details"><a
              href="http://database.documentatiegroep40-45.nl/details2.php?ID=14127"
              target="_blank" rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van Documentatiegroep 40-45</span></a><span>Vragen? <a
              href="mailto:info@documentatiegroep40-45.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>CBG|Centrum voor familiegeschiedenis: Duitse Overlijdensakten</h6>
            <div class="publisher-event-title">Geboren in Assen op 21 juni 1888</div>
            <div class="publisher-event-details"><a href="https://wiewaswie.nl/nl/detail/101889695" target="_blank"
                                                    rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van CBG|Centrum voor familiegeschiedenis: Duitse Overlijdensakten</span></a><span>Vragen? <a
              href="mailto:info@cbg.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>Stichting Oranjehotel: Doodenboeken</h6>
            <div class="publisher-event-title">Geboren in Assen op 21 juni 1888</div>
            <div class="publisher-event-details"><a
              href="http://www.gahetna.nl/collectie/archief/inventaris/gahetnascan/eadid/2.19.136/inventarisnr/2/scan-index/208"
              target="_blank" rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van Stichting Oranjehotel: Doodenboeken</span></a><span>Vragen? <a
              href="mailto:info@nationaalarchief.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>Arolsen Archives</h6>
            <div class="publisher-event-title">Geboren op 21 juni 1888</div>
            <div class="publisher-event-details"><a
              href="https://collections.arolsen-archives.org/en/search/people/392765" target="_blank" rel="noopener"
              tabindex="0" class="publisher-website">Website<span class="visually-hidden">Bezoek de website van Arolsen Archives</span></a>
            </div>
          </li>
          <li class="publisher"><h6>Kamp Dachau</h6>
            <div class="publisher-event-title">Geboren in Assen op 21 juni 1888</div>
            <div class="publisher-event-details"><a
              href="https://stevemorse.org/dachau/details.php?lastname=OOSTERWIJK&amp;firstname=Jan&amp;title=&amp;birthday=21&amp;birthmonth=Jun&amp;birthyear=1888&amp;birthplace=Assen&amp;from=&amp;town=Smilde&amp;street=A.%20134&amp;number=69305&amp;DateOfArrival=26%20May%201944%20Hert.&amp;disposition=gest.%2002%20Feb%201945&amp;comments=Check%20C&amp;category=Sch.%20Hol.&amp;ID=175312&amp;page=4293/Schü.&amp;disc=4&amp;image=529"
              target="_blank" rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van Kamp Dachau</span></a>
            </div>
          </li>
          <li class="publisher"><h6>Kamp Vught Carthoteek, Nationaal Archief</h6>
            <div class="publisher-event-title">Geboren op 21 juni 1888</div>
            <div class="publisher-event-details"></div>
          </li>
          <li class="publisher"><h6>Oorlogsgravenstichting</h6>
            <div class="publisher-event-title">Geboren in Assen op 21 juni 1888</div>
            <div class="publisher-event-details"><a
              href="https://www.oorlogsgravenstichting.nl/persoon/112662/jan-oosterwijk" target="_blank" rel="noopener"
              tabindex="0" class="publisher-website">Website<span class="visually-hidden">Bezoek de website van Oorlogsgravenstichting</span></a><span>Vragen? <a
              href="mailto:info@ogs.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>Stichting 18 September</h6>
            <div class="publisher-event-title">Geboren op 21 juni 1888</div>
            <div class="publisher-event-details"><a
              href="https://map.stolpersteine.app/nl/beilen/locaties/kanaalweg-207-smilde" target="_blank"
              rel="noopener"
              tabindex="0" class="publisher-website">Website<span class="visually-hidden">Bezoek de website van Stichting 18 September</span></a><span>Vragen? <a
              href="mailto:info@stichting18september.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
        </ul>
        <ul></ul>
      </div>
    </div>
    <div id="4" class="event fadein" style="animation-delay: 0.45s;">
      <header class="event__header"><span class="icon  event__header__icon"><img src="/cel.svg"
                                                                                 alt="Gevangen"></span>
        <time class="date"><span class="date-start day-with-month">22 november</span><span
          class="date-start year">1943</span><span class="date-separator-desktop">|</span><span
          class="date-separator-mobile">-</span><span class="day-with-month">24 mei</span><span class="year">1944</span>
        </time>
        <h2>Jan Oosterwijk zat gevangen in Kamp Vught</h2>
        <button on:click={open} class="event__metadata-toggle" style="color: rgb(204, 142, 58);"><span class="visually-hidden">Laat meer informatie zien</span>Meer
          informatie
        </button>
      </header>
      <div class="event__metadata">
        <ul>
          <li class="publisher slideup"><h6>Arolsen Archives</h6>
            <div class="publisher-event-title">Gevangen in Kamp Vught</div>
            <div class="publisher-event-details"><a
              href="https://collections.arolsen-archives.org/en/search/people/392765" target="_blank" rel="noopener"
              tabindex="0" class="publisher-website">Website<span class="visually-hidden">Bezoek de website van Arolsen Archives</span></a>
            </div>
          </li>
          <li class="publisher slideup"><h6>Kamp Vught Carthoteek, Nationaal Archief</h6>
            <div class="publisher-event-title">Gevangen in Kamp Vught op 22 november 1943 tot 24 mei 1944</div>
            <div class="publisher-event-details"></div>
          </li>
        </ul>
        <ul>
          <li class="description slideup"><h5>Kamp Vught</h5>
            <p>Kamp Vught (Duits: Konzentrationslager Herzogenbusch) was een concentratiekamp en doorgangskamp bij Vught
              in Noord-Brabant. Het kamp was in gebruik van 13 januari 1943 tot 16 september 1944. Na de oorlog werd het
              kampterrein gebruikt als Bewarings- en verblijfskamp voor (vermeende) politieke delinquenten.</p></li>
        </ul>
      </div>
    </div>
    <div id="5" class="event fadein" style="animation-delay: 0.6s;">
      <header class="event__header"><span class="icon  event__header__icon"><img src="/cel.svg"
                                                                                 alt="Gevangen"></span>
        <time class="date"><span class="date-start day-with-month">26 mei</span><span
          class="date-start year">1944</span><span class="date-separator-desktop">|</span><span
          class="date-separator-mobile">-</span><span class="day-with-month">Einddatum</span><span
          class="year">Onbekend</span></time>
        <h2>Jan Oosterwijk zat gevangen in Dachau</h2>
        <button on:click={open} class="event__metadata-toggle"><span class="visually-hidden">Laat meer informatie zien</span>Meer
          informatie
        </button>
      </header>
      <div class="event__metadata">
        <ul>
          <li class="publisher"><h6>Kamp Dachau</h6>
            <div class="publisher-event-title">Gevangen in Dachau op 26 mei 1944, tot onbekende datum</div>
            <div class="publisher-event-details"><a
              href="https://stevemorse.org/dachau/details.php?lastname=OOSTERWIJK&amp;firstname=Jan&amp;title=&amp;birthday=21&amp;birthmonth=Jun&amp;birthyear=1888&amp;birthplace=Assen&amp;from=&amp;town=Smilde&amp;street=A.%20134&amp;number=69305&amp;DateOfArrival=26%20May%201944%20Hert.&amp;disposition=gest.%2002%20Feb%201945&amp;comments=Check%20C&amp;category=Sch.%20Hol.&amp;ID=175312&amp;page=4293/Schü.&amp;disc=4&amp;image=529"
              target="_blank" rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van Kamp Dachau</span></a>
            </div>
          </li>
        </ul>
        <ul>
          <li class="description"><h5>Dachau</h5>
            <p>Dachau was een concentratiekamp van de SS, gelegen bij de Zuid-Duitse stad Dachau. Het was onafgebroken
              in
              gebruik van 22 maart 1933 tot aan de bevrijding door Amerikaanse troepen op 29 april 1945. In Dachau
              hebben
              206.000 personen van vele nationaliteiten gevangen gezeten, politieke gevangenen, verzetsmensen, Joden,
              Sinti en Roma, homoseksuelen en misdadigers, waarvan er zeker 41.500 stierven. Er hebben 1935 Nederlandse
              mannen en 200 Nederlandse vrouwen in Dachau gevangengezeten.</p></li>
        </ul>
      </div>
    </div>
    <div id="2" class="event fadein" style="animation-delay: 0.75s;">
      <header class="event__header"><span class="icon  event__header__icon"><img
        src="/circle.svg"
        alt="Omgekomen"></span>
        <time class="date"><span class="date-start day-with-month">2 februari</span><span
          class="date-start year">1945</span></time>
        <h2>Jan Oosterwijk is omgekomen in Dachau</h2>
        <button on:click={open} class="event__metadata-toggle"><span class="visually-hidden">Laat meer informatie zien</span>Meer
          informatie
        </button>
      </header>
      <div class="event__metadata">
        <ul>
          <li class="publisher"><h6>Documentatiegroep 40-45</h6>
            <div class="publisher-event-title">Omgekomen op 1 februari 1945</div>
            <div class="publisher-event-details"><a
              href="http://database.documentatiegroep40-45.nl/details2.php?ID=14127"
              target="_blank" rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van Documentatiegroep 40-45</span></a><span>Vragen? <a
              href="mailto:info@documentatiegroep40-45.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>CBG|Centrum voor familiegeschiedenis: Duitse Overlijdensakten</h6>
            <div class="publisher-event-title">Omgekomen in Dachau op 2 februari 1945</div>
            <div class="publisher-event-details"><a href="https://wiewaswie.nl/nl/detail/101889695" target="_blank"
                                                    rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van CBG|Centrum voor familiegeschiedenis: Duitse Overlijdensakten</span></a><span>Vragen? <a
              href="mailto:info@cbg.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>Stichting Oranjehotel: Doodenboeken</h6>
            <div class="publisher-event-title">Omgekomen in Dachau op 1 februari 1945</div>
            <div class="publisher-event-details"><a
              href="http://www.gahetna.nl/collectie/archief/inventaris/gahetnascan/eadid/2.19.136/inventarisnr/2/scan-index/208"
              target="_blank" rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van Stichting Oranjehotel: Doodenboeken</span></a><span>Vragen? <a
              href="mailto:info@nationaalarchief.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>Kamp Dachau</h6>
            <div class="publisher-event-title">Omgekomen in Kamp Dachau op 2 februari 1945</div>
            <div class="publisher-event-details"><a
              href="https://stevemorse.org/dachau/details.php?lastname=OOSTERWIJK&amp;firstname=Jan&amp;title=&amp;birthday=21&amp;birthmonth=Jun&amp;birthyear=1888&amp;birthplace=Assen&amp;from=&amp;town=Smilde&amp;street=A.%20134&amp;number=69305&amp;DateOfArrival=26%20May%201944%20Hert.&amp;disposition=gest.%2002%20Feb%201945&amp;comments=Check%20C&amp;category=Sch.%20Hol.&amp;ID=175312&amp;page=4293/Schü.&amp;disc=4&amp;image=529"
              target="_blank" rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van Kamp Dachau</span></a>
            </div>
          </li>
          <li class="publisher"><h6>Oorlogsgravenstichting</h6>
            <div class="publisher-event-title">Omgekomen in Dachau op 2 februari 1945</div>
            <div class="publisher-event-details"><a
              href="https://www.oorlogsgravenstichting.nl/persoon/112662/jan-oosterwijk" target="_blank" rel="noopener"
              tabindex="0" class="publisher-website">Website<span class="visually-hidden">Bezoek de website van Oorlogsgravenstichting</span></a><span>Vragen? <a
              href="mailto:info@ogs.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>Open Archieven: Overlijdensaktes</h6>
            <div class="publisher-event-title">Omgekomen in Dachau (Duitsland) op 2 februari 1945</div>
            <div class="publisher-event-details"><a
              href="https://www.openarch.nl/dar:168cbd2d-9d2b-4f6a-ad25-57f9c0525114" target="_blank" rel="noopener"
              tabindex="0" class="publisher-website">Website<span class="visually-hidden">Bezoek de website van Open Archieven: Overlijdensaktes</span></a><span>Vragen? <a
              href="mailto:info@openarch.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>WieWasWie: Overlijdensaktes</h6>
            <div class="publisher-event-title">Omgekomen in Dachau (Duitsland) op 2 februari 1945</div>
            <div class="publisher-event-details"><a href="https://www.wiewaswie.nl/nl/detail/15065207" target="_blank"
                                                    rel="noopener" tabindex="0" class="publisher-website">Website<span
              class="visually-hidden">Bezoek de website van WieWasWie: Overlijdensaktes</span></a><span>Vragen? <a
              href="mailto:helpdesk@wiewaswie.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
          <li class="publisher"><h6>Stichting 18 September</h6>
            <div class="publisher-event-title">Omgekomen in Dachau op 2 februari 1945</div>
            <div class="publisher-event-details"><a
              href="https://map.stolpersteine.app/nl/beilen/locaties/kanaalweg-207-smilde" target="_blank"
              rel="noopener"
              tabindex="0" class="publisher-website">Website<span class="visually-hidden">Bezoek de website van Stichting 18 September</span></a><span>Vragen? <a
              href="mailto:info@stichting18september.nl?subject=Oorlogslevens.nl&amp;body=Verzonden vanaf https%3A%2F%2Fwww.oorlogslevens.nl%2Ftijdlijn%2FJan-Oosterwijk%2F02%2F112662%3Flang%3Dnl"
              target="_blank" rel="noopener noreferrer">E-mail</a> deze organisatie</span></div>
          </li>
        </ul>
        <ul>
          <li class="description"><h5>Dachau</h5>
            <p>Dachau was een concentratiekamp van de SS, gelegen bij de Zuid-Duitse stad Dachau. Het was onafgebroken
              in
              gebruik van 22 maart 1933 tot aan de bevrijding door Amerikaanse troepen op 29 april 1945. In Dachau
              hebben
              206.000 personen van vele nationaliteiten gevangen gezeten, politieke gevangenen, verzetsmensen, Joden,
              Sinti en Roma, homoseksuelen en misdadigers, waarvan er zeker 41.500 stierven. Er hebben 1935 Nederlandse
              mannen en 200 Nederlandse vrouwen in Dachau gevangengezeten.</p></li>
        </ul>
      </div>
    </div>
  </div>
</section>

<style type="text/scss">
  #timeline {
    -webkit-locale: "nl";
    box-sizing: border-box;
    line-height: 1.15;
    -webkit-text-size-adjust: 100%;
    font-size: 100%;
    background-color: #F6F7EB;

    -webkit-text-size-adjust: 100%;
    font-size: 100%;
    box-sizing: border-box;
    margin: 0;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
    line-height: 1.6;
    color: #212121;
    background-color: #F6F7EB;
    overflow-x: hidden;

    -webkit-text-size-adjust: 100%;
    font-size: 100%;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
    line-height: 1.6;
    color: #212121;
    box-sizing: border-box;
    display: block;
    background: #F6F7EB;

    -webkit-text-size-adjust: 100%;
    font-size: 100%;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
    line-height: 1.6;
    color: #212121;
    box-sizing: border-box;
    display: block;
    min-height: 60vh;
    position: relative;
    width: 100%;
    padding-left: 8rem;
    margin-left: 0;
  }

  .timeline__events {
    -webkit-text-size-adjust: 100%;
    font-size: 100%;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
    line-height: 1.6;
    color: #212121;
    box-sizing: border-box;


    padding-top: 4rem;
    border-left: 0.33rem solid #4CC9AA;
    max-width: 34rem;
    margin: 0 auto;
  }

  .event {
    -webkit-text-size-adjust: 100%;
    font-size: 100%;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
    line-height: 1.6;
    box-sizing: border-box;

    margin-bottom: 2rem;
    position: relative;
    color: #404040;
    background: #FFFFFF;
    transition: box-shadow .3s;
    animation-delay: 0.6s;
  }

  .event__header {

    -webkit-text-size-adjust: 100%;
    font-size: 100%;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
    line-height: 1.6;
    color: #404040;
    box-sizing: border-box;
    display: block;
    position: relative;
    padding: 2rem 2rem 2rem 3rem;
  }

  .event__header__icon {
    -webkit-text-size-adjust: 100%;
    font-size: 100%;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
    line-height: 1.6;
    color: #404040;
    box-sizing: border-box;
    width: 4rem;
    display: flex;
    justify-content: center;
    position: absolute;
    transform: translate(-50%, -50%);
    top: 50%;
    left: -.16rem;
  }

  img {
    -webkit-text-size-adjust: 100%;
    font-size: 100%;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
    line-height: 1.6;
    color: #404040;
    box-sizing: border-box;
    border-style: none;
  }

  .date {
    -webkit-text-size-adjust: 100%;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
    line-height: 1.6;
    color: #404040;
    box-sizing: border-box;
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

  .day-with-month {
    -webkit-text-size-adjust: 100%;
    line-height: 1.6;
    font-size: .8rem;
    text-align: right;
    box-sizing: border-box;
    color: #404040;
    margin-right: .3rem;
    font-family: "Quatro Sans W00", "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    font-weight: 400;
  }

  .date-start year {
    -webkit-text-size-adjust: 100%;
    line-height: 1.6;
    text-align: right;
    box-sizing: border-box;
    color: #404040;
    font-family: "Quatro Sans W00", "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    display: block;
    font-size: 2rem;
    font-weight: 300;
  }

  h2 {
    -webkit-text-size-adjust: 100%;
    color: #404040;
    box-sizing: border-box;
    font-family: "Quatro Sans W00", "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    margin: .25rem 0;
    font-weight: 500;
    font-size: 1.5rem;
    line-height: 1.75rem;
  }

  button {
    -webkit-text-size-adjust: 100%;
    box-sizing: border-box;
    line-height: 1.15;
    margin: 0;
    text-transform: none;
    appearance: none;
    padding: 0;
    border: none;
    border-radius: 0;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    background-color: transparent;
    overflow: visible;
    cursor: pointer;
    margin-top: .5rem;
    color: #AF6708;
    transition: color .2s;
    font-size: 1rem;
  }

  .visually-hidden {
    -webkit-text-size-adjust: 100%;
    line-height: 1.15;
    text-transform: none;
    font-family: "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
    cursor: pointer;
    color: #AF6708;
    font-size: 1rem;
    box-sizing: border-box;
    width: 1px !important;
    height: 1px !important;
    padding: 0 !important;
    position: absolute !important;
    clip: rect(1px, 1px, 1px, 1px);
    word-wrap: normal;
    white-space: nowrap;
    overflow: hidden;
  }




</style>