package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "./studio.db"
var DB *sql.DB

func Init() error {
	var err error
	DB, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}
	
	statement, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS recordings (id INTEGER PRIMARY KEY, filename TEXT NOT NULL UNIQUE, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, name TEXT, notes TEXT, genre TEXT, protected BOOLEAN DEFAULT FALSE);`)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func AddRecording(filename string) (int64, error) {
	stmt, err := DB.Prepare("INSERT INTO recordings(filename, name) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(filename, filename)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

