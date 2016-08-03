package wemvc

import (
	"time"
	"net/http"
	"fmt"
	"net/url"
	"encoding/hex"
	"crypto/rand"
)

type SessionManager struct {
	provider SessionProvider
	config   *SessionConfig
}

// NewManager Create new Manager with provider name and json config string.
// provider name:
// 1. cookie
// 2. file
// 3. memory
// 4. redis
// 5. mysql
// xml config:
// 1. is https  default false
// 2. hashfunc  default sha1
// 3. hashkey default beegosessionkey
// 4. maxage default is none
func NewSessionManager(provideName string, config *SessionConfig) (*SessionManager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	config.EnableSetCookie = true
	if config.MaxLifetime == 0 {
		config.MaxLifetime = config.GcLifetime
	}
	err := provider.SessionInit(config.MaxLifetime, config.ProviderConfig)
	if err != nil {
		return nil, err
	}

	if config.SessionIDLength == 0 {
		config.SessionIDLength = 32
	}

	return &SessionManager{
		provider: provider,
		config: config,
	}, nil
}

// getSid retrieves session identifier from HTTP Request.
// First try to retrieve id by reading from cookie, session cookie name is configurable,
// if not exist, then retrieve id from querying parameters.
//
// error is not nil when there is anything wrong.
// sid is empty when need to generate a new session id
// otherwise return an valid session id.
func (manager *SessionManager) getSessionID(r *http.Request) (string, error) {
	cookie, errs := r.Cookie(manager.config.CookieName)
	if errs != nil || cookie.Value == "" || cookie.MaxAge < 0 {
		return "", nil
	}
	// HTTP Request contains cookie for sessionid info.
	return url.QueryUnescape(cookie.Value)
}

// Set cookie with https.
func (manager *SessionManager) isSecure(req *http.Request) bool {
	if !manager.config.Secure {
		return false
	}
	if req.URL.Scheme != "" {
		return req.URL.Scheme == "https"
	}
	if req.TLS == nil {
		return false
	}
	return true
}

func (manager *SessionManager) sessionID() (string, error) {
	b := make([]byte, manager.config.SessionIDLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", fmt.Errorf("Could not successfully read from the system CSPRNG.")
	}
	return hex.EncodeToString(b), nil
}

// SessionStart generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func (manager *SessionManager) SessionStart(w http.ResponseWriter, r *http.Request) (session SessionStore, err error) {
	sessionID, err := manager.getSessionID(r)
	if err != nil {
		return nil, err
	}

	if sessionID != "" && manager.provider.SessionExist(sessionID) {
		return manager.provider.SessionRead(sessionID)
	}

	// Generate a new session
	sessionID, err = manager.sessionID()
	if err != nil {
		return nil, err
	}

	session, err = manager.provider.SessionRead(sessionID)
	cookie := &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    url.QueryEscape(sessionID),
		Path:     "/",
		HttpOnly: true,
		Secure:   manager.isSecure(r),
	}

	if len(manager.config.Domain) > 0 {
		cookie.Domain = manager.config.Domain
	}

	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)
	return
}

// SessionDestroy Destroy session by its id in http request cookie.
func (manager *SessionManager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	manager.provider.SessionDestroy(sid)
	if manager.config.EnableSetCookie {
		expiration := time.Now()
		cookie = &http.Cookie{Name: manager.config.CookieName,
			Path:     "/",
			HttpOnly: true,
			Expires:  expiration,
			MaxAge:   -1}

		http.SetCookie(w, cookie)
	}
}

// GetSessionStore Get SessionStore by its id.
func (manager *SessionManager) GetSessionStore(sid string) (sessions SessionStore, err error) {
	sessions, err = manager.provider.SessionRead(sid)
	return
}

// GC Start session gc process.
// it can do gc in times after gc lifetime.
func (manager *SessionManager) GC() {
	manager.provider.SessionGC()
	time.AfterFunc(time.Duration(manager.config.GcLifetime)*time.Second, func() { manager.GC() })
}

// SessionRegenerateID Regenerate a session id for this SessionStore who's id is saving in http request.
func (manager *SessionManager) SessionRegenerateID(w http.ResponseWriter, r *http.Request) (session SessionStore) {
	sid, err := manager.sessionID()
	if err != nil {
		return
	}
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		//delete old cookie
		session, _ = manager.provider.SessionRead(sid)
		cookie = &http.Cookie{Name: manager.config.CookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			Secure:   manager.isSecure(r),
			Domain:   manager.config.Domain,
		}
	} else {
		oldsid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.SessionRegenerate(oldsid, sid)
		cookie.Value = url.QueryEscape(sid)
		cookie.HttpOnly = true
		cookie.Path = "/"
	}
	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)
	return
}

// GetActiveSession Get all active sessions count number.
func (manager *SessionManager) GetActiveSession() int {
	return manager.provider.SessionAll()
}

// SetSecure Set cookie with https.
func (manager *SessionManager) SetSecure(secure bool) {
	manager.config.Secure = secure
}