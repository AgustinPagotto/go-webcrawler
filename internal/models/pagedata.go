package models

import (
	"fmt"
	"strings"
	"time"
)

type PageData struct {
	URL          string
	Status       int
	TextAndLinks map[string]string
	LastCrawled  time.Time
}

func NewPageData(url string, status int, timeCrawled time.Time) *PageData {
	p := PageData{URL: url, Status: status, TextAndLinks: make(map[string]string), LastCrawled: timeCrawled}
	return &p
}

func (p *PageData) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s \t %d \t %d \t %s", p.URL, p.Status, len(p.TextAndLinks), p.LastCrawled.String()))
	return b.String()
}
