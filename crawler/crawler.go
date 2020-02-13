package crawler

import (
	"fmt"
	"net/http"
)

// CrawledEntry is an entry of data that was crawled from the site.
type CrawledEntry struct {
	depth    uint
	url      string
	fileType string // result of http.Response.Header.Get("Content-type")
	err      error
	children []*CrawledEntry
}

func (s *CrawledEntry) print(tabs string) {
	if s == nil {
		fmt.Println("This entry is nil and cannot be printed.")
		return
	}
	if s.err != nil {
		fmt.Printf("%s%s\n", tabs, s.err.Error())
		return
	}
	fmt.Printf("%s%s\n", tabs, s.url)
	nextTabs := tabs + "\t"
	for _, child := range s.children {
		if child != nil {
			child.print(nextTabs)
		}
	}
}

// Print prints out the tree structure of a crawled entry
func (s *CrawledEntry) Print() {
	s.print("")
}

// newCrawledEntry creates a new crawled entry
func newCrawledEntry(depth uint, url string, fileType string, err error, numChildren uint) *CrawledEntry {
	ret := new(CrawledEntry)
	ret.depth = depth
	ret.url = url
	ret.fileType = fileType
	ret.err = err
	ret.children = make([]*CrawledEntry, numChildren)
	return ret
}

// NewCrawlSettings creates a new crawled settings, need to do this before running a crawl operation
func NewCrawlSettings(maxDepth uint, waitTime uint, maxSimulRequests uint, maxQueueSize uint) *CrawlSettings {
	ret := new(CrawlSettings)
	ret.maxDepth = maxDepth
	ret.waitTime = waitTime
	ret.maxSimulRequests = maxSimulRequests
	ret.maxQueueSize = maxQueueSize
	return ret
}

func getFileType(resp *http.Response) string {
	if resp == nil {
		return "?"
	}
	return resp.Header.Get("Content-type")
}

func getResponseAndFileType(url string) (*http.Response, string, error) {
	val, err := http.Get(url)
	fileType := getFileType(val)
	return val, fileType, err
}

type crawler struct {
	c *cache
	q *queue
	s *CrawlSettings
}

func newCrawler(settings *CrawlSettings) *crawler {
	ret := new(crawler)
	ret.s = settings
	ret.q = newQueue(settings)
	ret.c = newCache()
	return ret
}

func (crawler *crawler) crawl(depth uint, url string, id int, entry *CrawledEntry) {
	val, fileType, err := getResponseAndFileType(url)
	curEntry := newCrawledEntry(depth, url, fileType, err, 0)
	entry.children[id] = curEntry
	if err == nil && depth != crawler.s.maxDepth {
		newURLs, err := getURLsFromResponse(val)
		if err != nil {
			return
		}
		curEntry.children = make([]*CrawledEntry, uint(len(newURLs)))
		for ind, newURL := range newURLs {
			if crawler.c.add(newURL.String()) {
				crawler.q.enqueue(func() { crawler.crawl(depth+1, newURL.String(), ind, curEntry) })
			}
		}
	}
}

func (crawler *crawler) run(url string, parentEntry *CrawledEntry) {
	go crawler.q.enqueue(func() { crawler.crawl(0, url, 0, parentEntry) })
	crawler.q.monitor()
}

// Crawl is a recursive crawl function that will crawl the site given by the url and with the settings that have been supplied.
func Crawl(url string, settings *CrawlSettings) *CrawledEntry {
	parentEntry := newCrawledEntry(0, "", "", nil, 1)
	newCrawler(settings).run(url, parentEntry)
	return parentEntry.children[0]
}
