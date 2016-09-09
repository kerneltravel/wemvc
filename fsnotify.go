package wemvc

import "fmt"

const (
	fsnCREATE = 1
	fsnMODIFY = 2
	fsnDELETE = 4
	fsnRENAME = 8

	fsnALL = fsnMODIFY | fsnDELETE | fsnRENAME | fsnCREATE
)

// Purge events from interal chan to external chan if passes filter
func (w *fsWatcher) purgeEvents() {
	for ev := range w.internalEvent {
		sendEvent := false
		w.fsnmut.Lock()
		fsnFlags := w.fsnFlags[ev.Name]
		w.fsnmut.Unlock()

		if (fsnFlags&fsnCREATE == fsnCREATE) && ev.IsCreate() {
			sendEvent = true
		}

		if (fsnFlags&fsnMODIFY == fsnMODIFY) && ev.IsModify() {
			sendEvent = true
		}

		if (fsnFlags&fsnDELETE == fsnDELETE) && ev.IsDelete() {
			sendEvent = true
		}

		if (fsnFlags&fsnRENAME == fsnRENAME) && ev.IsRename() {
			sendEvent = true
		}

		if sendEvent {
			w.Event <- ev
		}

		// If there's no file, then no more events for user
		// BSD must keep watch for internal use (watches DELETEs to keep track
		// what files exist for create events)
		if ev.IsDelete() {
			w.fsnmut.Lock()
			delete(w.fsnFlags, ev.Name)
			w.fsnmut.Unlock()
		}
	}

	close(w.Event)
}

// Watch a given file path
func (w *fsWatcher) Watch(path string) error {
	w.fsnmut.Lock()
	w.fsnFlags[path] = fsnALL
	w.fsnmut.Unlock()
	return w.watch(path)
}

// Watch a given file path for a particular set of notifications (FSN_MODIFY etc.)
func (w *fsWatcher) WatchFlags(path string, flags uint32) error {
	w.fsnmut.Lock()
	w.fsnFlags[path] = flags
	w.fsnmut.Unlock()
	return w.watch(path)
}

// Remove a watch on a file
func (w *fsWatcher) RemoveWatch(path string) error {
	w.fsnmut.Lock()
	delete(w.fsnFlags, path)
	w.fsnmut.Unlock()
	return w.removeWatch(path)
}

// String formats the event e in the form
// "filename: DELETE|MODIFY|..."
func (e *fileEvent) String() string {
	events := ""

	if e.IsCreate() {
		events += "|" + "CREATE"
	}

	if e.IsDelete() {
		events += "|" + "DELETE"
	}

	if e.IsModify() {
		events += "|" + "MODIFY"
	}

	if e.IsRename() {
		events += "|" + "RENAME"
	}

	if e.IsAttrib() {
		events += "|" + "ATTRIB"
	}

	if len(events) > 0 {
		events = events[1:]
	}

	return fmt.Sprintf("%q: %s", e.Name, events)
}
