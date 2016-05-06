package session

import "net/http"

type Store interface {
	Set(key, value interface{}) error     //set session value
	Get(key interface{}) interface{}      //get session value
	Delete(key interface{}) error         //delete session value
	SessionID() string                    //back current sessionID
	SessionRelease(w http.ResponseWriter) // release the resource & save data to provider & return the data
	Flush() error                         //delete all data
}

// Provider contains global session methods and saved SessionStores.
// it can operate a SessionStore by its id.
type Provider interface {
	SessionInit(gclifetime int64, config string) error
	SessionRead(sid string) (Store, error)
	SessionExist(sid string) bool
	SessionRegenerate(oldsid, sid string) (Store, error)
	SessionDestroy(sid string) error
	SessionAll() int //get all active session
	SessionGC()
}

var provides = make(map[string]Provider)

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provide
}