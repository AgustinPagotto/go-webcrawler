package crawl

import (
	"testing"
)

func TestCrawlPage(t *testing.T) {
	testCases := []struct {
		name                 string
		input                string
		depth                int
		expect_error         bool
		expect_to_crawl_page bool
	}{
		{
			name:                 "Empty Url",
			input:                "",
			depth:                1,
			expect_error:         true,
			expect_to_crawl_page: false,
		},
		{
			name:                 "Invalid Url",
			input:                "hello",
			depth:                1,
			expect_error:         true,
			expect_to_crawl_page: false,
		},
		{
			name:                 "Valid Url",
			input:                "https://www.google.com",
			depth:                1,
			expect_error:         false,
			expect_to_crawl_page: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pageData, err := CrawlPage(tc.input, 1)
			var errorExistence bool = err != nil
			if errorExistence != tc.expect_error {
				t.Errorf("Expected an error, got %v", err)
			}
			if pageData != nil {
				if tc.expect_to_crawl_page && len(pageData.TextAndLinks) < 1 {
					t.Errorf("Expected to have crawled info %v", err)
				}
			}

		})
	}
}

var testLinks = []string{"https://httpbin.org/", "https://wikipedia.com", "https://go.dev/"}

func BenchmarkCrawlOnePerLink(b *testing.B) {
	for b.Loop() {
		ConcurrentCrawl(testLinks)
	}
}
func BenchmarkCrawlPoolOfFive(b *testing.B) {
	for b.Loop() {
		ConcurrentCrawlNew(testLinks)
	}
}
