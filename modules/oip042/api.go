package oip042

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/oipwg/oip/httpapi"
	"gopkg.in/olivere/elastic.v6"
)

var artRouter = httpapi.NewSubRoute("/oip042/artifact")

func init() {
	artRouter.HandleFunc("/get/latest", handleLatest).Queries("nsfw", "{nsfw}")
	artRouter.HandleFunc("/get/latest", handleLatest)
}

var (
	o42ArtifactFsc = elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time", "meta.type")
)

func handleLatest(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
	)

	if n, ok := opts["nsfw"]; ok {
		nsfw, _ := strconv.ParseBool(n)
		if !nsfw {
			q.MustNot(elastic.NewTermQuery("artifact.info.nsfw", true))
		}
		log.Info("nsfw: %t", nsfw)
	}

	searchService := httpapi.BuildCommonSearchService(
		r.Context(),
		[]string{oip042ArtifactIndex},
		q,
		[]elastic.SortInfo{
			{Field: "meta.time", Ascending: false},
			{Field: "meta.txid", Ascending: true},
		},
		o42ArtifactFsc,
	)
	httpapi.RespondSearch(w, searchService)
}
