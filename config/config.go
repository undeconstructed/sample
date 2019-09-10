package config

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "config")

// Config holds config
type Config interface {
	common.Service
}

// New makes
func New(grpcBind, httpBind string, storeURL string) Config {
	a := &config{
		grpcBind: grpcBind,
		httpBind: httpBind,
		storeURL: storeURL,
	}

	return a
}

type config struct {
	grpcBind string
	httpBind string
	storeURL string

	stopped chan bool
	stop    context.CancelFunc
}

func (s *config) Start() error {
	log.Info("Starting")

	ctx, cancel := context.WithCancel(context.Background())
	s.stopped = make(chan bool)
	s.stop = cancel

	store := makeStore("config.json", s.storeURL)
	sched := makeSched()
	hsrvr, err := makeHSrv(s.httpBind)
	if err != nil {
		return err
	}
	gsrvr, err := makeGSrv(s.grpcBind)
	if err != nil {
		return err
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		for {
			select {
			case c := <-store.onCh:
				sched.cfgCh <- c
				hsrvr.cfgCh <- c
				gsrvr.cfgCh <- c
			case <-gctx.Done():
				return nil
			}
		}
	})
	grp.Go(func() error {
		return store.start(gctx)
	})
	grp.Go(func() error {
		return sched.start(gctx)
	})
	grp.Go(func() error {
		return gsrvr.start(gctx, sched)
	})
	grp.Go(func() error {
		return hsrvr.start(gctx, store)
	})

	go func() {
		<-ctx.Done()
		log.Info("Stopping")
		// cancel was automatically propogated into grp
		grp.Wait()
		close(s.stopped)
	}()

	return nil
}

func (s *config) Stop() error {
	s.stop()
	<-s.stopped
	return nil
}
