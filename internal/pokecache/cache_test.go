package pokecache

import (
    "testing"
    "time"
)

func TestAddGet(t *testing.T) {
    // Create a new cache with 5 second interval
    cache := NewCache(5 * time.Second)
    
    // Add some test data
    testURL := "https://example.com"
    testData := []byte("test data")
    
    // Add to cache
    cache.Add(testURL, testData)
    
    // Try to get the data back
    val, ok := cache.Get(testURL)
    
    // Check if we got data back
    if !ok {
        t.Error("expected to find data in cache, but didn't")
    }
    
    // Check if the data matches
    if string(val) != string(testData) {
        t.Errorf("expected %s but got %s", string(testData), string(val))
    }
}

func TestReapLoop(t *testing.T) {
    // Create cache with very short interval for testing
    interval := 10 * time.Millisecond
    cache := NewCache(interval)
    
    // Add test data
    cache.Add("test", []byte("test data"))
    
    // Verify data exists
    _, ok := cache.Get("test")
    if !ok {
        t.Error("data should exist before interval")
    }
    
    // Wait for reaper to run
    time.Sleep(interval * 2)
    
    // Verify data was reaped
    _, ok = cache.Get("test")
    if ok {
        t.Error("data should have been reaped")
    }
}
