package fetcher

import (
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
}

func (a *fetcher) Start() error {
	work := common.FetchWork{}

	client := resty.New()
	resp, err := client.R().EnableTrace().
		SetResult(&work).
		Get("http://" + a.configURL + "/work")

	if err != nil {
		return err
	}

	fmt.Printf("work: %v\n", resp)

	go func() {
		for {
			for _, j := range work.Jobs {
				fmt.Printf("Fetching %s\n", j.URL)
			}
			time.Sleep(10 * time.Second)
		}
	}()

	return nil
}

func (a *fetcher) Stop() {
}
