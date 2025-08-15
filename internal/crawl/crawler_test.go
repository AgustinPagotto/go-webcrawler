package crawl

import (
	"testing"
)

func TestCrawlPage(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		expect_error bool
	}{
		{
			name:         "Empty Url",
			input:        "",
			expect_error: true,
		},
		{
			name:         "Invalid Url",
			input:        "hello",
			expect_error: true,
		},
		{
			name:         "Valid Url",
			input:        "https://www.google.com",
			expect_error: false,
		},
	}
	for _, testContent := range testCases {
		_, err := CrawlPage(testContent.input)
		var errorExistence bool = err != nil
		if errorExistence != testContent.expect_error {
			t.Errorf("Expected an error, got %v", err)
		}
	}
}
