package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func openConToDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		return nil, fmt.Errorf("Error trying to connect to the db: \n%v", err)
	}
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")
	return db, nil
}

func InitiateDB() error {
	db, err := openConToDB()
	if err != nil {
		return err
	}
	defer db.Close()
	sqlQuery := `
	CREATE TABLE IF NOT EXISTS webs_crawled (
		id INTEGER NOT NULL PRIMARY KEY,
		url TEXT,
		status INTEGER,
		childs INTEGER,
		last_crawled DATETIME
	);
	`
	_, err = db.Exec(sqlQuery)
	if err != nil {
		return fmt.Errorf("Error trying to create webs_crawled table: \n%v", err)
	}
	sqlQuery = `
	CREATE TABLE IF NOT EXISTS child_webs(
		id INTEGER NOT NULL PRIMARY KEY,
		web_crawled_id INTEGER NOT NULL,
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

func EnterNewUrl(url string, status int, childs int) error {
	db, err := openConToDB()
	if err != nil {
		return err
	}
	defer db.Close()
	crawlDateTime := time.Now()
	sqlQuery := "INSERT INTO webs_crawled (url, status, childs, last_crawled) VALUES (?,?,?,?);"
	_, err = db.Exec(sqlQuery, url, status, childs, crawlDateTime)
	if err != nil {
		return fmt.Errorf("Error trying to insert new url: \n%v", err)
	}
	return nil
}

func EnterNewChilds(url string, child_urls []string) error {
	db, err := openConToDB()
	if err != nil {
		return err
	}
	defer db.Close()
	var id int
	sqlQuery := "SELECT DISTINCT id FROM webs_crawled WHERE url = ?;"
	err = db.QueryRow(sqlQuery, url).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("didn't find the url to put child into: \n%v", err)
	} else if err != nil {
		return fmt.Errorf("db query failed: %w", err)
	}
	fmt.Printf("We found a match for the url! the id of the site is %d", id)
	sqlQuery = "INSERT INTO child_webs (web_crawled_id, url) VALUES (?,?);"
	for _, v := range child_urls {
		_, err = db.Exec(sqlQuery, id, v)
		if err != nil {
			return fmt.Errorf("couldn't insert the url: \n%v", err)
		}
	}
	return nil
}

func WasUrlCrawled(url string) (bool, error) {
	db, err := openConToDB()
	if err != nil {
		return false, err
	}
	defer db.Close()
	var id int
	sqlQuery := "SELECT DISTINCT id FROM webs_crawled WHERE url = ?;"
	err = db.QueryRow(sqlQuery, url).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("didn't find the url: \n%v", err)
	} else if err != nil {
		return false, fmt.Errorf("db query failed: %w", err)
	}
	return true, nil
}
