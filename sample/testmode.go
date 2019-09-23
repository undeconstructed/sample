package main

import (
	"context"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
	"github.com/undeconstructed/sample/config"
	"github.com/undeconstructed/sample/fetcher"
	"github.com/undeconstructed/sample/frontend"
	"github.com/undeconstructed/sample/store"
)

// testService is just a hacky way to start all the services in one process.
type testService struct {
	services []common.Service
}

func makeTestService() common.Service {
	self, err := os.Hostname()
	if err != nil {
		self = "localhost"
	}
	return &testService{
		services: []common.Service{
			config.New(":8001", ":8087", "config.json", self+":8002"),
			fetcher.New(self + ":8001"),
			frontend.New(":8088", self+":8001"),
			store.New(":8002", "store.db"),
		},
	}
}

func (ts *testService) Start(ctx context.Context) error {
	grp, gctx := errgroup.WithContext(ctx)
	for _, service := range ts.services {
		service := service
		grp.Go(func() error {
			return service.Start(gctx)
		})
	}
	return grp.Wait()
}
