package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
)

var urlReg = regexp.MustCompile("(href|src|action)=\"([a-zA-Z0-9:/\\-._~]+)\"")

// CrawledEntry is an entry of data that was crawled from the site.
type CrawledEntry struct {
	depth    uint
	url      string
	fileType string // result of http.Response.Header.Get("Content-type")
	err      error
	children []*CrawledEntry
}

// Print prints out the tree structure of a crawled entry
func (s *CrawledEntry) Print() {
	if s == nil {
		fmt.Println("This entry is nil and cannot be printed.")
		return
	}
	tabs := ""
	for i := uint(0); i < s.depth; i++ {
		tabs += "\t"
	}
	if s.err != nil {
		fmt.Printf("%s%s\n", tabs, s.err.Error())
		return
	}
	fmt.Printf("%s%s\n", tabs, s.url)
	for _, child := range s.children {
		if child != nil {
			child.Print()
		}
	}
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

type queue struct {
	settings       *CrawlSettings
	requestsOut    int
	requestsTicker chan int
	queue          chan func()
}

func newQueue(settings *CrawlSettings) *queue {
	q := new(queue)
	q.requestsOut = 0
	q.requestsTicker = make(chan int)
	q.queue = make(chan func(), settings.maxQueueSize)
	q.settings = settings
	return q
}

// CrawlSettings are the settings necessary to run a scraping operation
type CrawlSettings struct {
	maxDepth         uint
	waitTime         uint
	maxSimulRequests uint
	maxQueueSize     uint
}

func (q *queue) enqueue(fun func()) {
	q.requestsTicker <- 1
	q.queue <- fun
}

func (q *queue) dequeue(fun func()) {
	fun()
	q.requestsTicker <- -1
}

func (q *queue) runTicker() {
	ticker := func() {
		for tick := range q.requestsTicker {
			q.requestsOut += tick
			if q.requestsOut == 0 {
				close(q.requestsTicker)
			}
		}
		close(q.queue)
	}
	go ticker()
}

func (q *queue) monitor() {
	q.runTicker()
	for fun := range q.queue {
		go q.dequeue(fun)
	}
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
	return resp.Header.Get("Content-type")
}

func findAll(data []byte) []string {
	all := urlReg.FindAllSubmatch(data, -1)
	ret := make([]string, len(all))
	for ind, match := range all {
		ret[ind] = string(match[2])
	}
	return ret
}

func getURLsFromResponse(resp *http.Response) []string {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return make([]string, 0)
	}
	ret := findAll(data)
	return ret
}

type cache struct {
	lock sync.Mutex
	set  map[string]bool
}

func newCache() *cache {
	ret := new(cache)
	ret.lock = sync.Mutex{}
	ret.set = make(map[string]bool)
	return ret
}

func (c *cache) add(url string) bool {
	defer c.lock.Unlock()
	c.lock.Lock()
	if c.set[url] {
		return false
	}
	c.set[url] = true
	return true
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
	val, err := http.Get(url)
	if err != nil {
		entry.children[id] = newCrawledEntry(depth, url, "?", err, 0)
	} else {
		fileType := getFileType(val)
		if depth == crawler.s.maxDepth {
			entry.children[id] = newCrawledEntry(depth, url, fileType, nil, 0)
		} else {
			newURLs := getURLsFromResponse(val)
			newEntry := newCrawledEntry(depth, url, fileType, nil, uint(len(newURLs)))
			entry.children[id] = newEntry
			for ind, newURL := range newURLs {
				if crawler.c.add(newURL) {
					crawler.q.enqueue(func() { crawler.crawl(depth+1, newURL, ind, newEntry) })
				}
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
