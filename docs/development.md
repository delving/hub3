# Developing RAPID

RAPID is a [golang](https://golang.org/) application. In this document we describe the development choices.

## Dependency management

We have decided to use [glide](https://glide.sh/) for vendoring. All pinned dependencies are stored in the `./vendor` directory.

Here are some basic commands to work with glide.

Install the dependencies and revisions listed in the lock file into the vendor directory. If no lock file exists an update is run.

    `$ glide install`


Install the latest dependencies into the vendor directory matching the version resolution information. The complete dependency tree is installed, importing Glide, Godep, GB, and GPM configuration along the way. A lock file is created from the final output.

    `$ glide update`

Add a new dependency to the glide.yaml, install the dependency, and re-resolve the dependency tree. Optionally, put a version after an anchor.

    $ glide get github.com/foo/bar

or 

    $ glide get github.com/foo/bar#^1.2.3
