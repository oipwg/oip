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
	"github.com/cloudflare/backoff"
	"github.com/pkg/errors"

	"github.com/oipwg/oip/config"
	"github.com/oipwg/oip/events"
)

var (
	clients []*rpcclient.Client
)

func AddCore(host, user, pass string) error {
	cfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		DisableTLS:   true,
		HTTPPostMode: true,
	}
	c, err := rpcclient.New(cfg, nil)
	clients = append(clients, c)
	return err
}

func WaitForFlod(ctx context.Context, host, user, pass string, tls bool) error {
	attempts := 0
	a := logger.Attrs{"host": host, "attempts": attempts}
	b := backoff.NewWithoutJitter(10*time.Minute, 1*time.Second)
	t := log.Timer()
	defer t.End("WaitForFlod", a)
	for {
		attempts++
		a["attempts"] = attempts
		log.Info("attempting connection to flod", a)
		err := AddFlod(host, user, pass, tls)
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

func AddFlod(host, user, pass string, tls bool) error {
	var certs []byte
	var err error
	if tls {
		certFile := config.GetFilePath("flod.certFile")
		certs, err = ioutil.ReadFile(certFile)
		if err != nil {
			return errors.Wrap(err, "unable to read rpc.cert")
		}
	}

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnFilteredBlockConnected: func(height int32, header *wire.BlockHeader, txns []*floutil.Tx) {
			log.Info("Block connected: %v (%d) %v",
				header.BlockHash(), height, header.Timestamp)
			events.Publish("flo:notify:onFilteredBlockConnected", height, header, txns)
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {
			log.Info("Block disconnected:  %v (%d) %v",
				header.BlockHash(), height, header.Timestamp)
			events.Publish("flo:notify:onFilteredBlockDisconnected", height, header)
		},
		OnTxAcceptedVerbose: func(txDetails *flojson.TxRawResult) {
			log.Info("Incoming TX: %v (Block: %v) floData: %v", txDetails.Txid, txDetails.FloData, txDetails.BlockHash)
			events.Publish("flo:notify:onTxAcceptedVerbose", txDetails)
		},
	}

	cfg := &rpcclient.ConnConfig{
		Host:         host,
		Endpoint:     "ws",
		User:         user,
		Pass:         pass,
		DisableTLS:   !tls,
		Certificates: certs,
	}
	c, err := rpcclient.New(cfg, &ntfnHandlers)
	if err != nil {
		return errors.Wrap(err, "unable to create new rpc client")
	}
	clients = append(clients, c)
	return nil
}

func Disconnect() {
	if len(clients) == 1 {
		clients[0].Disconnect()
	} else {
		for _, c := range clients {
			c.Disconnect()
		}
	}
}

func GetBlockCount() (blockCount int64, err error) {
	err = errors.New("no clients connected")

	if len(clients) == 1 {
		blockCount, err = clients[0].GetBlockCount()
	} else {
		for _, c := range clients {
			blockCount, err = c.GetBlockCount()
			if err == nil {
				return
			}
		}
	}
	return
}

func BeginNotifyBlocks() (err error) {
	err = errors.New("no clients connected")

	if len(clients) == 1 {
		err = clients[0].NotifyBlocks()
	} else {
		for _, c := range clients {
			err = c.NotifyBlocks()
			if err == nil {
				return
			}
		}
	}
	return
}

func BeginNotifyTransactions() (err error) {
	err = errors.New("no clients connected")

	if len(clients) == 1 {
		err = clients[0].NotifyNewTransactions(true)
	} else {
		for _, c := range clients {
			err = c.NotifyNewTransactions(true)
			if err == nil {
				return
			}
		}
	}
	return
}

func GetFirstClient() *rpcclient.Client {
	if len(clients) > 0 {
		return clients[0]
	}
	return nil
}

func GetBlockHash(i int64) (hash *chainhash.Hash, err error) {
	err = errors.New("no clients connected")
	for _, c := range clients {
		hash, err = c.GetBlockHash(i)
		if err == nil {
			return
		}
	}
	return
}

func GetBlockVerboseTx(hash *chainhash.Hash) (br *flojson.GetBlockVerboseResult, err error) {
	err = errors.New("no clients connected")
	if len(clients) == 1 {
		br, err = clients[0].GetBlockVerboseTx(hash)
	} else {
		for _, c := range clients {
			br, err = c.GetBlockVerboseTx(hash)
			if err == nil {
				return
			}
		}
	}
	return
}

func GetTxVerbose(hash *chainhash.Hash) (tr *flojson.TxRawResult, err error) {
	err = errors.New("no clients connected")
	if len(clients) == 1 {
		tr, err = clients[0].GetRawTransactionVerbose(hash)
	} else {
		for _, c := range clients {
			tr, err = c.GetRawTransactionVerbose(hash)
			if err == nil {
				return
			}
		}
	}
	return
}

func CheckAddress(address string) (bool, error) {
	var err error
	if config.IsTestnet() {
		_, err = floutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	} else {
		_, err = floutil.DecodeAddress(address, &chaincfg.MainNetParams)
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
		ok, err = flosig.CheckSignature(address, signature, message, "Florincoin", &chaincfg.TestNet3Params)
	} else {
		ok, err = flosig.CheckSignature(address, signature, message, "Florincoin", &chaincfg.MainNetParams)
	}
	if !ok && err == nil {
		err = errors.New("bad signature")
	}

	return ok, err
}
