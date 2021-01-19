package publishing

import (
	"strconv"
	"testing"

	"github.com/bitspill/flod/chaincfg"
	"github.com/bitspill/flod/flojson"
	"github.com/bitspill/floutil"
)

func TestBuildAncestralTrees(t *testing.T) {
	addr, err := floutil.DecodeAddress(floAddress, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
	wif, err := floutil.DecodeWIF(floWifKey)
	if err != nil {
		t.Fatal(err)
	}

	pub, err := NewAddress(nil, addr, wif, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}

	pubAddr, ok := pub.(*Address)
	if !ok {
		t.Fatal("improper publisher type")
	}

	utxo := map[string]*fastUtxo{
		"0bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2.0": {
			Conf:     0,
			Hash:     strPointer("0bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2"),
			Index:    0,
			PkScript: strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:    0.0001,
			VinPrevOut: []flojson.VinPrevOut{
				{Txid: "7ddd14c18cf8b3a8cfa97fe874b5b9c3cb7cff1532e7d6569512f11323133757", Vout: 0},
			},
		},
		"1bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2.0": {
			Conf:     0,
			Hash:     strPointer("1bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2"),
			Index:    0,
			PkScript: strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:    0.0001,
			VinPrevOut: []flojson.VinPrevOut{
				{Txid: "0bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2", Vout: 0},
			},
		},
		"2bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2.0": {
			Conf:     0,
			Hash:     strPointer("2bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2"),
			Index:    0,
			PkScript: strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:    0.0001,
			VinPrevOut: []flojson.VinPrevOut{
				{Txid: "1bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2", Vout: 0},
			},
		},
		"3bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2.0": {
			Conf:     0,
			Hash:     strPointer("3bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2"),
			Index:    0,
			PkScript: strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:    0.0001,
			VinPrevOut: []flojson.VinPrevOut{
				{Txid: "2bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2", Vout: 0},
			},
		},
		"51d12a2d815176d30a88022cd9e921a33157c43d5d18c4be1c0cffc57250c898.0": {
			Conf:       7343,
			Hash:       strPointer("51d12a2d815176d30a88022cd9e921a33157c43d5d18c4be1c0cffc57250c898"),
			Index:      0,
			PkScript:   strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:      18,
			VinPrevOut: nil,
		},
		"7a34b32b3f7db2825f7a95ac51e59f0d4e3392787969f5574b6989337f4999d6.0": {
			Conf:       12367,
			Hash:       strPointer("7a34b32b3f7db2825f7a95ac51e59f0d4e3392787969f5574b6989337f4999d6"),
			Index:      0,
			PkScript:   strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:      0.0379,
			VinPrevOut: nil,
		},
		"0ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3.0": {
			Conf:     0,
			Hash:     strPointer("0ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3"),
			Index:    0,
			PkScript: strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:    0.0387,
			VinPrevOut: []flojson.VinPrevOut{
				{Txid: "d5455ac27c1d5b03d22fc60f109558c1ded7751245afddfe17e6242a9d9548d3", Vout: 0},
			},
		},
		"1ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3.0": {
			Conf:     0,
			Hash:     strPointer("1ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3"),
			Index:    0,
			PkScript: strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:    0.0386,
			VinPrevOut: []flojson.VinPrevOut{
				{Txid: "0ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3", Vout: 0},
			},
		},
		"2ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3.0": {
			Conf:     0,
			Hash:     strPointer("2ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3"),
			Index:    0,
			PkScript: strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:    0.0385,
			VinPrevOut: []flojson.VinPrevOut{
				{Txid: "1ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3", Vout: 0},
			},
		},
		"3ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3.0": {
			Conf:     0,
			Hash:     strPointer("3ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3"),
			Index:    0,
			PkScript: strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:    0.0384,
			VinPrevOut: []flojson.VinPrevOut{
				{Txid: "2ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3", Vout: 0},
			},
		},
		"4ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3.0": {
			Conf:     0,
			Hash:     strPointer("4ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3"),
			Index:    0,
			PkScript: strPointer("76a914f5aac13500ad4698b5bea4476084e7e36352933088ac"),
			Value:    0.0383,
			VinPrevOut: []flojson.VinPrevOut{
				{Txid: "3ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3", Vout: 0},
			},
		},
	}

	// repeat many times to account for random map iteration order
	// ensuring results don't depend upon iteration order
	for i := 0; i < 100; i++ {
		pubAddr.resetUtxo()
		err = pubAddr.buildAncestralTrees(utxo)
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Println("iteration", i, "utxo", len(pubAddr.utxo))
		// for key := range pubAddr.unconfirmed {
		// 	fmt.Println("unconfirmed", key, len(pubAddr.unconfirmed[key]))
		// 	for k, v := range pubAddr.unconfirmed[key] {
		// 		fmt.Printf("  %d %s\n", k, v.Hash.String())
		// 	}
		// }

		if 5 != len(pubAddr.unconfirmed["0ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3.0"]) {
			t.Fatal("unexpected chain 0ffc length")
		} else {
			for k, v := range pubAddr.unconfirmed["0ffc7e36ba5ef96d86725c7652870c961891b1e22231de4b41375a5bda7adba3.0"] {
				target := strconv.Itoa(k)[0]
				got := v.Hash.String()[0]
				if target != got {
					t.Fatal("0ffc out of order")
				}
			}
		}
		if 4 != len(pubAddr.unconfirmed["0bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2.0"]) {
			t.Fatal("unexpected chain 0bd6 length ")
		} else {
			for k, v := range pubAddr.unconfirmed["0bd6a72a4c110ee28ce7eb905d3e8f6ced3f44914d9ae0d732435682eb6268d2.0"] {
				target := strconv.Itoa(k)[0]
				got := v.Hash.String()[0]
				if target != got {
					t.Fatal("0bd6 out of order")
				}
			}
		}
	}
}

func strPointer(str string) *string {
	return &str
}
