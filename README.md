## About go_webcrawler

go_webcraler is a personal and learn focused project, it's objective is to allow me to learn the basics of the GO language and other key concepts:
* How to build a CLI tool.
* How to structure a GO project.
* How to work with raw HTML data.
* How to work with bare SQL Querys.
* How to use the standard library of GO.
* Understand and implement concurrency.
* How to test and benchmark in GO.

### Built With

[![Go][Go]][Go-url]
[![sqlite][sqlite]][sqlite-url]

## Features

- Crawl websites starting from a seed URL
- Store crawled data in SQLite
- Search through indexed URLs and text
- Fast prefix/substring search support
- Simple CLI interface


## Installation

Make sure you have **Go 1.22+** installed.

```bash
git clone https://github.com/yourusername/confluai-crawler.git
cd confluai-crawler
go build -o crawler .
  ```
## Usage

Basic command sintax
```bash
./crawler [flags] [url/search]
  ```
Available Flags

* -url/-u – Will let you set the seed http url.
* -d/-depth – Will let you set the depth of the crawl.
* -s/-search – Will let you perform a search in the alrady stablished db.

### Examples

Crawl
```bash
./crawler -u https://google.com -d 3
  ```

Search
```bash
./crawler -s google
  ```

## Benchmarking differents crawling techniques

### CPU DATA
goos: linux
goarch: amd64
pkg: github.com/AgustinPagotto/go-webcrawler/internal/crawl
cpu: 12th Gen Intel(R) Core(TM) i5-1240P

### Crawl with goroutine/pool and pipeline with fan-out and fan-in approach with the number of goroutines set in runtime
BenchmarkCrawlPipelineApproach-16 &emsp; 2 &emsp; 581732292 ns/op &emsp; 348468 B/op &emsp; 3568 allocs/op
### Crawl with goroutine/pool and a constant number of goroutines (5)
BenchmarkCrawlPoolOfFive-16 &emsp; 1 &emsp; 1852374970 ns/op &emsp; 362872 B/op &emsp; 3489 allocs/op
### Crawl with a goroutine per url
BenchmarkCrawlOnePerLink-16 &emsp; 1 &emsp; 1693384338 ns/op &emsp; 1168136 B/op &emsp; 7754 allocs/op


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[Go-url]: https://go.dev/
[Go]: https://img.shields.io/badge/golang-00ADD8?&style=plastic&logo=go&logoColor=white
[sqlite-url]: https://sqlite.org/
[sqlite]: https://img.shields.io/badge/SQLite-07405E?style=flat&compact=true&logo=sqlite&logoColor=white

