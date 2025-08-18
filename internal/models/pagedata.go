package models

import (
	"fmt"
	"strings"
)

type PageData struct {
	URL          string
	Status       int
	TextAndLinks map[string]string
}

func NewPageData(url string, status int) *PageData {
	p := PageData{URL: url, Status: status, TextAndLinks: make(map[string]string)}
	return &p
}

func (p *PageData) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s \t %d \t %d", p.URL, p.Status, len(p.TextAndLinks)))
	return b.String()
}
