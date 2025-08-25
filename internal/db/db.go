package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/AgustinPagotto/go-webcrawler/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"time"
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
	return nil
}

func EnterNewUrl(db *sql.DB, pgData *models.PageData) error {
	sqlQuery := "INSERT INTO webs_crawled (url, status, last_crawled) VALUES (?,?,?);"
	_, err := db.Exec(sqlQuery, pgData.URL, pgData.Status, pgData.LastCrawled)
	if err != nil {
		return fmt.Errorf("Error trying to insert new url: \n%v", err)
	}
	return nil
}

func EnterNewChilds(db *sql.DB, pgData *models.PageData) error {
	var id int
	sqlQuery := "SELECT DISTINCT id FROM webs_crawled WHERE url = ?;"
	err := db.QueryRow(sqlQuery, pgData.URL).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("didn't find the url to put child into: \n%v", err)
	} else if err != nil {
		return fmt.Errorf("db query failed: %w", err)
	}
	sqlQuery = "INSERT INTO child_webs (web_crawled_id, url_text, url) VALUES (?,?,?);"
	for url_text, url := range pgData.TextAndLinks {
		_, err = db.Exec(sqlQuery, id, url_text, url)
		if err != nil {
			return fmt.Errorf("couldn't insert the url: \n%v", err)
		}
	}
	return nil
}

func IsUrlOnDb(db *sql.DB, url string) (*models.PageData, error) {
	var id, status int
	var timeCrawled time.Time
	sqlQuery := "SELECT DISTINCT id, status, last_crawled FROM webs_crawled WHERE url = ?;"
	err := db.QueryRow(sqlQuery, url).Scan(&id, &status, &timeCrawled)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("didn't find the url in the db: \n%w", err)
	} else if err != nil {
		return nil, fmt.Errorf("consult of url in db query failed: %w", err)
	}
	pgData := models.NewPageData(url, status, timeCrawled)
	sqlQuery = "SELECT DISTINCT url_text, url FROM child_webs WHERE web_crawled_id = ?;"
	rows, err := db.Query(sqlQuery, id)
	if err != nil {
		return nil, fmt.Errorf("consult of child urls in db query failed: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var urlText, urlLink string
		if err := rows.Scan(&urlText, &urlLink); err != nil {
			return pgData, err
		}
		pgData.TextAndLinks[urlText] = urlLink
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return pgData, nil
}

func UpdateLastCrawled(db *sql.DB, url string) error {
	sqlQuery := "UPDATE webs_crawled set last_crawled = ? WHERE url = ?;"
	_, err := db.Exec(sqlQuery, time.Now(), url)
	if err != nil {
		return err
	}
	return nil
}

func FilterOldChilds(db *sql.DB, pgData *models.PageData) error {
	var id, cont int
	sqlQuery := "SELECT id FROM webs_crawled WHERE url = ?;"
	err := db.QueryRow(sqlQuery, pgData.URL).Scan(&id)
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
		if val, ok := pgData.TextAndLinks[urlText]; ok && val == urlLink {
			cont = cont + 1
			delete(pgData.TextAndLinks, urlText)
		}
	}
	if err = rows.Err(); err != nil {
		return err
	}
	fmt.Printf("Deleted %d already saved links", cont)
	return nil
}
