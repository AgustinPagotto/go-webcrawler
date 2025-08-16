package db

import (
	"database/sql"
	"testing"

	"github.com/AgustinPagotto/go-webcrawler/internal/models"
	_ "github.com/mattn/go-sqlite3"
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
	url := "www.google.com"
	pgData := models.NewPageData(url, 200)
	pgData.Status = 200
	err := EnterNewUrl(db, pgData)
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
	child_webs := map[string]string{
		"images": "www.google.com/images",
		"duck":   "www.duckduckgo.com",
	}
	pgData := models.NewPageData(url, 200)
	pgData.TextAndLinks = child_webs
	pgData.Status = 200
	err := EnterNewUrl(db, pgData)
	if err != nil {
		t.Fatal(err)
	}
	err = EnterNewChilds(db, pgData)
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
	pgData := models.NewPageData(url, 200)
	err := EnterNewUrl(db, pgData)
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
