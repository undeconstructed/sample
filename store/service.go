package store

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "store")

// New makes a new store
func New(grpcBind string) common.Service {
	feeds := someFeeds{}

	return &service{
		grpcBind: grpcBind,
		feeds:    feeds,
	}
}

type service struct {
	grpcBind string

	feeds someFeeds
}

func (s *service) Start(ctx context.Context) error {
	log.Info("Starting")

	gsrvr, err := makeGSrv(s.grpcBind, s.feeds)
	if err != nil {
		return err
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return gsrvr.Start(gctx)
	})

	return grp.Wait()
}
