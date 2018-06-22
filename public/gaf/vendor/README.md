GENERAL ARCHIVE FUNCTIONS (gaf) jQuery plugins
==============================================
The GAF jQuery plugins can be used to get the archives of various archival
institutions. It is a javascript, html and css frontend to a JSON based
backend api. This is meant as a plug and play implementation on various
website to get archives quick and easy. 
It's only dependency on the site is jQuery 2. The other dependencies are 
bundled inside the distribution javascript file.
It uses handlebars templates and compiled SCSS styling to display the 
results. The templates and css can be overridden via the configuration
options or creating overriding css rules in your own css files.


Archival institutions supported
-------------------------------
Content from these archives are supported. The integer identifier is used
as the configuration option archivalInstituteId in the GAFARCHIVES plugin.


* Het Nationaal Archief (id:2437)
* Historisch Centrum Overrijssel (id:808)
* Brabants Historisch Informatie Centrum (id:797)
* Drents Archief (id:2434)
* Tresoar (id:834)
* Gelders Archief (id:786)
* Groninger Archief (id:789)
* Utrechts Archief (id:811)
* Noord-Hollands Archief (id:803)
* Regionaal Historisch Centrum Limburg (id:796)
* Zeeuws Archief (id:813)


Use cases
---------
* List the archives of one archival institution via GAFARCHIVES.
* Paginate the list of archives of the institution.
* Search the archives of all supported institutions via GAFSEARCH.
* Facet filter the results.
* Paginate the search results.
* Display text archives.
* Display images of archives.
* Deep link archives and search results via the HTML5 history api.


Browser compatibility
---------------------
* IE 11+
* Chrome 30+
* Firefox 50+
* Safari 10+
* Opera 10.6+


Implementation
--------------
Copy the complete directory to your local javascript distribution folder to 
/vendor/gaf. Link the javascript and css files in your page.

```html
<link href="/vendor/gaf/css/main.min.css" rel="stylesheet">
<script src="/vendor/gaf/js/bundle.min.js" type="text/javascript"></script>
```

See next paragraphs for the plugin implementations.

Usage GAFARCHIVES
-----------------
For listing the archives of one institute attach the GAFARCHIVES plugin to an empty
div container with an id '#gaf' and configure it with the configuration options.
With these options we can list the archives on the page /archives. You can place
the plugin on any page you want but the baseUrl needs to reflect the page on which
you implemented the plugin.

```javascript
$(document).ready(function(){
  var options = {
    baseUrl: '/archives',
    sourceDirectory: '/vendor/gaf/',
    archivalInstituteId: 2437,
    resultOptions: [{value: 10}, {value: 20}, {value: 50}],
    resultsPerPage: '20',
    language: 'NL',
    scrollToTop: true
  };
  $('#gaf').GAFARCHIVES(options);
});
```

### OPTIONS GAFARCHIVES
####`baseUrl` default '/archives'
Controls the base pagination links. Example: /archives?page=2

####`sourceDirectory` default '/'
Sets the base url for loading the template files. Should be set to where you place
these plugin files. Example: /vendor/gaf/

####`archivalInstituteId` default ''
Id (integer) to control from which archival institute we need to get content 
from. See the section *Archival institutions supported* for the list of
institutions and their identifier. 

####`resultOptions` default [{value: 10}, {value: 20}, {value: 50}]
Controls the dropdown options on how many results per page the user can see.

####`resultsPerPage` default 20
Controls how many results per page are shown on initial page load.

####`language` default 'NL'
Sets the default language of the plugin and translates the applicable labels.
The content of the archives are not translated. You can override the labels,
see `translatedLabels` option.

####`translatedLabels`
A json object containing the languages, the label keys and their translated
value. You can override each language and label by setting the correct value.

