package pokecache

import (
    "sync"
    "time"
)

type Cache struct {
	entries 	map[string]cacheEntry
	mutex 		sync.Mutex
}

type cacheEntry struct {
	createdAt	time.Time
	val			[]byte
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]cacheEntry),
	}
	
	go c.reapLoop(interval)
	
	return c
}

func (c *Cache) reapLoop(interval time.Duration) {
    ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		c.mutex.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.Sub(entry.createdAt) > interval {
				delete(c.entries, key)
			}
		}
		c.mutex.Unlock()
	}
}

func (c *Cache) Add(key string, val []byte) {
    // Lock the mutex while we modify the map
    c.mutex.Lock()
    
    // Create a new cache entry with current time and data
    c.entries[key] = cacheEntry{
        createdAt: time.Now(),
        val:       val,
    }
    
    // Don't forget to unlock!
    c.mutex.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
    // Lock the mutex while we read from the map
    c.mutex.Lock()
    
    // Get the entry
    entry, exists := c.entries[key]
    
    // Don't forget to unlock!
    c.mutex.Unlock()
    
    // If we found it, return the value and true
    // If we didn't find it, return nil and false
    return entry.val, exists
}