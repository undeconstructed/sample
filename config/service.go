package config

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "config")

// New makes a config service
func New(grpcBind, httpBind string, storeURL string) common.Service {
	s := &service{
		grpcBind: grpcBind,
		httpBind: httpBind,
		storeURL: storeURL,
	}

	return s
}

type service struct {
	grpcBind string
	httpBind string
	storeURL string

	stopped chan bool
	stop    context.CancelFunc
}

func (s *service) Start() error {
	log.Info("Starting")

	ctx, cancel := context.WithCancel(context.Background())
	s.stopped = make(chan bool)
	s.stop = cancel

	store, err := makeStore("config.json", s.storeURL)
	if err != nil {
		return err
	}
	sched, err := makeSched()
	if err != nil {
		return err
	}
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

func (s *service) Stop() error {
	s.stop()
	<-s.stopped
	return nil
}
