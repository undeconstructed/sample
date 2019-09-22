package store

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "store")

// New makes a new store
func New(grpcBind string, dataPath string) common.Service {
	return &service{
		grpcBind: grpcBind,
		dataPath: dataPath,
	}
}

// NewFromArgs tries to parse command line args into a service
func NewFromArgs(args []string) common.Service {
	return New(":8000", args[0])
}

type service struct {
	grpcBind string
	dataPath string
}

func (s *service) Start(ctx context.Context) error {
	log.Info("Starting")

	bend, err := makeBackend(s.dataPath)
	if err != nil {
		return err
	}
	gsrvr, err := makeGSrv(s.grpcBind)
	if err != nil {
		return err
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return bend.Start(gctx)
	})
	grp.Go(func() error {
		return gsrvr.Start(gctx, bend)
	})

	return grp.Wait()
}
