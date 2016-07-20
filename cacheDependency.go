package wemvc

import (
	"github.com/Simbory/wemvc/utils"
	"github.com/Simbory/wemvc/fsnotify"
)

type CacheDependency interface {
	SetManager(manager CacheManager)
	SetKey(key string)
	Watch()
	GetKey() string
}

type FileCacheDependency struct {
	manager CacheManager
	key     string
	Files   []string
}

func (c *FileCacheDependency) SetManager(m CacheManager) {
	c.manager = m
}

func (c *FileCacheDependency)SetKey(key string) {
	c.key = key
}

var watcher *fsnotify.Watcher
var started bool
var watchMap map[string]string
func (c *FileCacheDependency)Watch() {
	if len(c.Files) < 1 {
		return
	}
	if watcher == nil {
		w,err := fsnotify.NewWatcher()
		if err != nil {
			return
		}
		watcher = w
	}
	for _,f := range c.Files {
		watcher.Watch(f)
	}
	go c.watchFile()
}

func (c *FileCacheDependency)watchFile() {
	if (started) {
		return
	}
	started = true
	for {
		select {
		case ev := <-c.watcher.Event:
			c.manager.Delete(c.GetKey())
			println("Cache deleted reason:", ev.Name)
		}
	}
}

func NewFileCacheDependency(files ...string) CacheDependency {
	if len(files) < 1 {
		return nil
	}
	for _, f := range files {
		if !utils.IsFile(f) {
			return nil
		}
	}
	return &FileCacheDependency{Files:files}
}