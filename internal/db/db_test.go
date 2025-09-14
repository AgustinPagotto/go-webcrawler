package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/AgustinPagotto/go-webcrawler/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"testing"
	"time"
)

func setupConTestDB(t *testing.T) *sql.DB {
	t.Helper()
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
	t.Helper()
	db := setupConTestDB(t)
	_ = InitiateDB(db)
	return db
}

func insertDataInDb(t *testing.T, db *sql.DB) *sql.DB {
	t.Helper()
	url := "www.google.com"
	child_webs := map[string]string{
		"images": "www.google.com/images",
		"duck":   "www.duckduckgo.com",
	}
	urlWithoutChildren := "www.yahoo.com"
	status := 200
	pgData := models.NewPageData(url, status, time.Now())
	pgDataNoChilds := models.NewPageData(urlWithoutChildren, status, time.Now())
	pgData.TextAndLinks = child_webs
	err := EnterNewUrl(db, pgData)
	if err != nil {
		t.Fatal(err)
	}
	err = EnterNewUrl(db, pgDataNoChilds)
	if err != nil {
		t.Fatal(err)
	}
	err = EnterNewChilds(db, pgData)
	if err != nil {
		t.Fatal(err)
	}
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
	pgData := models.NewPageData(url, 200, time.Now())
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
	pgData := models.NewPageData(url, 200, time.Now())
	pgData.TextAndLinks = child_webs
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

func TestIsUrlOnDb(t *testing.T) {
	db := setupTestDB(t)
	db = insertDataInDb(t, db)
	defer db.Close()
	urlNotInDb := "www.reddit.com"
	urlInDb := "www.google.com"
	urlInDbWoChildren := "www.yahoo.com"
	pgData, err := IsUrlOnDb(db, urlNotInDb)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	pgData, err = IsUrlOnDb(db, urlInDb)
	if err != nil || len(pgData.TextAndLinks) == 0 {
		fmt.Print(pgData.TextAndLinks)
		t.Fatal(err)
	}
	pgData, err = IsUrlOnDb(db, urlInDbWoChildren)
	if err != nil || len(pgData.TextAndLinks) != 0 {
		t.Fatal(err)
	}
}

func TestUpdateLastCrawled(t *testing.T) {
	db := setupTestDB(t)
	now := time.Now()
	url := "www.google.com"
	pastDate := now.Add(12 * time.Hour)
	pgData := models.NewPageData(url, 200, pastDate)
	err := EnterNewUrl(db, pgData)
	if err != nil {
		t.Fatal(err)
	}
	var firstTimeInDb time.Time
	var updatedTimeInDb time.Time
	sqlQuery := "SELECT DISTINCT last_crawled FROM webs_crawled WHERE url = ?;"
	err = db.QueryRow(sqlQuery, pgData.URL).Scan(&firstTimeInDb)
	if err != nil {
		t.Fatal(err)
	}
	err = UpdateLastCrawled(db, url)
	err = db.QueryRow(sqlQuery, pgData.URL).Scan(&updatedTimeInDb)
	if time.Time.Equal(firstTimeInDb, updatedTimeInDb) {
		t.Fatal("the times are the same, the time wasn't updated")
	}
}

func TestFilterOldChilds(t *testing.T) {
	db := setupTestDB(t)
	db = insertDataInDb(t, db)
	url := "www.google.com"
	child_webs := map[string]string{
		"images":    "www.google.com/images",
		"duck":      "www.duckduckgo.com",
		"other_url": "www.google.com/map",
	}
	pgData := models.NewPageData(url, 200, time.Now())
	pgData.TextAndLinks = child_webs
	err := FilterOldChilds(db, pgData)
	if err != nil {
		t.Fatal(err)
	}
}
