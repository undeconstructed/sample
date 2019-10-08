package config

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "config")

// New makes a config service
func New(grpcBind, httpBind string, path string, storeURL string) common.Service {
	return &service{
		grpcBind: grpcBind,
		httpBind: httpBind,
		path:     path,
		storeURL: storeURL,
	}
}

type service struct {
	grpcBind string
	httpBind string
	path     string
	storeURL string
}

func (s *service) Start(ctx context.Context) error {
	log.Info("Starting")

	store, err := makeStore(s.path, s.storeURL)
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
				return gctx.Err()
			}
		}
	})
	grp.Go(func() error {
		return store.Start(gctx)
	})
	grp.Go(func() error {
		return sched.Start(gctx)
	})
	grp.Go(func() error {
		return gsrvr.Start(gctx, store, sched)
	})
	grp.Go(func() error {
		return hsrvr.Start(gctx, store)
	})

	return grp.Wait()
}
