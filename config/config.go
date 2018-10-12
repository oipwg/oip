package config

import (
	"os"
	"strings"

	"github.com/azer/logger"
	"github.com/micro/go-config"
	"github.com/micro/go-config/reader"
	"github.com/micro/go-config/source/file"
	"github.com/micro/go-config/source/memory"
)

func init() {
	logger.SetOutput(os.Stdout)

	err := config.Load(file.NewSource(file.WithPath("config/config.json")))
	if err != nil {
		log.Error("Unable to load configuration file, using default values")
	}
}

func Get(path string) reader.Value {
	p := strings.Split(path, ".")
	return config.Get(p...)
}

func IsTestnet() bool {
	return config.Get("testnet").Bool(false)
}

func SetTestnet(testnet bool) {
	// ToDo: is this really the best way?
	t := []byte(`{"testnet": true}`)
	f := []byte(`{"testnet": false}`)

	data := t
	if !testnet {
		data = f
	}

	memorySource := memory.NewSource(
		memory.WithData(data),
	)

	config.Load(memorySource)
}