```javascript
translatedLabels: {
  NL: {
    findingAidNo: 'Toegangsnummer',
    findingAidTitle: 'Titel',
    unitDate: 'Periode',
    first: 'Eerste',
    previous: 'Vorige',
    next: 'Volgende',
    last: 'Laatste'
  },
  EN: {
    findingAidNo: 'Access number',
    findingAidTitle: 'Title',
    unitDate: 'Period',
    first: 'First',
    previous: 'Previous',
    next: 'Next',
    last: 'Last'
  },
  FY: {
    findingAidNo: 'Koade',
    findingAidTitle: 'Titel',
    unitDate: 'Perioade',
    first: 'Earst',
    previous: 'Foarige',
    next: 'Folgjende',
    last: 'Lêste'
  }
}
```

Example override Toegangsnummer label and change it to 'Toegang ID':

```javascript
translatedLabels: {
  NL: {
    findingAidNo: 'Toegang ID'    
  }
}
```

####`resultLanguageOptions` default [{value: 'NL'}, {value: 'EN'}, {value: 'FY'}]
Controls which languages are permitted in the translations.

####`mainTemplate` default 'templates/archives.hbs'
Sets which handlebars template the archive results are rendered with. If you 
want to override the html it is best you copy archives.hbs to archives-2.hbs
and use that as the mainTemplate option.

####`apiBase` default 'https://www.nationaalarchief.nl'
Sets the base host url where the JSON backend is located. This is used in
the OTAP deployment configuration.

####`apiUrl` default '/gaf/retrieve/archives/{{archivalInstituteId}}'
The {{archivalInstituteId}} is replaced by the configured institute id. Then
the full endpoint is constructed with the `apiBase` for the ajax requests.

####`docType` default 'fa'
Document type for the results items:
* fa. Finding Aid
* hg. Holdings Guide
* sg. Source Guide


Usage GAFSEARCH
---------------
For searching inside the archives of the supported institutions you can
use the GAFSEARCH plugin. Create an empty div with id #gaf and attach
the plugin to it. This creates a search input and when the user enters
a search term the results are rendered beneath the search input with
facet filters and pagination for the search results.

```javascript
$(document).ready(function(){
  var options = {
    baseUrl: '/search',
    sourceDirectory: '/vendor/gaf/',
    language: 'NL'
  };
  $('#gaf').GAFSEARCH(options);
});
```

### OPTIONS GAFSEARCH
####`baseUrl` default '/search'
The baseUrl is the page url where you implement the GAFSEARCH plugin. This
option value is used for the pagination and results urls. Deep linking is
based on these baseUrl. Example: /search?searchTerm=test&page=2

####`sourceDirectory` default '/'
Sets the base url for loading the template files. Should be set to where you place
these plugin files. Example: /vendor/gaf/

####`mainTemplate` default 'templates/main-template.hbs'
Sets which handlebars template the main template uses that holds the search
input and the html tabs. The results of the tabs are rendered with another
template.

####`language` default 'NL'
Sets the default language of the plugin and translates the applicable labels.
The content of the archives are not translated. You can override the labels,
see `translatedLabels` option.

####`translatedLabels`
A json object containing the languages, the label keys and their translated
value. You can override each language and label by setting the correct value.

```javascript
translatedLabels: {
  NL: {
    searchPlaceHolder: "Zoek",
    search: "Zoeken",
    totalSearchResultsLabel: "%amount% resultaten"
  },
  EN: {
    searchPlaceHolder: "Search",
    search: "Search",
    totalSearchResultsLabel: "%amount% results"
  },
  FY: {
    searchPlaceHolder: "Zoeke",
    search: "Sykje",
    totalSearchResultsLabel: "%amount% resultaten"
  }
}
```

Example override english Search button text and change it to 'Search term'

```javascript
translatedLabels: {
  EN: {
    search: 'Search term'    
  }
}
```

####`appendTitle` default true
This setting indicates whether the title of the page is provided with an extra 
string. This sting indicates which screen the user is in: 
searchpage, resultpage, contentpage or viewerpage.
The strings for these titles can be provided with translatedLabels.

####`resultLanguageOptions` default [{value: 'NL'}, {value: 'EN'}, {value: 'FY'}]
Controls which languages are permitted in the translations.

