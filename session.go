package wemvc

import (
	"net/http"
	"time"
	"errors"
)

type SessionProvider interface {
	SetExpire(string, time.Time)
	IsExpired(string) bool
	TryDelSessionData(string, string)
	TryDelSession(string)
	TryGetSessionData(string, string) interface{}
	TrySetSessionData(string, string, interface{})
	RecycleJob()
	SetExpireMinute(uint64)
}

type session struct {
	expireMinutes uint64
	id            string
	provider      SessionProvider
}

func (s *session) Get(key string) interface{} {
	return s.provider.TryGetSessionData(s.id, key)
}

func (s *session) Set(key string, data interface{}) {
	s.provider.TrySetSessionData(s.id, key, data)
	s.provider.SetExpire(s.id, time.Now().Add(time.Duration(s.expireMinutes) * time.Minute))
}

func (s *session) Del(key string) {
	s.provider.TryDelSessionData(s.id, key)
}

func (s *session) Clear() {
	s.provider.TryDelSession(s.id)
}

type sessionManager struct {
	ExpireMinutes uint64
	SessionID     string
	provider      SessionProvider
}

func (sMgr *sessionManager) SetProvider(prov SessionProvider) {
	if sMgr.provider != nil {
		panic(errors.New("session provider is already provided"))
	}
	sMgr.provider = prov
	/*
	go func() {
		for {
			sMgr.provider.RecycleJob()
			time.Sleep(10 * time.Second)
		}
	}()
	*/
}

func (sMgr *sessionManager) CreateSessionContext(res http.ResponseWriter, req *http.Request) *session {
	if sMgr.provider == nil {
		panic(errors.New("the session provider cannot be empty"))
	}
	sessionCookie,err := req.Cookie(sMgr.SessionID)
	if err != nil || len(sessionCookie.Value) < 1 {
		sessionCookie = &http.Cookie {
			Name: sMgr.SessionID,
			Value: RandomAlphabetic(32),
			Path: "/",
			Secure: req.URL.Scheme == "https",
		}
		sMgr.provider.SetExpire(sessionCookie.Value, time.Now().Add(time.Duration(sMgr.ExpireMinutes) * time.Minute))
		http.SetCookie(res, sessionCookie)
		req.AddCookie(sessionCookie)
	} else {
		if sMgr.provider.IsExpired(sessionCookie.Value) {
			sMgr.provider.TryDelSession(sessionCookie.Value)
		}
	}
	sMgr.provider.SetExpireMinute(sMgr.ExpireMinutes)
	return &session{
		id: sessionCookie.Value,
		expireMinutes: sMgr.ExpireMinutes,
		provider: sMgr.provider,
	}
}