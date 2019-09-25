package main

import (
	"context"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/auth"
	"github.com/undeconstructed/sample/common"
	"github.com/undeconstructed/sample/config"
	"github.com/undeconstructed/sample/fetcher"
	"github.com/undeconstructed/sample/frontend"
	"github.com/undeconstructed/sample/store"
	"github.com/undeconstructed/sample/user"
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

	authH := self + ":8086"
	configG := self + ":8001"
	configH := self + ":8087"
	frontendH := self + ":8088"
	storeG := self + ":8002"
	userG := self + ":8003"

	return &testService{
		services: []common.Service{
			auth.New(authH, userG),
			config.New(configG, configH, "config.json", storeG),
			fetcher.New(configG),
			frontend.New(frontendH, configG, userG),
			store.New(storeG, "store.db"),
			user.New(userG),
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