####`activeTab` default 'gaf'
Controls which tab is active on page load. NOTICE: this option is overrided
by the GET parameter activeTab in the url. This is needed for deeplinking
current active tab results.
Potential values:
- gaf : returns search results from APE archival content
- hub3 : return search results from image repository
- A tab key defined in extraTabs option defined by you.

####`tabVisibility` default []
Controls the visibility AND order of the tab datasources. When this option
is empty it will load the default tabs and extraTabs. NOTICE: make sure
activeTab is defined if you do not have gaf as the default.
See activeTab section for possible keys to use. Example: ['hub3', 'gaf'] 
if you want to show hub3 tab first then gaf. Set activeTab: 'hub3' if you
want the results also start in hub3 when loading the results.

####`extraTabs` default []
Add your own custom tab to the search results page. See the section 
*Integration and events* for integration.

####`desktopBreakPoint` default 768
The plugin is responsive enabled and uses this breakpoint to determine when
the sidebar needs to be moved from a slide in to static column position.
The tabs switch from dropdown to horizontal tab layout based on a number
of variables. One variable is the tabSelectBreakpoint number. If the
screen is smaller than this number the tabs will move into a dropdown layout.
But if there are more tabs than there is room to display them horizontally
on desktop, it will also default to a dropdown layout.

####`beforeRender` default function(viewData){}
You can use this option to add your own function to edit the viewData created
by the api and plugin. This function is called before rendering the html.
If you have your own mainTemplate you can use this function to add your custom
view variables or change existing variables. To check the returned viewData
use the following snippet.

```javascript
beforeRender: function(viewData){
  console.log(viewData);
  console.log(this); // this is scoped to the plugin object.
}
```

####`afterRender` default function($html){}
This option can be used to alter the rendered $html before it is added to the
DOM.

```javascript
beforeRender: function($html){
  console.log($html); // The rendered html containing tabs, input and results.
  console.log(this); // this is scoped to the plugin object.
}
```

