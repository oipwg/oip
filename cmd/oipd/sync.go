package main

import (
	"github.com/bitspill/flod/wire"
	"github.com/bitspill/floutil"
	"github.com/bitspill/oip/events"
)

func init() {
	events.Bus.SubscribeAsync("flo:notify:onFilteredBlockConnected", onFilteredBlockConnected, true)
}

func onFilteredBlockConnected(height int32, header *wire.BlockHeader, txns []*floutil.Tx) {

}
