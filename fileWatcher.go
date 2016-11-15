package wemvc

import (
	"github.com/howeyc/fsnotify"
	"errors"
	"path"
)

type WatcherHandler func(ctxData interface{}, ev *fsnotify.FileEvent) bool

type WatcherDetector interface {
	CanHandle(path string) (bool,interface{})
}

type FileWatcher struct {
	watcher *fsnotify.Watcher
	handlers map[WatcherDetector]WatcherHandler
	started  bool
}

func (fw *FileWatcher) AddWatch(path string) error {
	return fw.watcher.Watch(path)
}

func (fw *FileWatcher) RemoveWatch(strFile string) error {
	return fw.watcher.RemoveWatch(strFile)
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
	if fw.started {
		return
	}
	fw.started = true
	go func() {
		for {
			select {
			case ev := <- fw.watcher.Event:
				println(ev.Name)
				for det, h := range fw.handlers {
					ok,ctx := det.CanHandle(path.Clean(ev.Name))
					if ok {
						if !h(ctx, ev) {
							break
						}
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