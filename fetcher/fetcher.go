package fetcher

import (
	"context"
	"fmt"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/mmcdole/gofeed"

	"github.com/undeconstructed/sample/common"
)

type Fetcher interface {
	common.Service
}

func New(configURL string) Fetcher {
	return &fetcher{configURL: configURL}
}

type fetcher struct {
	configURL string
	stop      context.CancelFunc
	client    *resty.Client
}

func (a *fetcher) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.stop = cancel
	a.client = resty.New()

	go func() {
		for {
			a.doFetch()
			t := time.After(10 * time.Second)
			select {
			case <-t:
				// continue
				// only one fetch just now
				return
			case <-ctx.Done():
				fmt.Println("Fetcher stopping")
				return
			}
		}
	}()

	return nil
}

func (a *fetcher) doFetch() {
	work := common.FetchWork{}

	_, err := a.client.R().
		SetResult(&work).
		Get("http://" + a.configURL + "/work")

	if err != nil {
		fmt.Printf("Error fetching work list: %v\n", err)
		return
	}

	for _, job := range work.Jobs {
		// fmt.Printf("Fetching %s\n", job.URL)
		a.fetchFeed(job)
	}
}

func (a *fetcher) fetchFeed(job common.FetchJob) {
	// TODO - etag or similar
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(job.URL)

	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", job.URL, err)
		return
	}

	articles := []common.StoreArticle{}
	for _, item := range feed.Items {
		articles = append(articles, common.StoreArticle{
			ID:    item.GUID,
			Title: item.Title,
			Date:  *item.PublishedParsed,
			Body:  item.Content,
		})
	}

	toPush := common.InputFeed{
		Articles: articles,
	}

	_, err = a.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(toPush).
		Post("http://" + job.Store + "/feeds/" + job.ID)

	if err != nil {
		fmt.Printf("Error storing %s: %v\n", job.URL, err)
		return
	}

	// TODO - ack job done
}

func (a *fetcher) Stop() {
	a.stop()
}
