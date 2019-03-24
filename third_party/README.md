# Third Party Libs

This package contains source for rendering webcontent. All these files are included in the generated binary as a virtual filesystem using (vfsgen)[github.com/shurcooL/vfsgen] library. Currently, the following libraries are included:

* Yasgui sparql viewer, see (link)[https://about.yasgui.org/].
* Linked Data Fragments viewer, see (link)[http://linkeddatafragments.org/].
* Dutch National Archive Archival explorer proof-of-concept (currently not publicly available).

Each directory of a third party library has a `version.json` file that tracks the following information (yasgui used as example):

```json
{
    "version": "2.7.29",
    "git-sha": "9566563",
    "repo": "https://github.com/OpenTriply/YASGUI",
    "release-date": "17-07-2018"
}

```

TODO: add Delving Hub3 Dashboard viewer
