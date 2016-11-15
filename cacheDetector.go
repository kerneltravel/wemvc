package wemvc

import "strings"

type cacheDetector struct {
	cacheManager *CacheManager
}

func (cd *cacheDetector) CanHandle(path string) (bool, interface{}) {
	for name, data := range cd.cacheManager.dataMap {
		if len(data.dependencies) == 0 {
			continue
		}
		for _, file := range data.dependencies {
			if strings.EqualFold(file, path) {
				return true, name
			}
		}
	}
	return false, nil
}

func newCacheDetector(manager *CacheManager) *cacheDetector {
	return &cacheDetector{cacheManager: manager}
}