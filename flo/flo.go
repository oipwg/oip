package flo

import (
	"context"
	"io/ioutil"
	"net"
	"time"

	"github.com/azer/logger"
	"github.com/bitspill/flod/chaincfg"
	"github.com/bitspill/flod/chaincfg/chainhash"
	"github.com/bitspill/flod/flojson"
	"github.com/bitspill/flod/rpcclient"
	"github.com/bitspill/flod/wire"
	"github.com/bitspill/flosig"
	"github.com/bitspill/floutil"
	"github.com/bitspill/oip/config"
	"github.com/bitspill/oip/events"
	"github.com/cloudflare/backoff"
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

func (f *RPC) WaitForFlod(ctx context.Context, host string, user string, pass string) error {
	attempts := 0
	a := logger.Attrs{"host": host, "attempts": attempts}
	b := backoff.NewWithoutJitter(10*time.Minute, 1*time.Second)
	t := log.Timer()
	defer t.End("WaitForFlod", a)
	for {
		attempts++
		a["attempts"] = attempts
		log.Info("attempting connection to flod", a)
		err := f.AddFlod(host, user, pass)
		if err != nil {
			a["err"] = err
			log.Error("unable to connect to flod", a)
			delete(a, "err")
			c := errors.Cause(err)
			if _, ok := c.(*net.OpError); !ok {
				// not a network error, something else is wrong
				return err
			}
			// it's a network error, delay and retry
			d := b.Duration()
			a["delay"] = d
			log.Info("delaying connection to flod retry", a)
			delete(a, "delay")
			select {
			case <-ctx.Done():
				a["err"] = ctx.Err()
				log.Error("context timeout/cancelled", a)
				return ctx.Err()
			case <-time.After(d):
				// loop around for another try
			}
		} else {
			break
		}
	}
	return nil
}

func (f *RPC) AddFlod(host string, user string, pass string) error {
	// Connect to flod RPC server using websockets.
	certs, err := ioutil.ReadFile(config.MainFlod.CertFile)
	if err != nil {
		return errors.Wrap(err, "unable to read rpc.cert")
	}

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnFilteredBlockConnected: func(height int32, header *wire.BlockHeader, txns []*floutil.Tx) {
			log.Info("Block connected: %v (%d) %v",
				header.BlockHash(), height, header.Timestamp)
			events.Bus.Publish("flo:notify:onFilteredBlockConnected", height, header, txns)
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {
			log.Info("Block disconnected: %v (%d) %v",
				header.BlockHash(), height, header.Timestamp)
			events.Bus.Publish("flo:notify:onFilteredBlockDisconnected", height, header)
		},
		OnTxAcceptedVerbose: func(txDetails *flojson.TxRawResult) {
			log.Info("New tx", logger.Attrs{"txid": txDetails.Txid,
				"floData": txDetails.FloData, "blockHash": txDetails.BlockHash})
			events.Bus.Publish("flo:notify:onTxAcceptedVerbose", txDetails)
		},
	}

	cfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		Endpoint:     "ws",
		Certificates: certs,
	}
	c, err := rpcclient.New(cfg, &ntfnHandlers)
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

func (f *RPC) BeginNotifyBlocks() (err error) {
	err = errors.New("no clients connected")

	if len(f.clients) == 1 {
		err = f.clients[0].NotifyBlocks()
	} else {
		for _, c := range f.clients {
			err = c.NotifyBlocks()
			if err == nil {
				return
			}
		}
	}
	return
}

func (f *RPC) BeginNotifyTransactions() (err error) {
	err = errors.New("no clients connected")

	if len(f.clients) == 1 {
		err = f.clients[0].NotifyNewTransactions(true)
	} else {
		for _, c := range f.clients {
			err = c.NotifyNewTransactions(true)
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
