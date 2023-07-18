package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

const (
	maxKeyLen         = 20
	maxValueLen       = 200
	maxDatabaseLength = 2000
	filePerm          = 0o600
)

type database struct {
	mu sync.Mutex
	db map[string]string
}

func (db *database) delete(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	_, ok := db.db[key]
	if !ok {
		return &NoEntryError{key: key}
	}
	delete(db.db, key)
	return nil
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

type NoEntryError struct {
	key string
}

func (e *NoEntryError) Error() string {
	return fmt.Sprintf("error: key \"%s\" does not exist", e.key)
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

func (db *database) put(key string, value string) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if len(key) >= maxKeyLen {
		return 0, &KeyError{maxLen: maxKeyLen}
	}
	if len(value) >= maxValueLen {
		return 0, &ValueError{maxLen: maxValueLen}
	}
	if len(db.db) >= maxDatabaseLength {
		return 0, &DatabaseError{maxLen: maxDatabaseLength}
	}
	_, ok := db.db[key]
	db.db[key] = value
	if !ok {
		return http.StatusCreated, nil
	}
	return http.StatusOK, nil
}

func (db *database) persist() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	jsonDB, err := json.Marshal(&db)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	err = os.WriteFile("./database.json", jsonDB, filePerm)
	if err != nil {
		return fmt.Errorf("can't write to file: %w", err)
	}
	return nil
}
