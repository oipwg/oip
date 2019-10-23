package publishing

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"

	"github.com/bitspill/flod/wire"
	"github.com/bitspill/flosig"
	"github.com/bitspill/floutil"
	"github.com/golang/protobuf/proto"

	"github.com/oipwg/oip/modules/oip"
)

const baseHeader0 = 380
const baseHeaderX = 414
const baseDataSize0 = 660
const baseDataSizeX = /* experiment */ 626

type MultiPart struct {
	pub       Publisher
	addr      floutil.Address
	addrBytes []byte
	wif       floutil.WIF
	floData   []byte
	partCount uint32
	part0Tx   *wire.MsgTx
	part0TxId string
	parts     []pendingPart
	consumed  []byte
	remaining []byte
}

type pendingPart struct {
	mp      *oip.MultiPart
	encoded []byte
}

func NewMultiPart(pub Publisher, signingAddress floutil.Address, wif floutil.WIF) *MultiPart {
	return &MultiPart{
		pub:       pub,
		addr:      signingAddress,
		addrBytes: []byte(signingAddress.EncodeAddress()),
		wif:       wif,
	}
}

func (mp *MultiPart) SetFloData(floData []byte) {
	mp.floData = floData
	mp.partCount = uint32(basePartCount(len(floData)))
	mp.resetState()
}

func (mp *MultiPart) resetState() {
	mp.part0Tx = nil
	mp.parts = nil
	mp.consumed = nil
	mp.remaining = mp.floData
}

func (mp *MultiPart) Build() (*wire.MsgTx, [][]byte, error) {
	lastCount := mp.partCount
buildIt:
	mp.resetState()

	err := mp.genPart0()
	if err != nil {
		return nil, nil, err
	}

	i := uint32(1)
	for len(mp.remaining) > 0 {
		_, err := mp.genPart(i)
		if err != nil {
			return nil, nil, err
		}
		i++
	}

	actual := uint32(len(mp.parts))
	if len(mp.remaining) != 0 {
		mp.partCount++
		lastCount = actual
		goto buildIt
	}

	if actual != mp.partCount {
		// ToDo: don't throw away all previous sizing effort
		//  attempt to only change partCount value first
		//  before starting from scratch
		mp.partCount = actual
		if actual != lastCount {
			lastCount = actual
			goto buildIt
		}
	}

	var serParts [][]byte
	for _, p := range mp.parts {
		serParts = append(serParts, p.encoded)
	}

	return mp.part0Tx, serParts, nil
}

func basePartCount(dataLen int) int {
	return (dataLen-baseDataSize0)/baseDataSizeX + 2
}

func (mp *MultiPart) genPart0() error {
	g64, err := mp.genPart(0)
	if err != nil {
		return err
	}

	tx, err := mp.pub.CreateAndSignTx(g64)
	if err != nil {
		return err
	}

	mp.part0Tx = tx
	mp.part0TxId = tx.TxHash().String()

	return nil
}

func (mp *MultiPart) genPart(i uint32) ([]byte, error) {
	var ref *oip.Txid = nil
	baseDataSize := baseDataSizeX
	if i == 0 {
		baseDataSize = baseDataSize0
	} else {
		ref = oip.TxidFromString(mp.part0TxId)
	}

	partFloData := mp.remaining
	if len(partFloData) > baseDataSize {
		partFloData = mp.remaining[0:baseDataSize]
	}

	lastGrowthSize := 0
	sizing := true

encodePart:
	part := &oip.MultiPart{
		CurrentPart: i,
		CountParts:  mp.partCount,
		RawData:     partFloData,
		Reference:   ref,
	}

	serPart, err := mp.genSerializedSignedMessage(part)
	if err != nil {
		return nil, err
	}

	g64, err := toGp64(serPart)
	if err != nil {
		return nil, err
	}
	p64 := toP64(serPart)

	serFloData := g64
	if len(p64) < len(g64) {
		serFloData = p64
	}

	if len(serFloData) > MaxFloDataLen {
		delta := len(serFloData) - MaxFloDataLen
		if delta >= len(partFloData) {
			delta = len(partFloData) - 1
		}
		newLen := len(partFloData) - delta
		partFloData = mp.remaining[:newLen]
		goto encodePart
	}
	if sizing && len(serFloData) < MaxFloDataLen-5 {
		delta := MaxFloDataLen - len(serFloData)
		newLen := len(partFloData) + delta
		if newLen > len(mp.remaining) {
			newLen = len(mp.remaining)
		}
		partFloData = mp.remaining[:newLen]
		if len(serFloData) == lastGrowthSize {
			sizing = false
		}
		lastGrowthSize = len(serFloData)
		goto encodePart
	}

	pp := pendingPart{
		mp:      part,
		encoded: serFloData,
	}

	if uint32(len(mp.parts)) != i {
		return nil, errors.New("parts generated out of order")
	}

	mp.parts = append(mp.parts, pp)
	mp.consumed = mp.floData[:len(mp.consumed)+len(partFloData)]
	mp.remaining = mp.remaining[len(partFloData):]

	return g64, nil
}

func (mp *MultiPart) genSerializedSignedMessage(o *oip.MultiPart) ([]byte, error) {
	serializedProtoMessage, err := proto.Marshal(o)
	if err != nil {
		return nil, err
	}
	spm64 := base64.StdEncoding.EncodeToString(serializedProtoMessage)
	sig64, err := flosig.SignMessagePk(spm64, "Florincoin", mp.wif.PrivKey, mp.wif.CompressPubKey)
	if err != nil {
		return nil, err
	}
	sig, err := base64.StdEncoding.DecodeString(sig64)
	if err != nil {
		return nil, err
	}
	msg := &oip.SignedMessage{
		SerializedMessage: serializedProtoMessage,
		MessageType:       oip.MessageTypes_Multipart,
		SignatureType:     oip.SignatureTypes_Flo,
		PubKey:            mp.addrBytes,
		Signature:         sig,
	}
	serializedSignedMessage, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return serializedSignedMessage, nil
}

func toP64(data []byte) []byte {
	enc := base64.StdEncoding
	out := make([]byte, 4+enc.EncodedLen(len(data)))
	copy(out, "p64:")
	enc.Encode(out[4:], data)
	return out
}

func toGp64(data []byte) ([]byte, error) {
	buf := &bytes.Buffer{}
	gzw := gzip.NewWriter(buf)

	n, err := gzw.Write(data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, errors.New("unable to encode data")
	}

	err = gzw.Close()
	if err != nil {
		return nil, err
	}

	enc := base64.StdEncoding
	out := make([]byte, 5+enc.EncodedLen(len(buf.Bytes())))
	copy(out, "gp64:")
	enc.Encode(out[5:], buf.Bytes())

	return out, nil
}
