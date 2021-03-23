package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

const (
	topstories = "https://hacker-news.firebaseio.com/v0/topstories.json"
	itemUrl    = "https://hacker-news.firebaseio.com/v0/item/"
)

// Item -- single article or comment.
type Item struct {
	By          string `json:"by"`
	Descendants int    `json:"descendants"`
	ID          int    `json:"id"`
	Kids        []int  `json:"kids"`
	Score       int    `json:"score"`
	Time        int    `json:"time"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Content     []byte
}

type Items struct {
	Items []Item
}

type StoriesInt []int

// Stories
type Stories struct {
	Items []Item
}

func getItem(id int) string {
	itemAdress := fmt.Sprintf("%s%d.json", itemUrl, id)
	return itemAdress
}

func (item *Item) getJson(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Panic("cannot fetch URL %q: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Panic("unexpected http GET status: %s", resp.Status)
	}
	// We could check the resulting content type
	// here if desired.
	err = json.NewDecoder(resp.Body).Decode(&item)
	if err != nil {
		log.Panic("cannot decode JSON: %v", err)
	}
	if len(item.URL) > 0 {
		item.fetchUrl()
	}
}

func (stories *StoriesInt) getJson(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Panic("cannot fetch URL ", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Panic("unexpected http GET status: %s", resp.Status)
	}
	// We could check the resulting content type
	// here if desired.
	err = json.NewDecoder(resp.Body).Decode(&stories)
	if err != nil {
		log.Panic("cannot decode JSON: %v", err)
	}
}

func fetchItem(id int, itemsChan chan Item, wg *sync.WaitGroup) {
	defer wg.Done()
	var item Item
	url := getItem(id)
	item.getJson(url)
	itemsChan <- item
}

func (item *Item) fetchUrl() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", item.URL, nil)
	req.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("cannot fetch: ", item.Title, err)
		item.Type = "None"
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("from page: %s", item.URL)
		log.Println("unexpected http GET status: %s", resp.Status)
		item.Type = "None"
	}

	webpage, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic("Failed to readall response body")
	}

	item.Content = webpage
}

func (item Item) saveContent() {
	err := ioutil.WriteFile("stories/"+strconv.Itoa(item.ID)+".html", item.Content, 0644)
	if err != nil {
		log.Panic("Save failed:")
		log.Panic(err)
	}
}

func monitor(wg *sync.WaitGroup, itemsChan chan Item) {
	wg.Wait()
	close(itemsChan)
}

func (stories StoriesInt) FetchAll() Items {
	itemsChan := make(chan Item)
	var items []Item
	wg := sync.WaitGroup{}
	wg.Add(len(stories))
	for _, id := range stories {
		go fetchItem(id, itemsChan, &wg)
	}
	go monitor(&wg, itemsChan)

	for item := range itemsChan {
		if item.Type != "None" {
			item.saveContent()
			items = append(items, item)
		}
	}

	return Items{items}
}

// FetchTop hacker news stories
func FetchTop() Items {
	var stories StoriesInt
	stories.getJson(topstories)
	return stories.FetchAll()
}
