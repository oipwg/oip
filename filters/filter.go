package filters

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/azer/logger"
	"github.com/gobuffalo/packr/v2"
	"github.com/oipwg/oip/datastore"
	"github.com/spf13/viper"
	"gopkg.in/olivere/elastic.v6"
)

var filterBox = packr.New("bundled", "./bundled")
var filterMap = make(map[string]int)  // first 8 txid: label id
var filterMapB = make(map[string]int) // first 8 txid: label id
var filterLabels = []string{""}
var filterBundled map[string]int

func InitViper(ctx context.Context) {
	bundled := viper.GetStringSlice("oip.blacklist.bundled")
	err := loadBundledLists(bundled)
	if err != nil {
		log.Error("unable to load bundled lists")
		panic(err)
	}

	filterBundled = filterMapB

	updateRemoteBlacklists()

	refresh := viper.GetString("oip.blacklist.remote.refresh")
	if refresh != "false" {
		d, err := time.ParseDuration(refresh)
		if err != nil {
			log.Error("unable to parse refresh duration", logger.Attrs{"refresh": refresh, "err": err})
		} else {
			go startRefreshInterval(ctx, d)
		}
	}
}

func startRefreshInterval(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-time.After(interval):
			updateRemoteBlacklists()
		case <-ctx.Done():
			return
		}
	}
}

func updateRemoteBlacklists() {
	log.Info("updating remote blacklists")
	remote := viper.GetStringMapString("oip.blacklist.remote.urls")
	for label, url := range remote {
		attrs := logger.Attrs{"label": label, "url": url}

		log.Info("fetching remote list", attrs)

		res, err := http.Get(url)
		if err != nil {
			attrs["err"] = err
			log.Error("unable to get remote list", attrs)
			continue
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			attrs["err"] = err
			log.Error("unable to get remote list", attrs)
			continue
		}
		_ = res.Body.Close()
		processList(string(body), label)
	}

	swapLists()

	log.Info("current filter list size %d", len(filterMap))
}

func loadBundledLists(lists []string) error {
	for _, label := range lists {
		list, err := filterBox.FindString(label + ".txt")
		if err != nil {
			return err
		}
		processList(list, label)
	}

	return nil
}

func processList(list, label string) {
	labelId := getLabelId(label)
	lines := strings.Split(list, "\n")
	for _, line := range lines {
		id := strings.SplitN(line, " ", 2)[0]
		if len(id) == 64 {
			filterMapB[id] = labelId
		}
	}
}

func getLabelId(label string) int {
	for id, l := range filterLabels {
		if label == l {
			return id
		}
	}
	filterLabels = append(filterLabels, label)
	return len(filterLabels) - 1
}

func Add(txid string, label string) {
	// ToDo requires tweaks as manually added items are lost on next remote update
	labelId := getLabelId(label)
	filterMap[txid] = labelId
}

func Contains(txid string) bool {
	_, ok := filterMap[txid]
	return ok
}

func ContainsWithLabel(txid string) (bool, string) {
	lid, ok := filterMap[txid]
	if ok {
		if lid < len(filterLabels) {
			return true, filterLabels[lid]
		}
		return true, ""
	}
	return false, ""
}

func Clear() {
	filterMap = make(map[string]int)
}

func swapLists() {
	var added []string
	var removed []string

	if len(filterMap) == 0 {
		added = make([]string, len(filterMapB))
		i := 0
		for key := range filterMapB {
			added[i] = key
			i++
		}
	} else {
		for key := range filterMapB {
			if _, ok := filterMap[key]; !ok {
				added = append(added, key)
			}
		}
		for key := range filterMap {
			if _, ok := filterMapB[key]; !ok {
				removed = append(removed, key)
			}
		}
	}

	log.Info("Swapping lists", logger.Attrs{
		"lenA":    len(filterMap),
		"lenB":    len(filterMapB),
		"added":   len(added),
		"removed": len(removed),
	})

	filterMap = filterMapB
	filterMapB = make(map[string]int, len(filterBundled))
	for key, value := range filterBundled {
		filterMapB[key] = value
	}

	for _, value := range added {
		label := ""
		lid := filterMap[value]
		if lid < len(filterLabels) {
			label = filterLabels[lid]
		}
		up := elastic.NewBulkUpdateRequest().
			Index(datastore.Index("oip042_artifact")).
			Id(value).
			Type("_doc").
			Doc(map[string]interface{}{
				"meta": map[string]interface{}{
					"blacklist": map[string]interface{}{
						"blacklisted": true,
						"filter":      label,
					},
				},
			})
		datastore.AutoBulk.Add(up)
	}
	for _, value := range removed {
		up := elastic.NewBulkUpdateRequest().
			Index(datastore.Index("oip042_artifact")).
			Id(value).
			Type("_doc").
			Doc(map[string]interface{}{
				"meta": map[string]interface{}{
					"blacklist": map[string]interface{}{
						"blacklisted": false,
						"filter":      "",
					},
				},
			})
		datastore.AutoBulk.Add(up)
	}
}
