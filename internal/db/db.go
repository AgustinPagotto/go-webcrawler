package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AgustinPagotto/go-webcrawler/internal/crawl"
	_ "github.com/mattn/go-sqlite3"
)

func OpenConToDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./crawl.db")
	if err != nil {
		return nil, fmt.Errorf("Error trying to connect to the db: \n%v", err)
	}
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")
	return db, nil
}

func InitiateDB(db *sql.DB) error {
	sqlQuery := `
	CREATE TABLE IF NOT EXISTS webs_crawled (
		id INTEGER NOT NULL PRIMARY KEY,
		url TEXT,
		status INTEGER,
		last_crawled DATETIME
	);
	`
	_, err := db.Exec(sqlQuery)
	if err != nil {
		return fmt.Errorf("Error trying to create webs_crawled table: \n%v", err)
	}
	sqlQuery = `
	CREATE TABLE IF NOT EXISTS child_webs(
		id INTEGER NOT NULL PRIMARY KEY,
		web_crawled_id INTEGER NOT NULL,
		url_text TEXT,
		url TEXT,
		FOREIGN KEY (web_crawled_id) REFERENCES webs_crawled(id) ON DELETE CASCADE
	);
	`
	_, err = db.Exec(sqlQuery)
	if err != nil {
		return fmt.Errorf("Error trying to create child_webs table: \n%v", err)
	}
	sqlQuery = `CREATE INDEX IF NOT EXISTS idx_child_webs_url_and_text ON child_webs(url_text, url);`
	_, err = db.Exec(sqlQuery)
	if err != nil {
		return fmt.Errorf("Error trying to create index of url from child_webs table: \n%v", err)
	}
	return nil
}

func EnterNewUrl(db *sql.DB, crawler crawl.Crawler) error {
	sqlQuery := "INSERT INTO webs_crawled (url, status, last_crawled) VALUES (?,?,?);"
	_, err := db.Exec(sqlQuery, crawler.URL, crawler.Status, crawler.LastTimeCrawled)
	if err != nil {
		return fmt.Errorf("Error trying to insert new url: \n%v", err)
	}
	return nil
}

func EnterNewChilds(db *sql.DB, crawler crawl.Crawler) error {
	var id int
	sqlQuery := "SELECT DISTINCT id FROM webs_crawled WHERE url = ?;"
	err := db.QueryRow(sqlQuery, crawler.URL).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("didn't find the url to put child into: \n%v", err)
	} else if err != nil {
		return fmt.Errorf("db query failed: %w", err)
	}
	sqlQuery = "INSERT INTO child_webs (web_crawled_id, url_text, url) VALUES (?,?,?);"
	for url_text, url := range crawler.TextLinksCrawled {
		_, err = db.Exec(sqlQuery, id, url_text, url)
		if err != nil {
			return fmt.Errorf("couldn't insert the url: \n%v", err)
		}
	}
	return nil
}

func IsUrlOnDb(db *sql.DB, url string) (*crawl.Crawler, error) {
	var id, status int
	var timeCrawled time.Time
	sqlQuery := "SELECT DISTINCT id, status, last_crawled FROM webs_crawled WHERE url = ?;"
	err := db.QueryRow(sqlQuery, url).Scan(&id, &status, &timeCrawled)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("didn't find the url in the db: \n%w", err)
	} else if err != nil {
		return nil, fmt.Errorf("consult of url in db query failed: %w", err)
	}
	crawler := crawl.New(url, status, timeCrawled)
	sqlQuery = "SELECT DISTINCT url_text, url FROM child_webs WHERE web_crawled_id = ?;"
	rows, err := db.Query(sqlQuery, id)
	if err != nil {
		return nil, fmt.Errorf("consult of child urls in db query failed: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var urlText, urlLink string
		if err := rows.Scan(&urlText, &urlLink); err != nil {
			return crawler, err
		}
		crawler.TextLinksCrawled[urlText] = urlLink
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return crawler, nil
}

func UpdateLastCrawled(db *sql.DB, url string) error {
	sqlQuery := "UPDATE webs_crawled set last_crawled = ? WHERE url = ?;"
	_, err := db.Exec(sqlQuery, time.Now(), url)
	if err != nil {
		return err
	}
	return nil
}

func FilterOldChilds(db *sql.DB, crawler *crawl.Crawler) error {
	var id, cont int
	sqlQuery := "SELECT id FROM webs_crawled WHERE url = ?;"
	err := db.QueryRow(sqlQuery, crawler.URL).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("didn't find the url in the db: \n%w", err)
	} else if err != nil {
		return fmt.Errorf("consult of url in db query failed: %w", err)
	}
	sqlQuery = "SELECT DISTINCT url_text, url FROM child_webs WHERE web_crawled_id = ?;"
	rows, err := db.Query(sqlQuery, id)
	if err != nil {
		return fmt.Errorf("consult of child urls in db query failed: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var urlText, urlLink string
		if err := rows.Scan(&urlText, &urlLink); err != nil {
			return err
		}
		if val, ok := crawler.TextLinksCrawled[urlText]; ok && val == urlLink {
			cont = cont + 1
			delete(crawler.TextLinksCrawled, urlText)
		}
	}
	if err = rows.Err(); err != nil {
		return err
	}
	fmt.Printf("Deleted %d already saved links", cont)
	return nil
}

func SearchTerm(db *sql.DB, searchTerm string) ([]string, error) {
	cutSearchTerm := firstN(searchTerm, 3)
	var urlsFound []string
	sqlQuery := "SELECT url FROM child_webs WHERE url_text LIKE ?;"
	newSearchTerm := fmt.Sprintf("%%%s%%", cutSearchTerm)
	rows, err := db.Query(sqlQuery, newSearchTerm)
	if err != nil {
		return nil, fmt.Errorf("there was an error trying to search that term: %v ", err)
	}
	for rows.Next() {
		var urlLink string
		if err := rows.Scan(&urlLink); err != nil {
			return nil, err
		}
		urlsFound = append(urlsFound, urlLink)
	}
	defer rows.Close()
	return urlsFound, nil
}

func firstN(str string, n int) string {
	runes := []rune(str)
	if n >= len(runes) {
		return str
	}
	return string(runes[:n])
}
