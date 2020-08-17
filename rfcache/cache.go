package rfcache

import (
	"encoding/json"
	"io"
	"sync"
)

// Cache a mem cache
type Cache struct {
	data map[string]string
	sync.RWMutex
}

// NewCache create cache
func NewCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

// Get data
func (c *Cache) Get(key string) string {
	c.RLock()
	defer c.RUnlock()

	return c.data[key]
}

// Set data
func (c *Cache) Set(key, value string) {
	c.Lock()
	defer c.Unlock()

	c.data[key] = value
}

// Marshal cache data to json
func (c *Cache) Marshal() ([]byte, error) {
	c.Lock()
	defer c.Unlock()

	cacheJSON, err := json.Marshal(c.data)
	if err != nil {
		return nil, err
	}

	return cacheJSON, nil
}

// Unmarshal data from io reader
func (c *Cache) Unmarshal(serialized io.ReadCloser) error {
	data := make(map[string]string)

	if err := json.NewDecoder(serialized).Decode(&data); err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	c.data = data

	return nil
}
