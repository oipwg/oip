package main

import (
	"context"
	"time"

	"github.com/azer/logger"
	"github.com/bitspill/oip/config"
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
		"goVersion":  version.GoVersion,
	})

	FloRPC = flo.RPC{}
	defer FloRPC.Disconnect()

	rootContext := context.Background()

	ctx, cancel := context.WithTimeout(rootContext, 10*time.Minute)
	defer cancel()
	err := FloRPC.WaitForFlod(ctx, config.MainFlod.Host, config.MainFlod.User, config.MainFlod.Pass)
	if err != nil {
		log.Error("Unable to connect to Flod after 10 minutes", logger.Attrs{"host": config.MainFlod.Host, "err": err})
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
