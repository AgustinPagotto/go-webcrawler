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
	if len(p.TextAndLinks) == 0 {
		return ""
	}
	var b strings.Builder
	for k, v := range p.TextAndLinks {
		b.WriteString(fmt.Sprintf("%s \t \t %s\n", k, v))
	}
	return b.String()
}
