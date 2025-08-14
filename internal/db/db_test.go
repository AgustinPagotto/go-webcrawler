package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		t.Fatal(err)
	}
	return db
}
func Test_InitiateDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	err := InitiateDB(db)
	if err != nil {
		t.Fatal(err)
	}
	sqlQuery := "SELECT sql FROM sqlite_master WHERE name = 'webs_crawled';"

}
