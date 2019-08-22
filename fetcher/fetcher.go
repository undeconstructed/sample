package fetcher

import (
	"context"
	"fmt"
	"time"

	resty "github.com/go-resty/resty/v2"

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
}

func (a *fetcher) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.stop = cancel

	go func() {
		for {
			a.doFetch()
			t := time.After(10 * time.Second)
			select {
			case <-t:
				continue
			case <-ctx.Done():
				fmt.Println("Fetcher stopping")
				break
			}
		}
	}()

	return nil
}

func (a *fetcher) doFetch() {
	work := common.FetchWork{}

	client := resty.New()
	_, err := client.R().
		SetResult(&work).
		Get("http://" + a.configURL + "/work")

	if err != nil {
		fmt.Printf("Error fetching work list: %v\n", err)
		return
	}

	for _, j := range work.Jobs {
		fmt.Printf("Fetching %s\n", j.URL)
	}
}

func (a *fetcher) Stop() {
	a.stop()
}
