package wemvc

import (
	"strings"

	"github.com/howeyc/fsnotify"
)

type cacheDetector struct {
	cacheManager *CacheManager
	cacheKey     string
}

// CanHandle detect the fsnotify path can handled by current detector
func (cd *cacheDetector) CanHandle(path string) bool {
	for name, data := range cd.cacheManager.dataMap {
		if len(data.dependencies) == 0 {
			continue
		}
		for _, file := range data.dependencies {
			if strings.EqualFold(file, path) {
				cd.cacheKey = name
				return true
			}
		}
	}
	return false
}

// Handle handle the fsnotify changes
func (cd *cacheDetector) Handle(ev *fsnotify.FileEvent) {
	cd.cacheManager.locker.Lock()
	delete(cd.cacheManager.dataMap, cd.cacheKey)
	cd.cacheManager.locker.Lock()
}

func newCacheDetector(manager *CacheManager) *cacheDetector {
	return &cacheDetector{cacheManager: manager}
}
