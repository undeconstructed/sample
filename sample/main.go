package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/undeconstructed/sample/config"
	"github.com/undeconstructed/sample/fetcher"
	"github.com/undeconstructed/sample/frontend"
	"github.com/undeconstructed/sample/store"
)

type Service interface {
	Start() error
	Stop()
}

type testService struct {
	services []Service
}

func makeTestService() *testService {
	return &testService{
		services: []Service{
			config.New(8001, "localhost:8002"),
			fetcher.New(),
			frontend.New(8088, "localhost:8001"),
			store.New(8002),
		},
	}
}

func (ts *testService) Start() error {
	for _, service := range ts.services {
		err := service.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ts *testService) Stop() {
	for _, service := range ts.services {
		service.Stop()
	}
}

func main() {
	comp := os.Args[1]

	var service Service

	switch comp {
	case "test":
		service = makeTestService()
	case "config":
		service = config.New(8081, os.Args[2])
	case "frontend":
		service = frontend.New(8080, os.Args[2])
	case "fetcher":
		service = fetcher.New()
	case "store":
		service = store.New(8082)
	default:
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	err := service.Start()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("ok")

	s := <-c
	fmt.Println("Got signal:", s)
	service.Stop()
}
