package config

import (
	"testing"
)

func TestSetTestnet(t *testing.T) {
	// save pre test value
	testnet := IsTestnet()

	SetTestnet(true)
	if Get("testnet").Bool(false) != true {
		t.Error("expected true, received false")
	}

	SetTestnet(false)
	if Get("testnet").Bool(true) != false {
		t.Error("expected false, received true")
	}

	// restore
	SetTestnet(testnet)
}
