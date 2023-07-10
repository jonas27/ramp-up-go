package main

type database struct {
	db *map[string]string
}

func (db *database) delete(key string) {
	delete(*db.db, key)
}

func (db *database) get(key string) (string, bool) {
	d := *db.db
	value, ok := d[key]
	if !ok {
		return "", false
	}
	return value, ok
}

func (db *database) put(key string, value string) {
	d := *db.db
	d[key] = value
}
