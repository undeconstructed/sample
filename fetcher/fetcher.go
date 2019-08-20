package fetcher

import (
	"fmt"
	"time"
)

type Fetcher interface {
	Start() error
	Stop()
}

func New() Fetcher {
	return &fetcher{}
}

type fetcher struct {
}

func (a *fetcher) Start() error {
	go func() {
		for {
			fmt.Printf("fetch time!\n")
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

func (a *fetcher) Stop() {
}
