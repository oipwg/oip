package main

import (
	"context"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/flo"
	_ "github.com/bitspill/oip/modules"
	"github.com/bitspill/oip/version"
	"github.com/davecgh/go-spew/spew"
)

var FloRPC flo.RPC

func main() {
	log.Info("\n\n\n\n\n\n")
	log.Info(" OIP Daemon ", logger.Attrs{
		"commitHash": version.GitCommitHash,
		"buildDate":  version.BuildDate,
		"builtBy":    version.BuiltBy,
	})

	FloRPC = flo.RPC{}
	defer FloRPC.Disconnect()

	err := FloRPC.AddFlod("127.0.0.1:8334", "user", "pass")
	if err != nil {
		panic(err)
	}

	count, err := FloRPC.GetBlockCount()
	if err != nil {
		panic(err)
	}

	log.Info("FLO Block Count: %d", count)

	datastore.Setup(context.TODO())

	lb, err := InitialSync(context.TODO(), count)
	if err != nil {
		panic(err)
	}

	if false {
		spew.Dump(lb)
	}
}
