package config

import (
	"os"
	"path/filepath"

	"github.com/azer/logger"
	"github.com/bitspill/floutil"
	"github.com/spf13/viper"
)

var (
	appDir = floutil.AppDataDir("oipd", false)
)

func init() {
	logger.SetOutput(os.Stdout)

	loadDefaults()

	err := os.MkdirAll(filepath.Join(appDir, "certs"), os.ModePerm)
	if err != nil {
		panic(err)
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(appDir)
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		log.Error("error loading config file, utilizing defaults", logger.Attrs{"err": err})
	}
}

func loadDefaults() {
	// Elastic defaults
	viper.SetDefault("elastic.host", "http://127.0.0.1:9200")
	viper.SetDefault("elastic.cert_file", filepath.Join(appDir, "certs/oipd.pem"))
	viper.SetDefault("elastic.cert_key", filepath.Join(appDir, "certs/oipd.key"))
	viper.SetDefault("elastic.cert_root", filepath.Join(appDir, "certs/root-ca.pem"))

	// Flod defaults
	defaultFlodDir := floutil.AppDataDir("flod", false)
	defaultFlodCert := filepath.Join(defaultFlodDir, "rpc.cert")
	viper.SetDefault("flod.certFile", defaultFlodCert)
	viper.SetDefault("flod.host", "127.0.0.1:8334")
	viper.SetDefault("flod.user", "user")
	viper.SetDefault("flod.pass", "pass")

	// HttpApi defaults
	viper.SetDefault("api.listen", "127.0.0.1:1606")
	viper.SetDefault("api.enabled", false)

	// Testnet defaults
	viper.SetDefault("testnet", false)
}

func IsTestnet() bool {
	return viper.GetBool("testnet")
}

func SetTestnet(testnet bool) {
	viper.Set("testnet", testnet)
}
