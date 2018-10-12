package btc

import (
	"github.com/pkg/errors"
)

func CheckAddress(address string) (bool, error) {
	// ToDo: need BTC chaincfg params
	return false, errors.New("Bitcoin address validation not implemented")

	/*
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
	*/
}

func CheckSignature(address, signature, message string) (bool, error) {
	// ToDo: need BTC chaincfg params
	return false, errors.New("Bitcoin signature validation not implemented")

	/*
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
	*/
}
