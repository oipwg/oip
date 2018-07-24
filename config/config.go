package config

import (
	"os"
	"path/filepath"

	"github.com/azer/logger"
	"github.com/bitspill/floutil"
)

var (
	flodHomeDir = floutil.AppDataDir("flod", false)
	MainFlod    = FlodConnection{
		Host:     "127.0.0.1:8334",
		User:     "user",
		Pass:     "pass",
		CertFile: filepath.Join(flodHomeDir, "rpc.cert"),
	}
)

type FlodConnection struct {
	Host     string
	User     string
	Pass     string
	CertFile string
}

func init() {
	logger.SetOutput(os.Stdout)
}
