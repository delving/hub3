# Developing RAPID

RAPID is a [golang](https://golang.org/) application. In this document we describe the development choices.

## Dependency management

We have decided to use [glide](https://glide.sh/) for vendoring. All pinned dependencies are stored in the `./vendor` directory. 
