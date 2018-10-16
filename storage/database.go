package storage

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	database *sql.DB
}

func (db Database) OpenDb(path string) {
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	db.database = database
}

func (db Database) IsOpen() bool {
	return db.database != nil
}

func (db Database) Transact(statement string, params ...interface{}) error {
	if !db.IsOpen() {
		return errors.New("Database not loaded.")
	}
	_, err := db.database.Exec(statement, params)
	if err != nil {
		return err
	}
	return nil
}

func (db Database) Query(query string) (*sql.Rows, error) {
	if db.database == nil {
		return nil, errors.New("Database not loaded.")
	}
	return db.database.Query(query)
}
