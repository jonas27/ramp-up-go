package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const (
	maxKeyLen         = 20
	maxValueLen       = 200
	maxDatabaseLength = 2000
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

type KeyError struct {
	maxLen int
}

func (e *KeyError) Error() string {
	return fmt.Sprintf("error: key exceeds %d characters", e.maxLen)
}

type ValueError struct {
	maxLen int
}

func (e *ValueError) Error() string {
	return fmt.Sprintf("error: value exceeds %d characters", e.maxLen)
}

type DatabaseError struct {
	maxLen int
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("error: database exceeds %d entries", e.maxLen)
}

func (db *database) put(key string, value string) error {
	if len(key) >= maxKeyLen {
		return &KeyError{maxLen: maxKeyLen}
	}
	if len(value) >= maxValueLen {
		return &ValueError{maxLen: maxValueLen}
	}
	if len(db.db) >= maxDatabaseLength {
		return &DatabaseError{maxLen: maxDatabaseLength}
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	db.db[key] = value
	return nil
}

func (db *database) persist() error {
	jsonDB, err := json.Marshal(&db)
	if err != nil {
		return err
	}
	return os.WriteFile("database.json", jsonDB, 0600)
}
