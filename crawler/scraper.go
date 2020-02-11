package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

var urlReg = regexp.MustCompile("(href|src|action)=\"([a-zA-Z0-9:/\\-._~]+)\"")

// ScrapedEntry is an entry of data that was scraped from the site.
type ScrapedEntry struct {
	depth    uint
	url      string
	fileType string // html, txt, js, png, pdf, css
	err      error
	children []*ScrapedEntry
}

// Print prints out the tree structure of a scraped entry
func (s *ScrapedEntry) Print() {
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

// newScrapedEntry creates a new scraped entry
func newScrapedEntry(depth uint, url string, fileType string, err error, numChildren uint) *ScrapedEntry {
	ret := new(ScrapedEntry)
	ret.depth = depth
	ret.url = url
	ret.fileType = fileType
	ret.err = err
	ret.children = make([]*ScrapedEntry, numChildren)
	return ret
}

type queue struct {
	settings       *ScrapeSettings
	requestsOut    int
	requestsTicker chan int
	queue          chan func()
}

func newQueue(settings *ScrapeSettings) *queue {
	q := new(queue)
	q.requestsOut = 0
	q.requestsTicker = make(chan int)
	q.queue = make(chan func(), settings.maxQueueSize)
	q.settings = settings
	return q
}

// ScrapeSettings are the settings necessary to run a scraping operation
type ScrapeSettings struct {
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

// NewScrapeSettings creates a new scraped settings, need to do this before running a scrape operation
func NewScrapeSettings(maxDepth uint, waitTime uint, maxSimulRequests uint, maxQueueSize uint) *ScrapeSettings {
	ret := new(ScrapeSettings)
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

type scraper struct {
	q *queue
	s *ScrapeSettings
}

func newScraper(settings *ScrapeSettings) *scraper {
	ret := new(scraper)
	ret.s = settings
	ret.q = newQueue(settings)
	return ret
}

func (scraper *scraper) scrape(depth uint, url string, id int, entry *ScrapedEntry) {
	val, err := http.Get(url)
	if depth == scraper.s.maxDepth || err != nil {
		entry.children[id] = newScrapedEntry(depth, url, "NIL", err, 0)
	} else {
		fileType := getFileType(val)
		newURLs := getURLsFromResponse(val)
		newEntry := newScrapedEntry(depth, url, fileType, nil, uint(len(newURLs)))
		entry.children[id] = newEntry
		for ind, newURL := range newURLs {
			scraper.q.enqueue(func() { scraper.scrape(depth+1, newURL, ind, newEntry) })
		}
	}
}

func (scraper *scraper) run(url string, parentEntry *ScrapedEntry) {
	go scraper.q.enqueue(func() { scraper.scrape(0, url, 0, parentEntry) })
	scraper.q.monitor()
}

// Scrape is a recursive scrape function that that will scrape the site given by the url and with the settings that have been supplied.
func Scrape(url string, settings *ScrapeSettings) *ScrapedEntry {
	parentEntry := newScrapedEntry(0, "", "", nil, 1)
	newScraper(settings).run(url, parentEntry)
	return parentEntry.children[0]
}
