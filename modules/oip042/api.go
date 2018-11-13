package oip042

import (
	"context"
	"net/http"
	"strconv"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/httpapi"
	"github.com/gorilla/mux"
	"gopkg.in/olivere/elastic.v6"
)

var artRouter = httpapi.NewSubRoute("/oip042/artifact")

func init() {
	artRouter.HandleFunc("/get/latest/{limit:[0-9]+}", handleLatest).Queries("nsfw", "{nsfw}")
	artRouter.HandleFunc("/get/latest/{limit:[0-9]+}", handleLatest)
}

func handleLatest(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	size, _ := strconv.ParseInt(opts["limit"], 10, 0)
	if size <= 0 || size > 1000 {
		size = -1
	}

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
	)

	if n, ok := opts["nsfw"]; ok {
		nsfw, _ := strconv.ParseBool(n)
		if nsfw == false {
			q.MustNot(elastic.NewTermQuery("artifact.info.nsfw", true))
		}
		log.Info("nsfw: %t", nsfw)
	}

	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")

	results, err := datastore.Client().
		Search(datastore.Index(oip042ArtifactIndex)).
		Type("_doc").
		Query(q).
		Size(int(size)).
		Sort("meta.time", false).
		FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		httpapi.RespondJSON(w, 500, map[string]interface{}{
			"error": "database error",
		})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	httpapi.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(results.Hits.Hits),
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}
