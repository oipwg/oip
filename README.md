OIP Daemon
===

Process that monitors the FLO Blockchain indexing property formatted OIP messages and provides search and retrieval of
OIP records (aka artifacts).

## Build Instructions

1. Download the repository into $GOPATH/src/oipwg/oip. This _really_ matters to make development easier. This includes
developers external to OIPWG.
2. Install `dep` from https://github.com/golang/dep
3. run `dep ensure -v`
4. run `go build ./cmd/oipd`
5. The executable is `oipd` built in the root directory of the project
6. run tests with `go test -v -race`


## Contacts
Chris Chrysostom, cchrysostom@mediciland.com
