package config

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/azer/logger"
	"github.com/bitspill/floutil"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/viper"
)

var (
	appDir    = floutil.AppDataDir("oipd", false)
	configBox = packr.New("defaults", "./defaults")
)

func init() {
	logger.SetOutput(os.Stdout)

	loadDefaults()

	err := os.MkdirAll(filepath.Join(appDir, "certs"), os.ModePerm)
	if err != nil {
		panic(err)
	}

	b, err := configBox.Find("config.example.yml")
	if err != nil {
		panic(err)
	}
	err = viper.ReadConfig(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	_, err = os.Stat(filepath.Join(appDir, "config.yml"))
	if os.IsNotExist(err) {
		log.Info("config.yml not found, writing default config file")
		err = ioutil.WriteFile(filepath.Join(appDir, "config.yml"), b, os.ModePerm)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
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
	viper.SetDefault("elastic.useCert", false)
	viper.SetDefault("elastic.certFile", filepath.Join(appDir, "certs/oipd.pem"))
	viper.SetDefault("elastic.certKey", filepath.Join(appDir, "certs/oipd.key"))
	viper.SetDefault("elastic.certRoot", filepath.Join(appDir, "certs/root-ca.pem"))

	// Flod defaults
	defaultFlodDir := floutil.AppDataDir("flod", false)
	defaultFlodCert := filepath.Join(defaultFlodDir, "rpc.cert")
	viper.SetDefault("flod.certFile", defaultFlodCert)
	viper.SetDefault("flod.host", "127.0.0.1:8334")
	viper.SetDefault("flod.user", "user")
	viper.SetDefault("flod.pass", "pass")

	// Testnet defaults
	viper.SetDefault("oip.network", "mainnet")

	// HttpApi defaults
	viper.SetDefault("oip.api.listen", "127.0.0.1:1606")
	viper.SetDefault("oip.api.enabled", false)
}

func IsTestnet() bool {
	return viper.GetString("oip.network") != "mainnet"
}

func SetTestnet(testnet bool) {
	n := "mainnet"
	if testnet {
		n = "testnet"
	}
	viper.Set("oip.network", n)
}

func GetFilePath(key string) string {
	v := viper.GetString(key)
	if filepath.IsAbs(v) {
		return v
	}
	return filepath.Join(appDir, v)
}
