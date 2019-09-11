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

type service struct {
	httpBind  string
	configURL string

	stopped chan bool
	stop    context.CancelFunc

	// article index (with full data)
	articles someArticles
}

func (s *service) Start() error {
	log.Info("Starting")

	ctx, cancel := context.WithCancel(context.Background())
	s.stopped = make(chan bool)
	s.stop = cancel

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
		return updater.start(gctx)
	})
	grp.Go(func() error {
		return hsrvr.start(gctx)
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
