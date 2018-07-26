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
	"os"
	"os/signal"
	"syscall"
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
	rootContext, cancelRoot := context.WithCancel(rootContext)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Error("Received signal %s", sig)
		cancelRoot()
	}()

	ctx, cancel := context.WithTimeout(rootContext, 10*time.Minute)
	defer cancel()
	err := FloRPC.WaitForFlod(ctx, config.MainFlod.Host, config.MainFlod.User, config.MainFlod.Pass)
	if err != nil {
		log.Error("Unable to connect to Flod after 10 minutes", logger.Attrs{"host": config.MainFlod.Host, "err": err})
		shutdown(err)
		return
	}

	count, err := FloRPC.GetBlockCount()
	if err != nil {
		log.Error("GetBlockCount failed", logger.Attrs{"err": err})
		shutdown(err)
		return
	}

	log.Info("FLO Block Count: %d", count)

	err = datastore.Setup(rootContext)
	if err != nil {
		log.Error("datastore setup failed", logger.Attrs{"err": err})
		shutdown(err)
		return
	}

	lb, err := InitialSync(rootContext, count)
	if err != nil {
		log.Error("Initial sync failed", logger.Attrs{"err": err})
		shutdown(err)
		return
	}

	datastore.AutoBulk.BeginTimedCommits(5 * time.Second)

	err = FloRPC.BeginNotifyBlocks()
	if err != nil {
		log.Error("BeginNotifyBlocks failed", logger.Attrs{"err": err})
		shutdown(err)
		return
	}
	err = FloRPC.BeginNotifyTransactions()
	if err != nil {
		log.Error("BeginNotifyTransactions failed", logger.Attrs{"err": err})
		shutdown(err)
		return
	}

	if false {
		spew.Dump(lb)
	}

	<-rootContext.Done()
	shutdown(nil)
	return
}

func shutdown(err error) {
	log.Error("Shutting down...")
}
