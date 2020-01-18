package cache

import (
	"github.com/derekyu332/goii/frame/base"
	"github.com/patrickmn/go-cache"
	"time"
)

type ICacheRecord interface {
	CacheKey() string
	CacheExpiration() time.Duration
	ReadOnly() bool
}

var gCache *cache.Cache

func LocalCache() *cache.Cache {
	if gCache == nil {
		gCache = cache.New(5*time.Minute, 10*time.Minute)
	}

	return gCache
}

func GetCacheRecord(key string) base.IActiveRecord {
	if data, found := LocalCache().Get(key); found {
		if record, ok := data.(base.IActiveRecord); ok {
			return record
		}
	}

	return nil
}

func SetCacheRecord(record base.IActiveRecord) {
	if cache_record, ok := record.(ICacheRecord); ok {
		LocalCache().Set(cache_record.CacheKey(), record,
			cache_record.CacheExpiration())
	}
}