Integration
-----------
You can integrate gafsearch within your own CMS with alot of options.
- Set the option `baseUrl` if you want to change the default search page.
- Set the option `sourceDirectory` if you have another directory position.
- You can add more tabs by setting additional configuration in the option
  `extraTabs`. This holds an array with configuration objects that the
  Tab class can use to instantiate more Tabs. Do not forget to properly
  configure the option `tabVisibility` and `activeTab`
  
  ```javascript
  var config = {
    extraTabs: [
      {
        viewerTemplateUrl: '../templates/my-results.hbs',
        apiUrl: '/search/content/my-results',
        key: 'my_tab',
        defaultSearchParams: {
        },
        sortOptions: false,
        translatedLabels: {
          NL: {
            tabLabel: 'Mijn Tab',
            storyCategory: 'Verhaalcategorie',
            author: 'Auteur',
            date: 'Datum'
          },
          EN: {
            tabLabel: 'My Tab',
            storyCategory: 'Story category',
            author: 'Author',
            date: 'Date'
          }
        },
        alterResponseData: function(responseData) {
          if (responseData.documents.length === 0) {
            responseData.noSearchResults = Drupal.t('No results found');
          }
        }
      }
    ],
    activeTab: 'my_tab',
    tabVisibility: ['my_tab', 'gaf']
  };
  
- You can override default configuration of the tabs itself by using
  the option `tabOverrides` and unique tab key, e.g.:
  
  ```javascript
  var config = {
    tabOverrides: {
      gaf: {
        eadDocTemplateUrl: '/templates/my-custom-template.hbs',
        afterRenderEadDocument: function (templateData) {
          // Add my property to the template.
          templateData.my_property = 'my_value';
        }
      }
    }
  };
  ```

## Custom tabs and overrides
The Tab function has multiple option callbacks to change the response
and request data strucutures. For the existing gaf tab you can alter
them in the `tabOverrides` option. If you create your own tab configuration
you just add them there. The options are.
  
####`init` default {}
Run in the constructor of the tab. A tab is constructed once on page load.

####`registerDefaultClickHandlers` default event handlers implemented.
Run in the renderFull of the gafsearch plugin. Usually run once on page load.
This function registers the event handlers for the pagination, sort results
and results per page events. Overriding this tab means you need to implement
your own event handlers for those events. @see registerTabClickHandlers
if you want your extra click handlers.

####`registerTabClickHandlers` default {}
Run in the renderFull of the gafsearch plugin. Usually run once on page load.
Some tabs have extra click functionality not needed in other tabs. Use this
option for your own click handlers.

####`alterRequestData` function (requestData) {}
Run in the getSearchRequestPromise function. This is run on every search request.
With this option you can add default search request parameters or change
existing ones.

####`alterRequestObject` function (requestObject) {}
Run in the getSearchRequestPromise function. This is run on every search request.
Change the constructed request object sent to the backend api.

####`alterResponseData` function (responseData) {}
Run in the getSearchDataPromise. Every time a search has been done but before
the view has been build by Handlebars.
The responseData also contains the translatedLabels for the current language
for use in the template. You can use this option for adding or changing view
variables for use in the template.

####`renderInLine` function () {}
Run in render function of the gafsearch plugin. Runs on every render event.
@see render levels.

####`afterSetPluginParent` function (pluginParent) {}
Run in setPluginParent of the tab. Runs once when instantiating the tab
object with the configuration and also run when using the registerTab
method of the gafsearch plugin.
Use this when you need to instantiate more tab options and you need the
gafsearch pluginParent object. For instance when adding more routes to
the pluginParent.AppRouter.

####`translatedLabels` default {}
Each tab can hold it's own translatedLabels object with keys for each
language. The translatedLabels of the tab are merged with the labels
of the pluginParent in such a way that the tab keys may override the
pluginParent keys. The translatedLabels option holds all the language
keys and variables, but the responseData translatedLabels only hold
the keys and variables for the currently set language.


How references are used in the archive description
--------------------------------------------------
By using specific DIV-containers with class `ead-text-container` 
You can reference EAD-elements in the template `'gaf-ead.hbs'`.
The elements can contain other elements like href's, subcontent, lists and will be rendered to HTML.

There are 3 reference options:

  1 - by **'encodinganalog'**, this is the default search type: `data-searchType='encodinganalog'` is not necessary
  ```html
      <div class="ead-text-container" data-searchFor="3.2.2"></div>
  ```
  2 - by **'label'**, use `data-searchType="label"`
  ```html
      <div class="ead-text-container" data-searchFor="Archiefvormers: " data-searchType="label"></div>
  ```
  3 - by **'head'**, use `data-searchType="head"`
  ```html
      <div class="ead-text-container" data-searchFor="Inhoud" data-searchType="head"></div>
  ```
  
If you want to override the label found in the EAD, you can use `data-replaceLabel="[NEWLABEL]"`
In this example `Periode` is replacing the label-string found in `3.1.3`
```html
<div class="ead-text-container" data-replaceLabel="Periode" data-searchFor="3.1.3"></div>
```


Internationalization (i18n)
---------------------------
The plugin currently has translations for dutch (NL), english (EN) and 
fries (FY). You can add more translations by adding the options to the option
keys `translatedLabels` `resultLanguageOptions` `language`. You can also
edit the labels by changing the `translatedLabels` option. See that option
section for an example.


Updates
-------
* 2.0.4.9    Search inside ead added and refactor of the gaf-ead subfunctions.
* 2.0.3.9    Load left and right ead trees within view.
* 2.0.2.9    Load ead basic content.
* 2.0.1      Add desktop page links and mobile page links for gafsearch, gafarchives and spyridon results.   
* 1.0.0      Initial release.


License
-------
I do not know :) Ask Nationaal Archief.


Authors
-------
* Daniel Karso <daniel@netbuffer.com>
* Arie Nieuwkoop <verwer073@gmail.com>
* Michel Mahieu <mwd@mahieu.nl>
* Jeroen Bijl <jeroen.bijl@clockwork.nl>

2017-2018
