package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
)

var wg sync.WaitGroup

// SitemapIndex ...
type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

// News ...
type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
	Dates     []string `xml:"url>news>publication_date"`
}

// NewsMap ...
type NewsMap struct {
	Title    string
	Keyword  string
	Location string
	Date     string
}

// NewsPage ...
type NewsPage struct {
	Title string
	News  map[int]NewsMap
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Whoa, Go is neat!</h1>")
	fmt.Fprintf(w, "Whoa, Go is neat!")
	fmt.Fprintf(w, "Whoa, Go is neat!")
}

func newsRoutine(c chan News, url string) {
	var allNews News
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &allNews)
	resp.Body.Close()

	c <- allNews
	defer wg.Done()
}

func newsPageHandler(w http.ResponseWriter, r *http.Request) {

	// Get all Locations
	var s SitemapIndex
	resp, _ := http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)
	resp.Body.Close()

	// Push News from Loc to Channel
	myChannel := make(chan News, 30)
	for _, Location := range s.Locations {
		wg.Add(1)
		go newsRoutine(myChannel, Location)
	}
	wg.Wait()
	close(myChannel)

	// Take Each item from channel and append to map
	newsMap := make(map[int]NewsMap)
	for item := range myChannel {
		for i := range item.Titles {
			newsMap[i] = NewsMap{item.Titles[i], item.Keywords[i], item.Locations[i], item.Dates[i]}
		}
	}

	page := NewsPage{"WashingtonPost News", newsMap}
	layout, _ := template.ParseFiles("news.html")
	layout.Execute(w, page)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/news", newsPageHandler)
	http.ListenAndServe(":8000", nil)
}
