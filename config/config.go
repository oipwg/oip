package config

import (
	"os"

	"github.com/azer/logger"
)

var (
	MainFlod = FlodConnection{
		"127.0.0.1:8334",
		"user",
		"pass",
	}
)

type FlodConnection struct {
	Host string
	User string
	Pass string
}

func init() {
	logger.SetOutput(os.Stdout)
}
