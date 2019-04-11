package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/azer/logger"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/filters"
	"github.com/oipwg/oip/flo"
	"github.com/oipwg/oip/httpapi"
	_ "github.com/oipwg/oip/modules"
	"github.com/oipwg/oip/sync"
	"github.com/oipwg/oip/version"
	"github.com/spf13/viper"
)

func main() {
	log.Info("\n\n\n\n\n\n")
	log.Info(" OIP Daemon ", logger.Attrs{
		"commitHash": version.GitCommitHash,
		"buildDate":  version.BuildDate,
		"builtBy":    version.BuiltBy,
		"goVersion":  version.GoVersion,
	})

	defer flo.Disconnect()

	rootContext := context.Background()
	rootContext, cancelRoot := context.WithCancel(rootContext)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Error("Received signal %s", sig)
		cancelRoot()
	}()

	tenMinuteCtx, cancel := context.WithTimeout(rootContext, 10*time.Minute)
	defer cancel()

	host := viper.GetString("flod.host")
	user := viper.GetString("flod.user")
	pass := viper.GetString("flod.pass")
	tls := viper.GetBool("flod.tls")

	err := flo.WaitForFlod(tenMinuteCtx, host, user, pass, tls)
	if err != nil {
		log.Error("Unable to connect to Flod", logger.Attrs{"host": host, "err": err})
		shutdown(err)
		return
	}

	apiEnabled := viper.GetBool("oip.api.enabled")
	if apiEnabled {
		log.Info("starting http api")
		go httpapi.Serve()
	}

	count, err := flo.GetBlockCount()
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

	filters.InitViper(rootContext)

	_, err = sync.InitialSync(rootContext, count)
	if err != nil {
		log.Error("Initial sync failed", logger.Attrs{"err": err})
		shutdown(err)
		return
	}

	sync.IsInitialSync = false
	datastore.AutoBulk.BeginTimedCommits(5 * time.Second)

	err = flo.BeginNotifyBlocks()
	if err != nil {
		log.Error("BeginNotifyBlocks failed", logger.Attrs{"err": err})
		shutdown(err)
		return
	}
	err = flo.BeginNotifyTransactions()
	if err != nil {
		log.Error("BeginNotifyTransactions failed", logger.Attrs{"err": err})
		shutdown(err)
		return
	}

	<-rootContext.Done()
	shutdown(nil)
}

func shutdown(err error) {
	log.Error("Shutting down...", err)
}
