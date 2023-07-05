package main

type database struct {
	db *map[string]string
}

func (db *database) keyExists(key string) bool {
	d := *db.db
	_, ok := d[key]
	return ok
}

func (db *database) delete(key string) {
	delete(*db.db, key)
}

func (db *database) get(key string) (string, bool) {
	if !db.keyExists(key) {
		return "", false
	}
	d := *db.db
	return d[key], true
}

func (db *database) put(key string, value string) {
	d := *db.db
	d[key] = value
}
