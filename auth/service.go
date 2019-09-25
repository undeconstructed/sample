package auth

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "auth")

// New makes a new auth service
func New(httpBind string, userURL string) common.Service {
	return &service{
		httpBind: httpBind,
		userURL:  userURL,
	}
}

type service struct {
	httpBind string
	userURL  string
}

func (s *service) Start(ctx context.Context) error {
	log.Info("Starting")

	hsrvr, err := makeHSrv(s.httpBind)
	if err != nil {
		return err
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return hsrvr.Start(gctx)
	})

	return grp.Wait()
}
