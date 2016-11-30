package wemvc

import (
	"sort"
	"strings"
)

type filterContainer struct {
	filters map[string][]CtxFilter
}

func (fc *filterContainer) execFilters(urlPath string, ctx *Context) {
	if len(fc.filters) < 1 {
		return
	}
	if !strings.HasSuffix(urlPath, "/") {
		urlPath = strAdd(urlPath, "/")
	}
	var tmpFilters = fc.filters
	var keys []string
	for key := range tmpFilters {
		if strings.HasPrefix(urlPath, key) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		filterFunc := tmpFilters[key]
		for _, f := range filterFunc {
			if f(ctx); ctx.ended {
				return
			}
		}
	}
}

func (fc *filterContainer) addFilter(pathPrefix string, filter CtxFilter) {
	if !strings.HasPrefix(pathPrefix, "") {
		panic(errFilterPrefix)
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = strAdd(pathPrefix, "/")
	}
	if fc.filters == nil {
		fc.filters = make(map[string][]CtxFilter)
	}
	fc.filters[pathPrefix] = append(fc.filters[pathPrefix], filter)
}
