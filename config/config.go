package config

import (
	"os"
	"path/filepath"

	"github.com/azer/logger"
	"github.com/bitspill/floutil"
	"github.com/json-iterator/go"
	"io/ioutil"
)

// ToDo: substitute a proper configuration management system rather than stacks of if statements

var (
	flodHomeDir = floutil.AppDataDir("flod", false)
	MainFlod    = FlodConnection{
		Host:     "127.0.0.1:8334",
		User:     "user",
		Pass:     "pass",
		CertFile: filepath.Join(flodHomeDir, "rpc.cert"),
	}
	Elastic = ElasticConnection{
		Host:     "http://127.0.0.1:9200",
		CertFile: "config/cert/oipd.pem",
		CertKey:  "config/cert/oipd.key",
		CertRoot: "config/cert/root-ca.pem",
		UseCert:  false,
	}
)

type cfgFile struct {
	MainFlod *FlodConnection    `json:"main_flod"`
	Elastic  *ElasticConnection `json:"elastic"`
}

type FlodConnection struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	CertFile string `json:"cert_file"`
}

type ElasticConnection struct {
	Host     string `json:"host"`
	UseCert  bool   `json:"use_cert"`
	CertFile string `json:"cert_file"`
	CertKey  string `json:"cert_key"`
	CertRoot string `json:"cert_root"`
}

func init() {
	logger.SetOutput(os.Stdout)

	b, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		log.Error("Unable to read config file, using defaults", logger.Attrs{"err": err})
		return
	}

	var cfg cfgFile
	err = jsoniter.Unmarshal(b, &cfg)
	if err != nil {
		log.Error("Unable to read config file, using defaults", logger.Attrs{"err": err})
		return
	}

	if cfg.Elastic != nil {
		if cfg.Elastic.UseCert {
			Elastic.UseCert = cfg.Elastic.UseCert
		}
		if len(cfg.Elastic.CertFile) > 0 {
			Elastic.CertFile = cfg.Elastic.CertFile
		}
		if len(cfg.Elastic.CertKey) > 0 {
			Elastic.CertKey = cfg.Elastic.CertKey
		}
		if len(cfg.Elastic.CertRoot) > 0 {
			Elastic.CertRoot = cfg.Elastic.CertRoot
		}
		if len(cfg.Elastic.Host) > 0 {
			Elastic.Host = cfg.Elastic.Host
		}
	}

	if cfg.MainFlod != nil {
		if len(cfg.MainFlod.Host) > 0 {
			MainFlod.Host = cfg.MainFlod.Host
		}
		if len(cfg.MainFlod.User) > 0 {
			MainFlod.User = cfg.MainFlod.User
		}
		if len(cfg.MainFlod.Pass) > 0 {
			MainFlod.Pass = cfg.MainFlod.Pass
		}
		if len(cfg.MainFlod.CertFile) > 0 {
			MainFlod.CertFile = cfg.MainFlod.CertFile
		}
	}
}
