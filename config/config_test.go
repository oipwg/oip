package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestSetTestnet(t *testing.T) {
	// save pre test value
	testnet := IsTestnet()

	SetTestnet(true)
	if viper.GetBool("testnet") != true {
		t.Error("expected true, received false")
	}

	SetTestnet(false)
	if viper.GetBool("testnet") != false {
		t.Error("expected false, received true")
	}

	// restore
	SetTestnet(testnet)
}
