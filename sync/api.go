package sync

import (
	"net/http"

	"github.com/azer/logger"
	"github.com/oipwg/oip/flo"
	"github.com/oipwg/oip/httpapi"
)

var syncRouter = httpapi.NewSubRoute("/sync")

func init() {
	syncRouter.HandleFunc("/status", HandleStatus)
}

func HandleStatus(w http.ResponseWriter, _ *http.Request) {
	lb := recentBlocks.PeekFront()

	count, err := flo.GetBlockCount()
	if err != nil {
		log.Error("/sync/status GetBlockCount failed", logger.Attrs{"err": err})
	}

	httpapi.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"IsInitialSync": IsInitialSync,
		"Height":        lb.Block.Height,
		"Timestamp":     lb.Block.Time,
		"LatestHeight":  count,
		"Progress":      float64(lb.Block.Height) / float64(count),
	})
}
