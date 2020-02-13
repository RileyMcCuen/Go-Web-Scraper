package crawler

import (
	"sync"
)

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
