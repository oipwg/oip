package oip042

import (
	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/json-iterator/go"
)

const oip042ArtifactIndex = `oip042_artifact`
const oip042PublisherIndex = `oip042_ publisher`
const oip042InfluencerIndex = `oip042_influencer`
const oip042PlatformIndex = `oip042_platform`
const oip042AutominerIndex = `oip042_autominer`
const oip042PoolIndex = `oip042_pool`
const oip042EditIndex = `oip042_edit`
const oip042TransferIndex = `oip042_transfer`
const oip042DeactivateIndex = `oip042_deactivate`

func init() {
	log.Info("init oip042 json")
	events.Bus.SubscribeAsync("modules:oip042:json", on42Json, false)

	datastore.RegisterMapping(oip042ArtifactIndex, publishOip042ArtifactMapping)
}

func on42Json(message jsoniter.RawMessage, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42Json", logger.Attrs{"txid": tx.Transaction.Txid})
	if !jsoniter.Valid(message) {
		log.Info("invalid json %s", tx.Transaction.Txid)
		return
	}

	publish := jsoniter.Get(message, "publish")
	err := publish.LastError()
	if err == nil {
		on42JsonPublish(publish, tx)
		return
	}
	register := jsoniter.Get(message, "register")
	err = publish.LastError()
	if err == nil {
		on42JsonRegister(register, tx)
		return
	}
	edit := jsoniter.Get(message, "edit")
	err = publish.LastError()
	if err == nil {
		on42JsonEdit(edit, tx)
		return
	}
	transfer := jsoniter.Get(message, "transfer")
	err = publish.LastError()
	if err == nil {
		on42JsonTransfer(transfer, tx)
		return
	}
	deactivate := jsoniter.Get(message, "deactivate")
	err = publish.LastError()
	if err == nil {
		on42JsonDeactivate(deactivate, tx)
		return
	}

	log.Error("no publisher/register message %s", tx.Transaction.Txid)
}

func on42JsonPublish(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonPublish", logger.Attrs{"txid": tx.Transaction.Txid})

	// artifact
	// publisher -- N/A
	// influencer -- N/A
	// platform -- N/A
	// pool -- N/A
	// miner -- N/A
	pub := any.Get("artifact")
	err := pub.LastError()
	if err == nil {
		on42JsonPublishArtifact(pub, tx)
		return
	}

	log.Error("no publish %s", tx.Transaction.Txid)
}

func on42JsonRegister(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonRegister", logger.Attrs{"txid": tx.Transaction.Txid})

	// artifact -- N/A
	// publisher
	// influencer
	// platform
	// pool
	// miner
	pub := any.Get("pub")
	err := pub.LastError()
	if err == nil {
		on42JsonRegisterPub(pub, tx)
		return
	}
	inf := any.Get("influencer")
	err = inf.LastError()
	if err == nil {
		on42JsonRegisterInfluencer(inf, tx)
		return
	}
	plat := any.Get("platform")
	err = plat.LastError()
	if err == nil {
		on42JsonRegisterPlatform(plat, tx)
		return
	}
	pool := any.Get("pool")
	err = pool.LastError()
	if err == nil {
		on42JsonRegisterPool(pool, tx)
		return
	}
	miner := any.Get("autominer")
	err = miner.LastError()
	if err == nil {
		on42JsonRegisterAutominer(miner, tx)
		return
	}

	log.Error("no supported register %s", tx.Transaction.Txid)
}

func on42JsonEdit(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonEdit", logger.Attrs{"txid": tx.Transaction.Txid})

	// artifact
	// publisher
	// influencer
	// platform
	// pool
	// miner
	art := any.Get("artifact")
	err := art.LastError()
	if err == nil {
		on42JsonEditArtifact(art, tx)
		return
	}
	pub := any.Get("pub")
	err = pub.LastError()
	if err == nil {
		on42JsonEditPub(pub, tx)
		return
	}
	inf := any.Get("influencer")
	err = inf.LastError()
	if err == nil {
		on42JsonEditInfluencer(inf, tx)
		return
	}
	plat := any.Get("platform")
	err = plat.LastError()
	if err == nil {
		on42JsonEditPlatform(plat, tx)
		return
	}
	pool := any.Get("pool")
	err = pool.LastError()
	if err == nil {
		on42JsonEditPool(pool, tx)
		return
	}
	miner := any.Get("autominer")
	err = miner.LastError()
	if err == nil {
		on42JsonEditAutominer(miner, tx)
		return
	}

	log.Error("no supported edit %s", tx.Transaction.Txid)
}

func on42JsonTransfer(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonTransfer", logger.Attrs{"txid": tx.Transaction.Txid})

	// artifact
	// publisher
	// influencer
	// platform
	// pool
	// miner
	art := any.Get("artifact")
	err := art.LastError()
	if err == nil {
		on42JsonTransferArtifact(art, tx)
		return
	}
	pub := any.Get("pub")
	err = pub.LastError()
	if err == nil {
		on42JsonTransferPub(pub, tx)
		return
	}
	inf := any.Get("influencer")
	err = inf.LastError()
	if err == nil {
		on42JsonTransferInfluencer(inf, tx)
		return
	}
	plat := any.Get("platform")
	err = plat.LastError()
	if err == nil {
		on42JsonTransferPlatform(plat, tx)
		return
	}
	pool := any.Get("pool")
	err = pool.LastError()
	if err == nil {
		on42JsonTransferPool(pool, tx)
		return
	}
	miner := any.Get("autominer")
	err = miner.LastError()
	if err == nil {
		on42JsonTransferAutominer(miner, tx)
		return
	}

	log.Error("no supported transfer %s", tx.Transaction.Txid)
}

func on42JsonDeactivate(any jsoniter.Any, tx datastore.TransactionData) {
	t := log.Timer()
	defer t.End("on42JsonDeactivate", logger.Attrs{"txid": tx.Transaction.Txid})

	// artifact
	// publisher
	// influencer
	// platform
	// pool
	// miner
	art := any.Get("artifact")
	err := art.LastError()
	if err == nil {
		on42JsonDeactivateArtifact(art, tx)
		return
	}
	pub := any.Get("pub")
	err = pub.LastError()
	if err == nil {
		on42JsonDeactivatePub(pub, tx)
		return
	}
	inf := any.Get("influencer")
	err = inf.LastError()
	if err == nil {
		on42JsonDeactivateInfluencer(inf, tx)
		return
	}
	plat := any.Get("platform")
	err = plat.LastError()
	if err == nil {
		on42JsonDeactivatePlatform(plat, tx)
		return
	}
	pool := any.Get("pool")
	err = pool.LastError()
	if err == nil {
		on42JsonDeactivatePool(pool, tx)
		return
	}
	miner := any.Get("autominer")
	err = miner.LastError()
	if err == nil {
		on42JsonDeactivateAutominer(miner, tx)
		return
	}

	log.Error("no supported deactivate %s", tx.Transaction.Txid)
}
