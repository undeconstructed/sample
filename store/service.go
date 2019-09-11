package store

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "store")

// New makes
func New(grpcBind string) common.Service {
	feeds := someFeeds{}

	return &service{
		grpcBind: grpcBind,
		feeds:    feeds,
	}
}

type service struct {
	grpcBind string

	stopped chan bool
	stop    context.CancelFunc

	feeds someFeeds
}

func (s *service) Start() error {
	log.Info("Starting")

	ctx, cancel := context.WithCancel(context.Background())
	s.stopped = make(chan bool)
	s.stop = cancel

	gsrvr, err := makeGSrv(s.grpcBind, s.feeds)
	if err != nil {
		return err
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return gsrvr.start(gctx)
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
