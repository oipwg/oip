package btc

import (
	"testing"

	"github.com/oipwg/oip/config"
)

func TestCheckSignature(t *testing.T) {
	// save setting to restore post-test
	testnet := config.IsTestnet()

	// MainNet
	config.SetTestnet(false)
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
	config.SetTestnet(true)
	// ToDo: add testnet test case

	// restore pre-test setting
	config.SetTestnet(testnet)
}
