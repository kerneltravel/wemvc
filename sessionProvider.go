package wemvc

import (
	"container/list"
	"sync"
	"time"
)

// SessionProvider contains global session methods and saved SessionStores.
// it can operate a SessionStore by its id.
type SessionProvider interface {
	SessionInit(gcLifetime int64, config string) error
	SessionRead(sid string) (SessionStore, error)
	SessionExist(sid string) bool
	SessionRegenerate(oldSid, sid string) (SessionStore, error)
	SessionDestroy(sid string) error
	SessionAll() int //get all active session
	SessionGC()
}

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func (app *server) RegSessionProvider(name string, provide SessionProvider) Server {
	if app.locked {
		return app
	}
	if provide == nil {
		panic(errSessionProvNil)
	}
	if _, dup := app.sessionProvides[name]; dup {
		panic(errSessionRegTwice(name))
	}
	app.sessionProvides[name] = provide
	return app
}

// MemSessionProvider Implement the provider interface
type MemSessionProvider struct {
	lock        sync.RWMutex             // locker
	sessions    map[string]*list.Element // map in memory
	list        *list.List               // for gc
	maxLifetime int64
	savePath    string
}

// SessionInit init memory session
func (prov *MemSessionProvider) SessionInit(maxLifeTime int64, savePath string) error {
	prov.maxLifetime = maxLifeTime
	prov.savePath = savePath
	return nil
}

// SessionRead get memory session store by sid
func (prov *MemSessionProvider) SessionRead(sid string) (SessionStore, error) {
	prov.lock.RLock()
	if element, ok := prov.sessions[sid]; ok {
		go prov.SessionUpdate(sid)
		prov.lock.RUnlock()
		return element.Value.(*MemSessionStore), nil
	}
	prov.lock.RUnlock()
	prov.lock.Lock()
	newStore := &MemSessionStore{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
	element := prov.list.PushFront(newStore)
	prov.sessions[sid] = element
	prov.lock.Unlock()
	return newStore, nil
}

// SessionExist check session store exist in memory session by sid
func (prov *MemSessionProvider) SessionExist(sid string) bool {
	prov.lock.RLock()
	defer prov.lock.RUnlock()
	if _, ok := prov.sessions[sid]; ok {
		return true
	}
	return false
}

// SessionRegenerate generate new sid for session store in memory session
func (prov *MemSessionProvider) SessionRegenerate(oldsid, sid string) (SessionStore, error) {
	prov.lock.RLock()
	if element, ok := prov.sessions[oldsid]; ok {
		go prov.SessionUpdate(oldsid)
		prov.lock.RUnlock()
		prov.lock.Lock()
		element.Value.(*MemSessionStore).sid = sid
		prov.sessions[sid] = element
		delete(prov.sessions, oldsid)
		prov.lock.Unlock()
		return element.Value.(*MemSessionStore), nil
	}
	prov.lock.RUnlock()
	prov.lock.Lock()
	newStore := &MemSessionStore{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
	element := prov.list.PushFront(newStore)
	prov.sessions[sid] = element
	prov.lock.Unlock()
	return newStore, nil
}

// SessionDestroy delete session store in memory session by id
func (prov *MemSessionProvider) SessionDestroy(sid string) error {
	prov.lock.Lock()
	defer prov.lock.Unlock()
	if element, ok := prov.sessions[sid]; ok {
		delete(prov.sessions, sid)
		prov.list.Remove(element)
		return nil
	}
	return nil
}

// SessionGC clean expired session stores in memory session
func (prov *MemSessionProvider) SessionGC() {
	prov.lock.RLock()
	for {
		element := prov.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*MemSessionStore).timeAccessed.Unix() + prov.maxLifetime) < time.Now().Unix() {
			prov.lock.RUnlock()
			prov.lock.Lock()
			prov.list.Remove(element)
			delete(prov.sessions, element.Value.(*MemSessionStore).sid)
			prov.lock.Unlock()
			prov.lock.RLock()
		} else {
			break
		}
	}
	prov.lock.RUnlock()
}

// SessionAll get count number of memory session
func (prov *MemSessionProvider) SessionAll() int {
	return prov.list.Len()
}

// SessionUpdate expand time of session store by id in memory session
func (prov *MemSessionProvider) SessionUpdate(sid string) error {
	prov.lock.Lock()
	defer prov.lock.Unlock()
	if element, ok := prov.sessions[sid]; ok {
		element.Value.(*MemSessionStore).timeAccessed = time.Now()
		prov.list.MoveToFront(element)
		return nil
	}
	return nil
}
