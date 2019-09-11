package store

import (
	"sort"
	"time"
)

type someFeeds map[string]*feedHolder

// storeArticle is an article as stored in the store.
type storeArticle struct {
	ID    string    `json:"id"`
	Title string    `json:"title"`
	Date  time.Time `json:"date"`
	Body  string    `json:"body"`
}

type feedSorter []storeArticle

func (a feedSorter) Len() int           { return len(a) }
func (a feedSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a feedSorter) Less(i, j int) bool { return a[i].Date.Before(a[j].Date) }

type feedHolder struct {
	id       string
	articles map[string]storeArticle
}

func newFeedHolder(id string) *feedHolder {
	return &feedHolder{
		id:       id,
		articles: map[string]storeArticle{},
	}
}

func (f *feedHolder) add(a storeArticle) {
	if _, exists := f.articles[a.ID]; exists {
		return
	}
	// log.WithField("article", a).Info("Storing article")
	f.articles[a.ID] = a
}

// XXX
func (f *feedHolder) getSomeArticles() []storeArticle {
	articles := make(feedSorter, 0, len(f.articles))
	for _, a := range f.articles {
		articles = append(articles, a)
		sort.Sort(articles)
	}
	return articles
}
