package main

import (
	"flag"
	"fmt"
	"go_web_scraper/crawler"
	"os"
)

func main() {
	url := flag.String("url", "", "Enter the url for entry point of the scraping operation. (Required)")
	depth := flag.Uint("depth", 1, "The maximum depth to which you want to scrape.")
	flag.Parse()
	if *url == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	fmt.Println("Starting...")
	settings := crawler.NewScrapeSettings(*depth, 1, 1, 1)
	fmt.Println("Beginning Scraping...")
	ret := crawler.Scrape(*url, settings)
	fmt.Println("Done!\nHere are your results:")
	fmt.Println("")
	ret.Print()
	fmt.Println("")
}
