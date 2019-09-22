package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
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

func makeTestService([]string) common.Service {
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
		log.WithField("state", service).Info("Starting")
		service := service
		grp.Go(func() error {
			return service.Start(gctx)
		})
	}
	return grp.Wait()
}

func main() {
	if len(os.Args) < 2 {
		log.Error("no component specified")
		os.Exit(1)
	}

	comp := os.Args[1]
	log.Info("Starting")

	var makeService func([]string) common.Service

	switch comp {
	case "test":
		makeService = makeTestService
	case "config":
		makeService = config.NewFromArgs
	case "fetcher":
		makeService = fetcher.NewFromArgs
	case "frontend":
		makeService = frontend.NewFromArgs
	case "store":
		makeService = store.NewFromArgs
	default:
		log.Error("unknown component")
		os.Exit(1)
	}

	service := makeService(os.Args[2:])

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, stop := context.WithCancel(context.Background())
	errCh := make(chan error)

	go func() {
		err := service.Start(ctx)
		if err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case e := <-errCh:
		log.WithError(e).Info("Result")
		return
	case s := <-c:
		log.WithField("signal", s).Info("Got signal")
		stop()
	}

	select {
	case e := <-errCh:
		log.WithError(e).Info("Result")
	case <-time.After(10 * time.Second):
		os.Exit(1)
	}
}
