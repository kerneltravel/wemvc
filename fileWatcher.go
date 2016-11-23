package wemvc

import (
	"errors"
	"path"

	"github.com/howeyc/fsnotify"
)

// WatcherHandler the watcher handler interface
type WatcherHandler interface {
	CanHandle(path string) bool
	Handle(ev *fsnotify.FileEvent)
}

// FileWatcher the file watcher struct
type FileWatcher struct {
	watcher  *fsnotify.Watcher
	handlers []WatcherHandler
	started  bool
}

// AddWatch add path to watch
func (fw *FileWatcher) AddWatch(path string) error {
	return fw.watcher.Watch(path)
}

// RemoveWatch remove path from watcher
func (fw *FileWatcher) RemoveWatch(strFile string) error {
	return fw.watcher.RemoveWatch(strFile)
}

// AddHandler add file watcher handler
func (fw *FileWatcher) AddHandler(detector WatcherHandler) error {
	if detector == nil {
		return errors.New("The parameter 'detector' cannot be nil")
	}
	fw.handlers = append(fw.handlers, detector)
	return nil
}

// Start star the file watcher
func (fw *FileWatcher) Start() {
	if fw.started {
		return
	}
	fw.started = true
	go func() {
		for {
			select {
			case ev := <-fw.watcher.Event:
				for _, detector := range fw.handlers {
					if detector.CanHandle(path.Clean(ev.Name)) {
						detector.Handle(ev)
					}
				}
			}
		}
	}()
}

// NewWatcher create the new watcher
func NewWatcher() (*FileWatcher, error) {
	tmpWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &FileWatcher{
		watcher: tmpWatcher,
	}
	return w, nil
}
