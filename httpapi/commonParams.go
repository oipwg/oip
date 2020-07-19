package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/azer/logger"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

var sortRe = regexp.MustCompile(`([0-9a-zA-Z._-]+:[ad]$?)+`)

func parseSort(sort string) ([]elastic.SortInfo, error) {
	var res []elastic.SortInfo

	if !sortRe.MatchString(sort) {
		log.Info("invalid sort request", logger.Attrs{"sort": sort})
		return nil, errors.New("sort doesn't match regex")
	}

	sortRequests := strings.Split(sort, "$")
	for _, sortRequest := range sortRequests {
		split := strings.Split(sortRequest, ":")
		if len(split) == 2 && len(split[1]) == 1 { // shouldn't happen per regex, but play safe
			res = append(res, elastic.SortInfo{
				Ascending: split[1][0] == 'a',
				Field:     split[0],
			})
		}
	}
	return res, nil
}

func commonParameterParser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		log.Info("commonParameterParser called")

		sort := r.FormValue("sort")
		if len(sort) > 0 {
			si, err := parseSort(sort)
			if err != nil {
				log.Error("unable to parse sort for request", logger.Attrs{"sort": sort, "err": err})
			} else {
				ctx = context.WithValue(ctx, oipdSortInfoKey, si)
			}
		}

		after := r.FormValue("after")
		if len(after) > 0 {
			var searchAfter []interface{}
			err := json.Unmarshal([]byte(after), &searchAfter)
			if err != nil {
				log.Error("unable to decode after", logger.Attrs{"after": after, "err": err})
			} else {
				ctx = context.WithValue(ctx, oipdSearchAfterKey, searchAfter)
			}
		}

		limit := r.FormValue("limit")
		var size int64 = -1
		if limit != "" {
			size, _ = strconv.ParseInt(limit, 10, 32)
		}
		ctx = context.WithValue(ctx, oipdSizeKey, int(size))

		var from int64
		pageString := r.FormValue("page")
		page, _ := strconv.ParseInt(pageString, 10, 32)
		if size != -1 && page > 0 {
			from = (page - 1) * size
			if from+size > 10000 {
				log.Error("from+size too large", logger.Attrs{"from": from, "size": size, "page": page})
				RespondJSON(r.Context(), w, 400, map[string]interface{}{
					"error": "page limit exceeded, use 'after' instead of 'page' for deep queries",
				})
				return
			}
		}
		ctx = context.WithValue(ctx, oipdFromKey, int(from))

		pretty := r.FormValue("pretty")
		prettyJson := false
		if pretty != "" {
			prettyJson, _ = strconv.ParseBool(pretty)
		}
		ctx = context.WithValue(ctx, oipdPrettyJsonKey, prettyJson)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type key int

const (
	oipdSortInfoKey key = iota
	oipdSearchAfterKey
	oipdSizeKey
	oipdFromKey
	oipdPrettyJsonKey
)

func GetSortInfoFromContext(ctx context.Context) []elastic.SortInfo {
	if si, ok := ctx.Value(oipdSortInfoKey).([]elastic.SortInfo); ok {
		return si
	}
	return nil
}

func GetSearchAfterFromContext(ctx context.Context) []interface{} {
	if sa, ok := ctx.Value(oipdSearchAfterKey).([]interface{}); ok {
		return sa
	}
	return nil
}

func GetSizeFromContext(ctx context.Context) int {
	if size, ok := ctx.Value(oipdSizeKey).(int); ok {
		return size
	}
	return -1
}

func GetFromFromContext(ctx context.Context) int {
	if from, ok := ctx.Value(oipdFromKey).(int); ok {
		return from
	}
	return 0
}

func GetPrettyJsonFromContext(ctx context.Context) bool {
	if pretty, ok := ctx.Value(oipdPrettyJsonKey).(bool); ok {
		return pretty
	}
	return false
}
