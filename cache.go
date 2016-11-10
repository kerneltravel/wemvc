package wemvc

import (
	"strings"
	"path"
	"errors"
	"fmt"
	"time"
)

type cacheData struct {
	data interface{}
	expire time.Time
}

type Cache struct {
	dataMap     map[string]cacheData
	gcFrequency int64
	fileWatcher *FileWatcher
}

func (c *Cache) Get(name string) interface{} {
	if c.dataMap == nil {
		return nil
	}
	data,ok := c.dataMap[name]
	if ok {
		if time.Now().Before(data.expire) {
			return data.data
		}
		delete(c.dataMap, name)
		return nil
	}
	return nil
}

func (c *Cache) AllKeys(name string)  {
	var keys []string
	for key := range c.dataMap {
		keys = append(keys, key)
	}
	return keys
}

func (c *Cache) AllData() map[string]interface{} {
	var data = make(map[string]interface{})
	var now = time.Now()
	for key, value := range c.dataMap {
		if now.Before(value.expire) {
			data[key] = value.data
		}
	}
	return data
}

func (c *Cache) Add(name string, data interface{}, dependencyFile string, expire *time.Time)  {
	// todo: write add cache code here
}

type cacheDetector struct {
	file string
}

func (cd *cacheDetector) CanHandle(path string) bool {
	return strings.EqualFold(cd.file, path)
}

func NewCacheDetector(filePath string) (WatcherDetector, error) {
	if len(filePath) == 0 {
		return nil, errors.New("'filePath' argument cannot be empty")
	}
	fixPath := path.Clean(fixPath(filePath))
	if !IsFile(fixPath) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}
	return &cacheDetector{file:fixPath}
}