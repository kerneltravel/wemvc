package wemvc

import (
	"strings"
	"path"
	"errors"
	"fmt"
	"time"
	"github.com/howeyc/fsnotify"
	"sync"
)

type cacheData struct {
	data         interface{}
	dependencies []string
	expire       time.Time
}

type CacheManager struct {
	dataMap     map[string]*cacheData
	gcFrequency time.Duration
	fileWatcher *FileWatcher
	locker      *sync.RWMutex
	started     bool
}

func (c *CacheManager) countDepFileUsage(fPath string) int {
	var count int = 0
	for _, data := range c.dataMap {
		if len(data.dependencies) > 0 {
			for _, f := range data.dependencies {
				if strings.EqualFold(fPath, f) {
					count = count + 1
				}
			}
		}
	}
	return count
}

func (c *CacheManager) Get(name string) interface{} {
	if c.dataMap == nil {
		return nil
	}
	data,ok := c.dataMap[name]
	if ok {
		if time.Now().Before(data.expire) {
			return data.data
		}
		c.locker.Lock()
		delete(c.dataMap, name)
		c.locker.Unlock()
		return nil
	}
	return nil
}

func (c *CacheManager) AllKeys(name string) []string {
	var keys []string
	for key := range c.dataMap {
		keys = append(keys, key)
	}
	return keys
}

func (c *CacheManager) AllData() map[string]interface{} {
	var data = make(map[string]interface{})
	var now = time.Now()
	for key, value := range c.dataMap {
		if now.Before(value.expire) {
			data[key] = value.data
		}
	}
	return data
}

func (c *CacheManager) Add(name string, data interface{}, dependencyFiles []string, expire *time.Time) error {
	if len(name) == 0 {
		return errors.New("The parameter 'name' cannot be empty")
	}
	if data == nil {
		return errors.New("The parameter 'data' cannot be nil")
	}
	c.locker.Lock()
	defer c.locker.Unlock()
	var dFiles []string
	if len(dependencyFiles) != 0 {
		for _, file := range dependencyFiles {
			if len(file) == 0 {
				continue
			}
			fPath := path.Clean(fixPath(file))
			if (!IsFile(fPath)) {
				return fmt.Errorf("The dependency file does not exist: %s", file)
			} else {
				dFiles = append(dFiles, fPath)
			}
		}
	}
	if expire == nil {
		t := time.Date(9999, 12, 31, 23, 59, 59, 999, time.UTC)
		expire = &t
	}
	cData := &cacheData{
		data: data,
		dependencies: dFiles,
		expire: *expire,
	}
	c.locker.Lock()
	c.dataMap[name] = cData
	c.locker.Unlock()
	if len(dFiles) > 0 {
		for _, f := range dFiles {
			c.fileWatcher.AddWatch(f)
		}
	}
	return nil
}

func (c *CacheManager) Remove(name string) {
	if len(name) == 0 {
		return
	}
	data,ok := c.dataMap[name]
	if !ok {
		return
	}
	if len(data.dependencies) > 0 {
		for _, f := range data.dependencies {
			if c.countDepFileUsage(f) == 1 {
				c.fileWatcher.RemoveWatch(f)
			}
		}
	}
	c.locker.Lock()
	delete(c.dataMap, name)
	c.locker.Unlock()
}

func (c *CacheManager) gc() {
	go func() {
		time.Sleep(c.gcFrequency)
		var now = time.Now()
		for name, data := range c.dataMap {
			if now.Before(data.expire) {
				c.locker.Lock()
				delete(c.dataMap, name)
				c.locker.Unlock()
			}
		}
	}()
}

func (c *CacheManager) onDepFileChange(ctxData interface{}, ev *fsnotify.FileEvent) bool {
	name,ok := ctxData.(string)
	if ok {
		c.locker.Lock()
		delete(c.dataMap, name)
		c.locker.Unlock()
	}
	return true
}

func (c *CacheManager) start() {
	if c.started {
		return
	}
	c.started = true
	c.fileWatcher.Start()
	detector := newCacheDetector(c)
	c.fileWatcher.AddHandler(detector, c.onDepFileChange)
	c.gc()
}

func newCacheManager(fw *FileWatcher, gcFrequency time.Duration) *CacheManager {
	return &CacheManager{
		locker: &sync.RWMutex{},
		dataMap: make(map[string]*cacheData),
		gcFrequency: gcFrequency,
		fileWatcher: fw,
	}
}

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