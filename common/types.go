package common

import (
	"context"
	"time"
)

// Service is a simple service
type Service interface {
	Start(context.Context) error
}

// RestSource is a source as known by the config service
type RestSource struct {
	ID     string           `json:"id"`
	Spec   RestSourceSpec   `json:"spec"`
	Status RestSourceStatus `json:"status"`
}

// RestSourceSpec to put new sources into config server.
type RestSourceSpec struct {
	URL   string `json:"url"`
	Store string `json:"store"`
}

// RestSourceStatus is how a source appears to be functioning
type RestSourceStatus struct {
	LastStatus string `json:"lastStatus"`
}

// RestSources is a list of sources
type RestSources struct {
	Sources []RestSource `json:"sources"`
}

// OutputFeed is how the frontend serves data.
type OutputFeed struct {
	Query    string          `json:"query"`
	Next     string          `json:"next"`
	Articles []OutputArticle `json:"articles"`
}

// OutputArticle is how the frontend presents articles
type OutputArticle struct {
	Source string    `json:"source"`
	ID     string    `json:"id"`
	Title  string    `json:"title"`
	Date   time.Time `json:"date"`
	Body   string    `json:"body"`
}
