package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

type cacheEntry struct {
	Spdx       string
	Confidence float64
	Update     time.Time
}

type cache struct {
	Entries map[string]cacheEntry
	Expiry  time.Duration
}

func NewCache(expiry time.Duration) *cache {
	c := cache{
		Entries: make(map[string]cacheEntry),
		Expiry:  expiry,
	}
	return &c
}

func LoadCache(filename string) (*cache, error) {
	var c cache

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file failed: %w", err)
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("decoding failed: %w", err)
	}
	return &c, nil
}

func (c *cache) Save(filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("opening file failed: %w", err)
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("encoding failed: %w", err)
	}
	return nil
}

func (c *cache) Add(pkg *packageInfo, spdx string, confidence float64) {
	c.Entries[pkg.name+"@"+pkg.version] = cacheEntry{spdx, confidence, time.Now()}
}

func (c *cache) Get(pkg *packageInfo) (string, float64, bool) {
	entry, found := c.Entries[pkg.name+"@"+pkg.version]
	if !found {
		return "", 0.0, false
	}
	now := time.Now()
	deadline := entry.Update.Add(c.Expiry)
	if c.Expiry > 0 && now.After(deadline) {

		return entry.Spdx, entry.Confidence, false
	}
	return entry.Spdx, entry.Confidence, true
}
