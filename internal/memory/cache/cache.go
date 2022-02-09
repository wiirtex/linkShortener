package cache

import (
	"errors"
	"linkShortener/internal/memory"
	"linkShortener/internal/memory/db"
	"linkShortener/pkg/helpers"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type cacheMemory struct {
	Longs  map[string]memory.MemoryRequest
	Shorts map[int64]memory.MemoryRequest
}

func (c *cacheMemory) ClearOutOfDate() int {
	lock.Lock()
	defer lock.Unlock()
	now := time.Now().UTC()
	var deleted = 0

	for key, value := range c.Longs {
		if value.CreatedAt.Add(helpers.GetConfig().CacheTimeToLive).Before(now) {
			delete(c.Longs, key)
			deleted += 1
		}
	}
	for key, value := range c.Shorts {
		if value.CreatedAt.Add(helpers.GetConfig().CacheTimeToLive).Before(now) {
			delete(c.Shorts, key)
			deleted += 1
		}
	}

	return deleted
}

// O(logn)
func (c *cacheMemory) FindEntry(m memory.MemoryRequest) (entry memory.MemoryRequest, err error) {
	lock.Lock()
	defer lock.Unlock()
	if m.Long != "" && m.Short == 0 {
		value, ok := c.Longs[m.Long]
		if !ok {
			return memory.MemoryRequest{}, errors.New("No such entry")
		}
		return value, nil
	} else if m.Long == "" && m.Short != 0 {
		value, ok := c.Shorts[m.Short]
		if !ok {
			return memory.MemoryRequest{}, errors.New("No such entry")
		}
		return value, nil
	} else {
		return m, errors.New("No such entry") // FIXME: error incorrect input data
	}

}

// O(logn) amortised
func (c *cacheMemory) AddEntry(m memory.MemoryRequest) (err error) {
	lock.Lock()
	defer lock.Unlock()
	c.Longs[m.Long] = m
	c.Shorts[m.Short] = m
	return nil
}

func init() {
	cache := GetDBCacheInstance()
	logger := logrus.New()
	logEntry := logrus.NewEntry(logger)

	go func() {
		for range time.Tick(time.Minute) {
			deleted := cache.ClearOutOfDate()
			if deleted != 0 {
				logEntry.Infof("Cleared %v out of date entries in cache", deleted)
			}
		}
	}()
}

type dbWithCache struct {
	cacheMemory // TODO: make this inaccessible not from this packet
	db          memory.Memory
}

var cacheSingleInstance *dbWithCache
var lock = sync.Mutex{} // TODO: check benchmark of this implementation of singletone

func GetDBCacheInstance() *dbWithCache {
	if cacheSingleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		cacheSingleInstance = &dbWithCache{
			db: db.GetDBInstance(),
			cacheMemory: cacheMemory{
				Longs:  make(map[string]memory.MemoryRequest),
				Shorts: make(map[int64]memory.MemoryRequest),
			},
		}
	}
	return cacheSingleInstance
}

func (c *dbWithCache) AddEntry(entry memory.MemoryRequest) (int64, error) {
	r, err := c.GetEntry(entry)
	if err != nil && err.Error() == "No such entry" {
		shortLink, err := c.db.AddEntry(entry)
		entry.Short = shortLink
		if err == nil {
			err = c.cacheMemory.AddEntry(entry)
			// TODO: make this more transactional (delete from db if cache didn't save the entry)
			return shortLink, err
		}
		return 0, err
	} else {
		return r.Short, err
	}
}
func (c *dbWithCache) GetEntry(entry memory.MemoryRequest) (memory.MemoryRequest, error) {
	r, err := c.cacheMemory.FindEntry(entry)
	if err == nil {
		return r, nil
	}
	if err.Error() != "No such entry" {
		return r, err
	}

	r, err = c.db.GetEntry(entry)
	if err == nil {
		err = c.cacheMemory.AddEntry(r)
		return r, err
	}
	return r, err
}

func (c *dbWithCache) Clear() error {
	lock.Lock()
	defer lock.Unlock()
	c.cacheMemory.Longs = make(map[string]memory.MemoryRequest)
	c.cacheMemory.Shorts = make(map[int64]memory.MemoryRequest)
	return nil
}
