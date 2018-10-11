package flo

import (
	"testing"

	"github.com/bitspill/oip/config"
)

func TestCheckSignature(t *testing.T) {
	// save setting to restore post-test
	testnet := config.Testnet

	// MainNet
	config.Testnet = false
	adr := "FDxa2dUXPw592svsebdHfGRHxB46DKWVUy"
	sig := "IMjnGVBNW4kvoSITwijwYkrguszkyMQ08TBNu9wvRiVZB3f+L8Me1gkkK30LT9EO2xyMj0lFHORkSi/zM3cOTF0="
	msg := "Flo signed message test"
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
