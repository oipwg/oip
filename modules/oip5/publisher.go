package oip5

import (
	"context"
	"errors"

	"github.com/azer/logger"
	"github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	jsoniter "github.com/json-iterator/go"
	"github.com/oipwg/proto/go/pb_oip5"
	"github.com/oipwg/proto/go/pb_oip5/pb_templates"
	"github.com/spf13/viper"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/config"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
)

var publisherCacheDepth = 1000
var publisherCache *lru.Cache

func init() {
	events.SubscribeAsync("modules:oip5:record", publisherListener)

	publisherCache, _ = lru.New(publisherCacheDepth)

	config.OnPostConfig(func(ctx context.Context) {
		pcd := viper.GetInt("oip.oip5.publisherCacheDepth")
		if pcd != publisherCacheDepth && pcd > 0 {
			publisherCacheDepth = pcd
			publisherCache.Resize(publisherCacheDepth)
		}
	})
}

func GetPublisherName(pubKey string) (string, error) {
	pni, found := publisherCache.Get(pubKey)
	if found {
		return pni.(string), nil
	}

	q := elastic.NewBoolQuery().Must(
		elastic.NewExistsQuery("record.details.tmpl_433C2783.name"),
		elastic.NewTermQuery("meta.signed_by", pubKey),
	)
	results, err := datastore.Client().
		Search(datastore.Index("oip5_record")).
		Type("_doc").
		Query(q).
		Size(1).
		Sort("meta.time", false).
		Do(context.TODO())
	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		return "", err
	}

	if len(results.Hits.Hits) > 0 {
		src := *results.Hits.Hits[0].Source
		pn := jsoniter.Get(src, "record", "details", "tmpl_433C2783", "name").ToString()
		publisherCache.Add(pubKey, pn)
		return pn, nil
	}

	return "", errors.New("unable to locate publisher")
}

func publisherListener(rec *pb_oip5.RecordProto, pubKey []byte, _ *datastore.TransactionData) {
	for _, det := range rec.Details.Details {
		if det.TypeUrl == "type.googleapis.com/oipProto.templates.tmpl_433C2783" {
			rp := &pb_templates.Tmpl_433C2783{}
			err := proto.Unmarshal(det.Value, rp)
			if err != nil {
				return
			}
			publisherCache.Add(string(pubKey), rp.Name)
		}
	}
}
