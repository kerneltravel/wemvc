package wemvc

import (
	"github.com/howeyc/fsnotify"
	"errors"
	"path"
)

type WatcherHandler func(*fsnotify.FileEvent) bool

type WatcherDetector interface {
	CanHandle(path string) bool
}

type FileWatcher struct {
	watcher *fsnotify.Watcher
	handlers map[WatcherDetector]WatcherHandler
}

func (fw *FileWatcher) AddWatch(path string) error {
	return fw.watcher.Watch(path)
}

func (fw *FileWatcher) RemoveWatch(strFile string) {
	fw.watcher.RemoveWatch(strFile)
}

func (fw *FileWatcher) AddHandler(detector WatcherDetector, h WatcherHandler) error {
	if detector == nil {
		return errors.New("The parameter 'detector' cannot be nil")
	}
	if h == nil {
		return errors.New("The parameter 'h' cannot be nil")
	}
	fw.handlers[detector] = h
	return nil
}

func (fw *FileWatcher) Start() {
	go func() {
		select {
		case ev := <- fw.watcher.Event:
			for det, h := range fw.handlers {
				if det.CanHandle(path.Clean(ev.Name)) {
					if !h(ev) {
						break
					}
				}
			}
		}
	}()
}

func NewWatcher() (*FileWatcher, error) {
	tmpWatcher,err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &FileWatcher{
		handlers: map[WatcherDetector]WatcherHandler{},
		watcher: tmpWatcher,
	}
	return w, nil
}