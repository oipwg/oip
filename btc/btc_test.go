package btc

import (
	"testing"

	"github.com/bitspill/oip/config"
)

func TestCheckSignature(t *testing.T) {
	// ToDo: requires BTC chaincfg params
	t.Skip("requires BTC chaincfg params")

	// save setting to restore post-test
	testnet := config.Testnet

	// MainNet
	config.Testnet = false
	adr := "1PVdqQygncV32a5YMWUmfEz2h3CqdHfXJe"
	sig := "G25OicB3g46g9kZ0dGOI8+d9ZTlGrH8yKbCa5Xcd10UHcXZ0NRncgwCsKKGyXkU2+BLy0aq3013a0dTFfWf6mDQ="
	msg := "Bitcoin signed message test"
	valid, err := CheckSignature(adr, sig, msg)
	if err != nil {
		t.Error(err)
	}
	if !valid {
		t.Fail()
	}

	// TestNet
	config.Testnet = true
	// ToDo: add testnet test case

	// restore pre-test setting
	config.Testnet = testnet
}
