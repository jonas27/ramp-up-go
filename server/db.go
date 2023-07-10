package main

import (
	"sync"
)

type database struct {
	mu sync.Mutex
	db map[string]string
}

func (db *database) delete(key string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.db, key)
}

func (db *database) get(key string) (string, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	value, ok := db.db[key]
	if !ok {
		return "", false
	}
	return value, ok
}

func (db *database) put(key string, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.db[key] = value
}
