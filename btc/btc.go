package btc

import (
	"github.com/bitspill/flod/chaincfg"
	"github.com/bitspill/flosig"
	"github.com/bitspill/floutil"
	"github.com/oipwg/oip/config"
	"github.com/pkg/errors"
)

func CheckAddress(address string) (bool, error) {
	var err error
	if config.IsTestnet() {
		_, err = floutil.DecodeAddress(address, &chaincfg.BtcTestNet3Params)
	} else {
		_, err = floutil.DecodeAddress(address, &chaincfg.BtcMainNetParams)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func CheckSignature(address, signature, message string) (bool, error) {
	var ok bool
	var err error
	if config.IsTestnet() {
		ok, err = flosig.CheckSignature(address, signature, message, "Bitcoin", &chaincfg.BtcTestNet3Params)
	} else {
		ok, err = flosig.CheckSignature(address, signature, message, "Bitcoin", &chaincfg.BtcMainNetParams)
	}
	if !ok && err == nil {
		err = errors.New("bad signature")
	}

	return ok, err
}
