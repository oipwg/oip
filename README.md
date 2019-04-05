OIP Daemon
===

Process that monitors the FLO Blockchain indexing property formatted OIP messages and provides search and retrieval of
OIP records (aka artifacts).

## Build Instructions

1. Download the repository into $GOPATH/src/oipwg/oip. This _really_ matters to make development easier. This includes
developers external to OIPWG.
2. Install `dep` from https://github.com/golang/dep
3. run `dep ensure -v`
4. `go get -u github.com/gobuffalo/packr/v2/packr2`
5. `cd $GOPATH/src/github.com/oipwg/oip/cmd/oipd && packr2 -v && cd -`
6. run `go build ./cmd/oipd`
7. The executable is `oipd` built in the root directory of the project
8. run tests with `go test -v -race`


## Contacts
Chris Chrysostom, cchrysostom@mediciland.com
