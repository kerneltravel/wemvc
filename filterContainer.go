package wemvc

import (
	"sort"
	"strings"
)

type filterContainer struct {
	filters map[string][]FilterFunc
}

func (fc *filterContainer) execFilters(ctx *context) bool {
	if len(fc.filters) < 1 {
		return false
	}
	var lowerURL = strings.ToLower(ctx.req.URL.Path)
	var tmpFilters = fc.filters
	var keys []string
	for key := range tmpFilters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if strings.HasPrefix(lowerURL+"/", key) {
			for _, f := range tmpFilters[key] {
				f(ctx)
			}
		}
	}
	return false
}

func (fc *filterContainer) setFilter(pathPrefix string, filter FilterFunc) {
	if !strings.HasPrefix(pathPrefix, "") {
		panic(filterPathError)
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	if fc.filters == nil {
		fc.filters = make(map[string][]FilterFunc)
	}
	fc.filters[strings.ToLower(pathPrefix)] = append(fc.filters[strings.ToLower(pathPrefix)], filter)
}
