package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func setupConTestDB(t *testing.T) *sql.DB {
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

func setupTestDB(t *testing.T) *sql.DB {
	db := setupConTestDB(t)
	_ = InitiateDB(db)
	return db
}

func TestInitiateDB(t *testing.T) {
	db := setupConTestDB(t)
	defer db.Close()
	err := InitiateDB(db)
	if err != nil {
		t.Fatal(err)
	}
	var schema string
	sqlQuery := "SELECT sql FROM sqlite_master WHERE name = ?;"
	err = db.QueryRow(sqlQuery, "webs_crawled").Scan(&schema)
	if err != nil {
		t.Fatal(err)
	}
	sqlQuery = "SELECT sql FROM sqlite_master WHERE name = 'child_webs';"
	err = db.QueryRow(sqlQuery, "child_webs").Scan(&schema)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnterNewUrl(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	err := EnterNewUrl(db, "www.google.com", 200, 20)
	if err != nil {
		t.Fatal(err)
	}
	var id int
	sqlQuery := "SELECT DISTINCT id FROM webs_crawled WHERE url = ?;"
	err = db.QueryRow(sqlQuery, "www.google.com").Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnterNewChilds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	url := "www.google.com"
	child_webs := []string{"www.google.com/images", "www.duckduckgo.com"}
	err := EnterNewUrl(db, url, 200, 5)
	if err != nil {
		t.Fatal(err)
	}
	err = EnterNewChilds(db, url, child_webs)
	if err != nil {
		t.Fatal(err)
	}
	var id int
	sqlQuery := "SELECT DISTINCT id FROM child_webs WHERE url = ?;"
	for _, v := range child_webs {
		err = db.QueryRow(sqlQuery, v).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestWasUrlCrawled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	url := "www.google.com"
	urlNotCrawled := "www.yahoo.com"
	err := EnterNewUrl(db, url, 200, 5)
	if err != nil {
		t.Fatal(err)
	}
	res, err := WasUrlCrawled(db, url)
	if err != nil {
		t.Fatal(err)
	} else if !res {
		t.Fatal(err)
	}
	res, err = WasUrlCrawled(db, urlNotCrawled)
	if err != nil {
		t.Fatal(err)
	} else if res {
		t.Fatal(err)
	}
}
