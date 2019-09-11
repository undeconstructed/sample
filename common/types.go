package common

import (
	"context"
	"time"
)

// Service is a simple service
type Service interface {
	Start(context.Context) error
}

// SourceConfig to put new sources into config server.
type SourceConfig struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Store string `json:"store"`
}

// SourcesConfig is a list of sources
type SourcesConfig struct {
	Sources []SourceConfig `json:"sources"`
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
