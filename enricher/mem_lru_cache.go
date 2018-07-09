package enricher

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

// MemLRUCache is an in-memory LRU/passthru cache that satisfies the Enricher
// interface. It can be used to add local caching to a parent Enricher.
type MemLRUCache struct {
	sync.Mutex
	cache  map[string]*appMetadata
	parent Enricher
}

type appMetadata struct {
	AppMetadata
	lastSeen time.Time
}

// NewMemLRUCache creates a MemLRUCache that uses the provided parent Enricher
// to resolve misses
func NewMemLRUCache(parent Enricher) Enricher {
	return &MemLRUCache{
		cache:  make(map[string]*appMetadata),
		parent: parent,
	}
}

// GetAppMetadata returns the application metadata for the specified application
// GUID. If the metadata is in the in-memory cache, it is returned directly;
// otherwise the parent Enriched is queried and the in-memory cache updated.
func (e *MemLRUCache) GetAppMetadata(appGUID string) (AppMetadata, error) {
	e.Lock()
	amd, ok := e.cache[appGUID]
	if ok {
		amd.lastSeen = time.Now() // this needs to be done while Lock()ed to avoid races
		e.Unlock()
		return amd.AppMetadata, nil
	}
	e.Unlock()

	md, err := e.parent.GetAppMetadata(appGUID)
	if err != nil {
		return AppMetadata{}, errors.Wrap(err, "getting app metadata from cf")
	}

	e.Lock()
	e.cache[appGUID] = &appMetadata{
		AppMetadata: md,
		lastSeen:    time.Now(),
	}
	e.Unlock()

	return md, nil
}

// Expire removes from the cache all entries that have not been queried in
// the specified duration.
func (e *MemLRUCache) Expire(olderThan time.Duration) {
	now := time.Now()
	e.Lock()
	defer e.Unlock()
	for k, v := range e.cache {
		if now.Sub(v.lastSeen) >= olderThan {
			delete(e.cache, k)
		}
	}
}

// Warmup merges the list of application metadata returned by the supplied
// function with the metadata in cache. If the cache already contains metadata
// about a certain application, but the supplied function does not include new
// metadata for it, the cached metadata is not modified/overwritten.
func (e *MemLRUCache) Warmup(mds []AppMetadata) {
	e.Lock()
	defer e.Unlock()
	now := time.Now()
	for _, md := range mds {
		e.cache[md.AppGUID] = &appMetadata{
			AppMetadata: md,
			lastSeen:    now,
		}
	}
}
