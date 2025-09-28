package db

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/AgustinPagotto/go-webcrawler/internal/crawl"
	_ "github.com/mattn/go-sqlite3"
)

func setupConTestStore(t *testing.T) *Store {
	t.Helper()
	db, _ := sql.Open("sqlite3", ":memory:")
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")
	return &Store{db: db}
}

func insertDataInDb(t *testing.T, s *Store) {
	t.Helper()
	url := "www.google.com"
	child_webs := map[string]string{
		"images": "www.google.com/images",
		"duck":   "www.duckduckgo.com",
	}
	urlWithoutChildren := "www.yahoo.com"
	status := 200
	crawler := crawl.New(url, 0, status, time.Now())
	crawlerWithNoChilds := crawl.New(urlWithoutChildren, 0, status, time.Now())
	crawler.TextLinksCrawled = child_webs
	err := s.EnterNewUrl(*crawler)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnterNewUrl(*crawlerWithNoChilds)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnterNewChilds(*crawler)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitiateDB(t *testing.T) {
	store := setupConTestStore(t)
	defer store.Close()
	err := store.InitiateDB()
	if err != nil {
		t.Fatal(err)
	}
	var schema string
	sqlQuery := "SELECT sql FROM sqlite_master WHERE name = ?;"
	err = store.db.QueryRow(sqlQuery, "webs_crawled").Scan(&schema)
	if err != nil {
		t.Fatal(err)
	}
	sqlQuery = "SELECT sql FROM sqlite_master WHERE name = 'child_webs';"
	err = store.db.QueryRow(sqlQuery, "child_webs").Scan(&schema)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnterNewUrl(t *testing.T) {
	store := setupConTestStore(t)
	store.InitiateDB()
	defer store.Close()
	url := "www.google.com"
	crawler := crawl.New(url, 0, 200, time.Now())
	err := store.EnterNewUrl(*crawler)
	if err != nil {
		t.Fatal(err)
	}
	var id int
	sqlQuery := "SELECT DISTINCT id FROM webs_crawled WHERE url = ?;"
	err = store.db.QueryRow(sqlQuery, "www.google.com").Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnterNewChilds(t *testing.T) {
	store := setupConTestStore(t)
	store.InitiateDB()
	defer store.Close()
	url := "www.google.com"
	child_webs := map[string]string{
		"images": "www.google.com/images",
		"duck":   "www.duckduckgo.com",
	}
	crawler := crawl.New(url, 0, 200, time.Now())
	crawler.TextLinksCrawled = child_webs
	err := store.EnterNewUrl(*crawler)
	if err != nil {
		t.Fatal(err)
	}
	err = store.EnterNewChilds(*crawler)
	if err != nil {
		t.Fatal(err)
	}
	var id int
	sqlQuery := "SELECT DISTINCT id FROM child_webs WHERE url = ?;"
	for _, v := range child_webs {
		err = store.db.QueryRow(sqlQuery, v).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestIsUrlOnDb(t *testing.T) {
	store := setupConTestStore(t)
	store.InitiateDB()
	insertDataInDb(t, store)
	defer store.Close()
	urlNotInDb := "www.reddit.com"
	urlInDb := "www.google.com"
	urlInDbWoChildren := "www.yahoo.com"
	_, err := store.IsUrlOnDb(urlNotInDb)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	crawler, err := store.IsUrlOnDb(urlInDb)
	if err != nil || len(crawler.TextLinksCrawled) == 0 {
		fmt.Print(crawler.TextLinksCrawled)
		t.Fatal(err)
	}
	crawler, err = store.IsUrlOnDb(urlInDbWoChildren)
	if err != nil || len(crawler.TextLinksCrawled) != 0 {
		t.Fatal(err)
	}
}

func TestUpdateLastCrawled(t *testing.T) {
	store := setupConTestStore(t)
	store.InitiateDB()
	now := time.Now()
	url := "www.google.com"
	pastDate := now.Add(12 * time.Hour)
	crawler := crawl.New(url, 0, 200, pastDate)
	err := store.EnterNewUrl(*crawler)
	if err != nil {
		t.Fatal(err)
	}
	var firstTimeInDb time.Time
	var updatedTimeInDb time.Time
	sqlQuery := "SELECT DISTINCT last_crawled FROM webs_crawled WHERE url = ?;"
	err = store.db.QueryRow(sqlQuery, crawler.URL).Scan(&firstTimeInDb)
	if err != nil {
		t.Fatal(err)
	}
	err = store.UpdateLastCrawled(url)
	err = store.db.QueryRow(sqlQuery, crawler.URL).Scan(&updatedTimeInDb)
	if time.Time.Equal(firstTimeInDb, updatedTimeInDb) {
		t.Fatal("the times are the same, the time wasn't updated")
	}
}

func TestFilterOldChilds(t *testing.T) {
	store := setupConTestStore(t)
	store.InitiateDB()
	insertDataInDb(t, store)
	url := "www.google.com"
	child_webs := map[string]string{
		"images":    "www.google.com/images",
		"duck":      "www.duckduckgo.com",
		"other_url": "www.google.com/map",
	}
	crawler := crawl.New(url, 0, 200, time.Now())
	crawler.TextLinksCrawled = child_webs
	err := store.FilterOldChilds(crawler)
	if err != nil {
		t.Fatal(err)
	}
}
