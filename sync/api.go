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

	height := int64(0)
	time := int64(0)

	if lb != nil && lb.Block != nil {
		height = lb.Block.Height
		time = lb.Block.Time
	}

	httpapi.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"IsInitialSync": IsInitialSync,
		"Height":        height,
		"Timestamp":     time,
		"LatestHeight":  count,
		"Progress":      float64(height) / float64(count),
	})
}
