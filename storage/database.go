package storage

import (
	"database/sql"
	"errors"
	"log"
	"sync"

	//blank import
	_ "github.com/mattn/go-sqlite3"
)

// Database convinient wrapper for accessing sql DB
type Database struct {
	database *sql.DB
	mutex    *sync.Mutex
}

// OpenDb opens a specified database
func (db *Database) OpenDb(path string) {
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	database.Exec("PRAGMA journal_mode=WAL;")

	db.database = database
	db.mutex = &sync.Mutex{}
}

// Close closes the database connection
func (db *Database) Close() {
	db.mutex.Lock()
	db.database.Close()
	db.mutex.Unlock()
}

// IsOpen returns true if the database connection has been established
func (db *Database) IsOpen() bool {
	return db.database != nil
}

// Transact performs a transaction on the database
func (db *Database) Transact(statement string, params ...interface{}) error {
	if !db.IsOpen() {
		return errors.New("database not loaded")
	}
	db.mutex.Lock()
	_, err := db.database.Exec(statement, params...)
	db.mutex.Unlock()
	return err
}

// Query performs a query on the database
func (db *Database) Query(query string, params ...interface{}) (*sql.Rows, error) {
	if db.database == nil {
		return nil, errors.New("database not loaded")
	}
	db.mutex.Lock()
	rows, err := db.database.Query(query, params...)
	db.mutex.Unlock()
	return rows, err
}
