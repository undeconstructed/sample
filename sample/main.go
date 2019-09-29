package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.StandardLogger()

func main() {
	cmdTestMode := makeTestMode()
	cmdRun := &cobra.Command{
		Use:   "run <service>",
		Short: "Run a service",
		Long:  `run run run.`,
		Args:  cobra.MinimumNArgs(0),
	}
	cmdRunConfig := makeRunConfig()
	cmdRunFetcher := makeRunFetcher()
	cmdRunFrontend := makeRunFrontend()
	cmdRunStore := makeRunStore()

	var rootCmd = &cobra.Command{Use: "sample"}
	rootCmd.AddCommand(cmdRun, cmdTestMode)
	cmdRun.AddCommand(cmdRunConfig, cmdRunFetcher, cmdRunFrontend, cmdRunStore)
	rootCmd.Execute()
}

func runService(service common.Service) {
	log.Info("Starting")

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
		stop()
		if e != nil {
			log.WithError(e).Error("Error")
			os.Exit(1)
		}
		os.Exit(0)
	case s := <-c:
		log.WithField("signal", s).Info("Got signal")
		stop()
	}

	select {
	case e := <-errCh:
		if e != nil {
			log.WithError(e).Error("Error")
			os.Exit(1)
		}
		os.Exit(0)
	case <-time.After(10 * time.Second):
		log.Error("Shutdown timeout")
		os.Exit(1)
	}
}
