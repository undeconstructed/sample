package frontend

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "frontend")

// New makes a new frontend
func New(httpBind string, configURL string, userURL string) common.Service {
	return &service{
		httpBind:  httpBind,
		configURL: configURL,
		userURL:   userURL,
	}
}

type service struct {
	httpBind  string
	configURL string
	userURL   string

	stopped chan bool
	stop    context.CancelFunc
}

func (s *service) Start(ctx context.Context) error {
	log.Info("Starting")

	index := &ArticleIndex{}

	updater, err := makeUpdater(s.configURL)
	if err != nil {
		return err
	}
	hsrvr, err := makeHSrv(s.httpBind)
	if err != nil {
		return err
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return updater.Start(gctx, index)
	})
	grp.Go(func() error {
		return hsrvr.Start(gctx, index)
	})

	return grp.Wait()
}
