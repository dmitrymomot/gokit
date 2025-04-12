package template

import (
	"html/template"
	"sync"
	"time"
)

// CacheEntry represents a cached template with metadata
type CacheEntry struct {
	Template *template.Template
	CreatedAt time.Time
	ExpireAt  time.Time
}

// CacheManager manages the template cache
type CacheManager struct {
	Cache map[string]*CacheEntry
	mutex sync.RWMutex
	TTL   time.Duration // Time to live for cache entries
}

// NewCacheManager creates a new cache manager with the specified TTL
func NewCacheManager(ttl time.Duration) *CacheManager {
	return &CacheManager{
		Cache: make(map[string]*CacheEntry),
		TTL:   ttl,
	}
}

// Get retrieves a template from the cache if it exists and is not expired
func (cm *CacheManager) Get(key string) *template.Template {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	entry, exists := cm.Cache[key]
	if !exists {
		return nil
	}

	// Check if the entry has expired
	if time.Now().After(entry.ExpireAt) {
		// Don't delete it here to avoid write lock, let cleanup handle it
		return nil
	}

	return entry.Template
}

// Set adds or updates a template in the cache
func (cm *CacheManager) Set(key string, tmpl *template.Template) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	cm.Cache[key] = &CacheEntry{
		Template: tmpl,
		CreatedAt: now,
		ExpireAt:  now.Add(cm.TTL),
	}
}

// Delete removes a template from the cache
func (cm *CacheManager) Delete(key string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.Cache, key)
}

// Clear removes all templates from the cache
func (cm *CacheManager) Clear() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.Cache = make(map[string]*CacheEntry)
}

// Cleanup removes expired entries from the cache
// This should be called periodically to prevent memory leaks
func (cm *CacheManager) Cleanup() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	for key, entry := range cm.Cache {
		if now.After(entry.ExpireAt) {
			delete(cm.Cache, key)
		}
	}
}

// StartCleanupRoutine starts a goroutine to periodically clean up expired cache entries
func (cm *CacheManager) StartCleanupRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			cm.Cleanup()
		}
	}()
}

// WithCacheTTL sets the cache time-to-live duration
func WithCacheTTL(ttl time.Duration) Option {
	return func(e *Engine) error {
		e.cacheManager = NewCacheManager(ttl)
		e.cacheManager.StartCleanupRoutine(ttl / 2) // Clean up at half the TTL interval
		return nil
	}
}