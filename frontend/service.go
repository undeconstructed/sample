package frontend

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "frontend")

// New makes a new frontend
func New(httpBind string, configURL string) common.Service {
	articles := &someArticlesX{}

	return &service{
		httpBind:  httpBind,
		configURL: configURL,
		articles:  articles,
	}
}

// NewFromArgs tries to parse command line args into a service
func NewFromArgs(args []string) common.Service {
	return New(":8080", args[0])
}

type service struct {
	httpBind  string
	configURL string

	stopped chan bool
	stop    context.CancelFunc

	// article index (with full data)
	articles someArticles
}

func (s *service) Start(ctx context.Context) error {
	log.Info("Starting")

	updater, err := makeUpdater(s.configURL, s.articles)
	if err != nil {
		return err
	}
	hsrvr, err := makeHSrv(s.httpBind, s.articles)
	if err != nil {
		return err
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return updater.Start(gctx)
	})
	grp.Go(func() error {
		return hsrvr.Start(gctx)
	})

	return grp.Wait()
}
