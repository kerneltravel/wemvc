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
		urlPath = urlPath + "/"
	}
	var tmpFilters = fc.filters
	var keys []string
	for key := range tmpFilters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if strings.HasPrefix(urlPath+"/", key) {
			for _, f := range tmpFilters[key] {
				f(ctx)
				if ctx.ended {
					return
				}
			}
		}
	}
}

func (fc *filterContainer) setFilter(pathPrefix string, filter CtxFilter) {
	if !strings.HasPrefix(pathPrefix, "") {
		panic(errFilterPrefix)
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	if fc.filters == nil {
		fc.filters = make(map[string][]CtxFilter)
	}
	fc.filters[pathPrefix] = append(fc.filters[pathPrefix], filter)
}
