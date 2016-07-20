package wemvc

import "time"

type CacheManager interface {
	Get(key string) (interface{}, error)
	Set(key string, data interface{}, dependency CacheDependency, expire time.Time) error
	Delete(key string)
	Clear()
	AllKeys() ([]string, error)
}

