package flo

import (
	"io/ioutil"
	"path/filepath"

	"github.com/bitspill/flod/chaincfg"
	"github.com/bitspill/flod/chaincfg/chainhash"
	"github.com/bitspill/flod/flojson"
	"github.com/bitspill/flod/rpcclient"
	"github.com/bitspill/flosig"
	"github.com/bitspill/floutil"
	"github.com/pkg/errors"
)

type RPC struct {
	clients []*rpcclient.Client
}

func (f *RPC) AddCore(host string, user string, pass string) error {
	cfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	c, err := rpcclient.New(cfg, nil)
	f.clients = append(f.clients, c)
	return err
}

func (f *RPC) AddFlod(host string, user string, pass string) error {
	// Connect to flod RPC server using websockets.
	// ToDo: configure flod rpc.cert location
	flodHomeDir := floutil.AppDataDir("flod", false)
	certs, err := ioutil.ReadFile(filepath.Join(flodHomeDir, "rpc.cert"))
	if err != nil {
		return errors.Wrap(err, "unable to read rpc.cert")
	}

	cfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		Endpoint:     "ws",
		Certificates: certs,
	}
	c, err := rpcclient.New(cfg, nil)
	if err != nil {
		return errors.Wrap(err, "unable to create new rpc client")
	}
	f.clients = append(f.clients, c)
	return nil
}

func (f *RPC) Disconnect() {
	if len(f.clients) == 1 {
		f.clients[0].Disconnect()
	} else {
		for _, c := range f.clients {
			c.Disconnect()
		}
	}
}

func (f *RPC) GetBlockCount() (blockCount int64, err error) {
	err = errors.New("no clients connected")

	if len(f.clients) == 1 {
		blockCount, err = f.clients[0].GetBlockCount()
	} else {
		for _, c := range f.clients {
			blockCount, err = c.GetBlockCount()
			if err == nil {
				return
			}
		}
	}
	return
}

func (f *RPC) GetFirstClient() *rpcclient.Client {
	if len(f.clients) > 0 {
		return f.clients[0]
	}
	return nil
}

func (f *RPC) GetBlockHash(i int64) (hash *chainhash.Hash, err error) {
	err = errors.New("no clients connected")
	for _, c := range f.clients {
		hash, err = c.GetBlockHash(i)
		if err == nil {
			return
		}
	}
	return
}

func (f *RPC) GetBlockVerboseTx(hash *chainhash.Hash) (br *flojson.GetBlockVerboseResult, err error) {
	err = errors.New("no clients connected")
	if len(f.clients) == 1 {
		br, err = f.clients[0].GetBlockVerboseTx(hash)
	} else {
		for _, c := range f.clients {
			br, err = c.GetBlockVerboseTx(hash)
			if err == nil {
				return
			}
		}
	}
	return
}

func (f *RPC) GetTxVerbose(hash *chainhash.Hash) (tr *flojson.TxRawResult, err error) {
	err = errors.New("no clients connected")
	if len(f.clients) == 1 {
		tr, err = f.clients[0].GetRawTransactionVerbose(hash)
	} else {
		for _, c := range f.clients {
			tr, err = c.GetRawTransactionVerbose(hash)
			if err == nil {
				return
			}
		}
	}
	return
}

func CheckAddress(address string, testnet bool) (bool, error) {
	var err error
	if testnet {
		_, err = floutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	} else {
		_, err = floutil.DecodeAddress(address, &chaincfg.MainNetParams)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func CheckSignature(address string, signature string, message string, testnet bool) (bool, error) {
	var ok bool
	var err error
	if testnet {
		ok, err = flosig.CheckSignature(address, signature, message, "Florincoin", &chaincfg.TestNet3Params)
	}
	ok, err = flosig.CheckSignature(address, signature, message, "Florincoin", &chaincfg.MainNetParams)

	if !ok && err == nil {
		err = errors.New("bad signature")
	}

	return ok, err
}
