package wemvc

import (
	"time"
	"encoding/json"
)

type sessionData struct {
	Expire time.Time
	Data   map[string]interface{}
}

var dataPool map[string]*sessionData

type defaultSessionProvider struct {
	expireMinute uint64
}

func (s *defaultSessionProvider)SetExpireMinute(m uint64) {
	s.expireMinute = m
}

func (s *defaultSessionProvider) SetExpire(id string, t time.Time) {
	s.printData()
	data := s.dataPool[id]
	if data == nil {
		s.dataPool[id] = &sessionData{
			Expire: t,
			Data: make(map[string]interface{}),
		}
	} else {
		data.Expire = t
	}
}

func (s *defaultSessionProvider) IsExpired(id string) bool {
	s.printData()
	data := s.dataPool[id]
	if data == nil {
		return true
	}
	return time.Now().Before(data.Expire)
}

func (s *defaultSessionProvider) TryDelSession(id string) {
	s.printData()
	delete(s.dataPool, id)
}

func (s *defaultSessionProvider) TryDelSessionData(id, key string) {
	s.printData()
	data := s.dataPool[id]
	if data != nil{
		delete(data.Data, key)
	}
}

func (s *defaultSessionProvider) TryGetSessionData(id, key string) interface{} {
	println("TryGetSessionData")
	s.printData()
	sData := s.dataPool[id]
	s.printData(sData)
	if sData != nil {
		return sData.Data[key]
	}
	return nil
}

func (s *defaultSessionProvider) TrySetSessionData(id, key string, data interface{}) {
	println("AddData")
	s.printData()
	sData := s.dataPool[id]
	if sData == nil {
		s.SetExpire(id, time.Now().Add(time.Duration(s.expireMinute) * time.Minute))
		sData = s.dataPool[id]
	}
	sData.Data[key] = data
}

func (s *defaultSessionProvider) RecycleJob() {
	s.printData()
	for id,data := range s.dataPool {
		if data.Expire.Before(time.Now()) {
			delete(s.dataPool, id)
		}
	}
}

func NewSessionProvider() SessionProvider {
	return &defaultSessionProvider{
		dataPool: make(map[string]*sessionData),
	}
}

func (s *defaultSessionProvider)printData(d... interface{}) {
	if len(d) > 0 {
		data,_ := json.Marshal(d[0])
		println(string(data))
		return
	}
	data,_ := json.Marshal(s.dataPool)
	println(string(data))
}