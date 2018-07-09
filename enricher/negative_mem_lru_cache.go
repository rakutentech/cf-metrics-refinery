package enricher

import (
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// NegativeMemLRUCache is an in-memory LRU/passthru cache that satisfies the Enricher
// interface. It can be used to add local caching to a parent Enricher.
type NegativeMemLRUCache struct {
	sync.Mutex
	cache  map[string]*appNotFound
	parent Enricher
}

type appNotFound struct {
	lastNotFound time.Time
}

// NewNegativeMemLRUCache creates a NegativeMemLRUCache that uses the provided parent Enricher
// to add negative cache
func NewNegativeMemLRUCache(parent Enricher) Enricher {
	return &NegativeMemLRUCache{
		cache:  make(map[string]*appNotFound),
		parent: parent,
	}
}

// GetAppMetadata returns the application metadata for the specified application
// GUID. If the metadata is in the in-memory negative cache, it is returned empty metadata directly;
// otherwise the parent Enriched is queried and the in-memory negative cache updated if the parent could not find the app.
func (nm *NegativeMemLRUCache) GetAppMetadata(appGUID string) (AppMetadata, error) {
	errNotFound := "CF-AppNotFound"

	nm.Lock()
	_, ok := nm.cache[appGUID]
	if ok {
		nm.Unlock()
		return AppMetadata{}, errors.New(errNotFound)
	}
	nm.Unlock()

	md, err := nm.parent.GetAppMetadata(appGUID)
	if err != nil && strings.Contains(err.Error(), errNotFound) {
		nm.Lock()
		nm.cache[appGUID] = &appNotFound{
			lastNotFound: time.Now(),
		}
		nm.Unlock()

		return AppMetadata{}, errors.Wrap(err, "getting app metadata from mem")
	}

	return md, err
}

// Expire removes from the negative cache all entries that have not been queried in
// the specified duration.
func (nm *NegativeMemLRUCache) Expire(olderThan time.Duration) {
	now := time.Now()
	nm.Lock()
	defer nm.Unlock()
	for k, v := range nm.cache {
		if now.Sub(v.lastNotFound) >= olderThan {
			delete(nm.cache, k)
		}
	}
}

// Warmup compares the list of running applications with the ones in the negative cache. If the application in the negative cache is in the list of running applications, then remove it, otherwise, keep it and update the time last not found.
func (nm *NegativeMemLRUCache) Warmup(mds []AppMetadata) {
	nm.Lock()
	defer nm.Unlock()
	for _, md := range mds {
		_, found := nm.cache[md.AppGUID]
		if found {
			delete(nm.cache, md.AppGUID)
		}
	}
}
