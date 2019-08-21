package common

// Service is a simple service
type Service interface {
	Start() error
	Stop()
}

// SourceConfig to put new sources into config server.
type SourceConfig struct {
	URL string `json:"url"`
}

// FetchJob tells a fetcher to do something.
type FetchJob struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Store string `json:"store"`
}

// FetchWork tells a fetcher all its jobs.
type FetchWork struct {
	Jobs []FetchJob `json:"jobs"`
}

// StoreFeed is a feed as store in the store.
type StoreFeed struct {
	ID       string         `json:"id"`
	Articles []StoreArticle `json:"articles"`
}

// StoreArticle is an article as stored in the store.
type StoreArticle struct {
	ID   string `json:"id"`
	Date string `json:"date"`
	Body string `json:"body"`
}

// InputFeed is for putting feed data into the store.
type InputFeed struct {
	Articles []StoreArticle `json:"articles"`
}

// OutputFeed is how the frontend serves data.
type OutputFeed struct {
	Query    string         `json:"query"`
	Next     string         `json:"next"`
	Articles []StoreArticle `json:"articles"`
}
