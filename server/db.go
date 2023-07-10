package main

type database struct {
	db *map[string]string
}

func (db *database) get(key string) string {
	d := *db.db
	return d[key]
}
