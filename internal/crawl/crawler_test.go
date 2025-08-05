package crawl

import "testing"

func Test_CrawlPageEmptyUrl(t *testing.T) {
	if _, err := CrawlPage(""); err == nil {
		t.Errorf("Expected an error, got %v", err)
	}
}

func Test_CrawlPageInvalidUrl(t *testing.T) {
	if _, err := CrawlPage("hello"); err == nil {
		t.Errorf("Expected an error, got %v", err)
	}
}

func Test_CrawlPageValidUrl(t *testing.T) {
	result, err := CrawlPage("https://www.google.com")
	if err != nil {
		t.Errorf("Not expected error, got %v", err)
	}
	if result == nil {
		t.Error("The result returned shouldn't be empty")
	}
}
