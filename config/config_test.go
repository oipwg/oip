package config

import (
	"testing"
)

func TestSetTestnet(t *testing.T) {
	// save pre test value
	testnet := IsTestnet()

	SetTestnet(true)
	if !(IsTestnet() == true) {
		t.Error("expected true, received false")
	}

	SetTestnet(false)
	if !(IsTestnet() == false) {
		t.Error("expected false, received true")
	}

	// restore
	SetTestnet(testnet)
}
