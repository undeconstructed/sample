package user

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "user")

// New makes a new user service
func New(grpcBind string) common.Service {
	return &service{
		grpcBind: grpcBind,
	}
}

type service struct {
	grpcBind string
}

func (s *service) Start(ctx context.Context) error {
	log.Info("Starting")

	gsrvr, err := makeGSrv(s.grpcBind)
	if err != nil {
		return err
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return gsrvr.Start(gctx)
	})

	return grp.Wait()
}
