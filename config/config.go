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
	Elastic = ElasticConnection{
		Host:     "https://127.0.0.1:9200",
		CertFile: "config/cert/oipd.pem",
		CertKey:  "config/cert/oipd.key",
		CertRoot: "config/cert/root-ca.pem",
		UseCert:  true,
	}
)

type FlodConnection struct {
	Host     string
	User     string
	Pass     string
	CertFile string
}

type ElasticConnection struct {
	Host     string
	UseCert  bool
	CertFile string
	CertKey  string
	CertRoot string
}

func init() {
	logger.SetOutput(os.Stdout)
}
