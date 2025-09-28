package crawl

import (
	"testing"
	"time"
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
			crawler := New(tc.input, tc.depth, 0, time.Now())
			err := crawler.Crawl()
			var errorExistence bool = err != nil
			if errorExistence != tc.expect_error {
				t.Errorf("Expected an error, got %v", err)
			}
			if crawler.TextLinksCrawled != nil {
				if tc.expect_to_crawl_page && len(crawler.TextLinksCrawled) < 1 {
					t.Errorf("Expected to have crawled info %v", err)
				}
			}

		})
	}
}

var testLinks = []string{"https://httpbin.org/", "https://wikipedia.com", "https://go.dev/"}

func BenchmarkCrawlPipelineApproach(b *testing.B) {
	for b.Loop() {
		concurrentCrawl(testLinks)
	}
}
