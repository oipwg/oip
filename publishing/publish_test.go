package publishing

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitspill/flod/chaincfg"
	"github.com/bitspill/flod/rpcclient"
	"github.com/bitspill/floutil"
	"github.com/davecgh/go-spew/spew"
)

var floAddress = "FUE5a3b45n9Jfr5apoq7VnsPxMkTVBnLJQ"
var floWifKey = "RBmWKRJpujYmRsRkGBx4AY2rL1GkiDdMfBv52625CzZBa7Ni4Peu"

func TestRPC(t *testing.T) {
	cfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:8334",
		Endpoint:     "ws",
		User:         "user",
		Pass:         "pass",
		DisableTLS:   true,
		Certificates: nil,
	}
	client, err := rpcclient.New(cfg, nil)
	if err != nil {
		t.Skip("flod not available")
	}

	addr, err := floutil.DecodeAddress(floAddress, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
	wif, err := floutil.DecodeWIF(floWifKey)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()

	pub, err := NewAddress(client, addr, wif, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}

	err = pub.UpdateUtxoSet()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("utxo built", time.Since(start))
	res, err := pub.SendToBlockchainMultipart([]byte(randomText))
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(res)

	fmt.Println("total", time.Since(start))

}

var randomText = `Talent she for lively eat led sister. Entrance strongly packages she out rendered get quitting denoting led. Dwelling confined improved it he no doubtful raptures. Several carried through an of up attempt gravity. Situation to be at offending elsewhere distrusts if. Particular use for considered projection cultivated. Worth of do doubt shall it their. Extensive existence up me contained he pronounce do. Excellence inquietude assistance precaution any impression man sufficient. 

Surprise steepest recurred landlord mr wandered amounted of. Continuing devonshire but considered its. Rose past oh shew roof is song neat. Do depend better praise do friend garden an wonder to. Intention age nay otherwise but breakfast. Around garden beyond to extent by. 

As am hastily invited settled at limited civilly fortune me. Really spring in extent an by. Judge but built gay party world. Of so am he remember although required. Bachelor unpacked be advanced at. Confined in declared marianne is vicinity. 

In by an appetite no humoured returned informed. Possession so comparison inquietude he he conviction no decisively. Marianne jointure attended she hastened surprise but she. Ever lady son yet you very paid form away. He advantage of exquisite resolving if on tolerably. Become sister on in garden it barton waited on. 

An country demesne message it. Bachelor domestic extended doubtful as concerns at. Morning prudent removal an letters by. On could my in order never it. Or excited certain sixteen it to parties colonel. Depending conveying direction has led immediate. Law gate her well bed life feet seen rent. On nature or no except it sussex. 

Neat own nor she said see walk. And charm add green you these. Sang busy in this drew ye fine. At greater prepare musical so attacks as on distant. Improving age our her cordially intention. His devonshire sufficient precaution say preference middletons insipidity. Since might water hence the her worse. Concluded it offending dejection do earnestly as me direction. Nature played thirty all him. 

Sportsman do offending supported extremity breakfast by listening. Decisively advantages nor expression unpleasing she led met. Estate was tended ten boy nearer seemed. As so seeing latter he should thirty whence. Steepest speaking up attended it as. Made neat an on be gave show snug tore. 

Perceived end knowledge certainly day sweetness why cordially. Ask quick six seven offer see among. Handsome met debating sir dwelling age material. As style lived he worse dried. Offered related so visitor we private removed. Moderate do subjects to distance. 

Led ask possible mistress relation elegance eat likewise debating. By message or am nothing amongst chiefly address. The its enable direct men depend highly. Ham windows sixteen who inquiry fortune demands. Is be upon sang fond must shew. Really boy law county she unable her sister. Feet you off its like like six. Among sex are leave law built now. In built table in an rapid blush. Merits behind on afraid or warmly. 

Death weeks early had their and folly timed put. Hearted forbade on an village ye in fifteen. Age attended betrayed her man raptures laughter. Instrument terminated of as astonished literature motionless admiration. The affection are determine how performed intention discourse but. On merits on so valley indeed assure of. Has add particular boisterous uncommonly are. Early wrong as so manor match. Him necessary shameless discovery consulted one but.`
