package web

import (
	"archive/zip"
	"strings"
	"sync"
)

type archiveCache struct {
	mu sync.Mutex
	// TODO: bound this cache (maybe LRU?). It's fine for now since I'll have just one archive
	cache map[string]*zip.Reader
}

func (a *archiveCache) Get(hash string) (*zip.Reader, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	r, ok := a.cache[strings.ToLower(hash)]
	return r, ok
}

func (a *archiveCache) Set(hash string, r *zip.Reader) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cache[strings.ToLower(hash)] = r
}
