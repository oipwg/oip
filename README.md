OIP Daemon
===

Process that monitors the FLO Blockchain indexing property formatted OIP messages and provides search and retrieval of
OIP records (aka artifacts).

Prior to running this project, you will need to have GoFlow and Elastisearch also running.

GoFlo - https://github.com/bitspill/flod

ElasticSearch - https://www.elastic.co/downloads/past-releases (get a 6.7 version)

Additionally, ensure your $GOPATH environment variable has been set, and that your $PATH includes $GOPATH/bin.

Lastly, to see output from OIPD, run "export LOG=*". this will output all logging to the console. 

## Build Instructions

1. Download the repository into $GOPATH/src/oipwg/oip. This _really_ matters to make development easier. This includes
developers external to OIPWG.
2. Install `dep` from https://github.com/golang/dep
3. In the oip directory, run `dep ensure -v`
4. `go get -u github.com/gobuffalo/packr/v2/packr2`
5. `cd $GOPATH/src/github.com/oipwg/oip/cmd/oipd && packr2 -v && cd -`
6. run `go build ./cmd/oipd`
7. Modify your ~/.oipd/config.yml to make sure you're running off testnet, and set tls to false 
8. The executable is `oipd` built in the root directory of the project
9. run tests with `go test -v -race`
10. To run OIPd with a profiler, use the cpuprofile and memprofile flags (e.g. -- cpuprofile=oipd_cpu.prof memprofile=oipd_mem.prof) 

## Contacts
Chris Chrysostom, cchrysostom@mediciland.com
