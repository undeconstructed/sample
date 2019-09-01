package config

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "config")

// Config holds config
type Config interface {
	common.Service
}

// New makes
func New(grpcBind, httpBind string, storeURL string) Config {
	a := &config{
		grpcBind: grpcBind,
		httpBind: httpBind,
		storeURL: storeURL,
		sources:  map[string]common.SourceConfig{},
	}

	// dummy date
	a.sources["bbc"] = common.SourceConfig{
		ID:    "bbc",
		URL:   "http://feeds.bbci.co.uk/news/uk/rss.xml",
		Store: storeURL,
	}
	a.sources["itv"] = common.SourceConfig{
		ID:    "itv",
		URL:   "http://itv.thing",
		Store: storeURL,
	}

	return a
}

type config struct {
	grpcBind string
	httpBind string
	storeURL string

	stopped chan bool
	stop    context.CancelFunc

	sources map[string]common.SourceConfig
}

func (a *config) Start() error {
	log.Info("Starting")

	ctx, cancel := context.WithCancel(context.Background())
	a.stopped = make(chan bool)
	a.stop = cancel

	grp, gctx := errgroup.WithContext(ctx)

	rl, err := net.Listen("tcp", a.grpcBind)
	if err != nil {
		return err
	}

	hl, err := net.Listen("tcp", a.httpBind)
	if err != nil {
		return err
	}

	grp.Go(func() error {
		return a.startGRPC(gctx, rl)
	})
	grp.Go(func() error {
		return a.startHTTP(gctx, hl)
	})

	go func() {
		<-ctx.Done()
		log.Info("Stopping")
		// cancel was automatically propogated into grp
		grp.Wait()
		close(a.stopped)
	}()

	return nil
}

func (a *config) startGRPC(ctx context.Context, l net.Listener) error {
	srv := grpc.NewServer()
	common.RegisterConfigServer(srv, a)

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	return srv.Serve(l)
}

func (a *config) startHTTP(ctx context.Context, l net.Listener) error {
	router := gin.Default()
	router.GET("/sources", a.getSources)
	router.PUT("/sources/:id", a.putSource)
	router.GET("/sources/:id", a.getSource)

	srv := http.Server{
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	return srv.Serve(l)
}

func (a *config) GetSources(context.Context, *common.Nil) (*common.ConfigSources, error) {
	sources := make([]*common.ConfigSource, 0)

	for i, s := range a.sources {
		sources = append(sources, &common.ConfigSource{
			ID:    i,
			URL:   s.URL,
			Store: s.Store,
		})
	}

	out := &common.ConfigSources{
		Sources: sources,
	}

	return out, nil
}

func (a *config) GetWork(context.Context, *common.Nil) (*common.FetchWork, error) {
	jobs := []*common.FetchJob{}

	for i, s := range a.sources {
		jobs = append(jobs, &common.FetchJob{
			ID:    i,
			URL:   s.URL,
			Store: a.storeURL,
		})
	}

	out := &common.FetchWork{
		Jobs: jobs,
	}

	return out, nil
}

func (a *config) getSources(c *gin.Context) {
	sources := []common.SourceConfig{}

	for i, s := range a.sources {
		// XXX - currently this just copies the slice for no reason
		sources = append(sources, common.SourceConfig{
			ID:    i,
			URL:   s.URL,
			Store: s.Store,
		})
	}

	c.JSON(http.StatusOK, common.SourcesConfig{
		Sources: sources,
	})
}

func (a *config) putSource(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *config) getSource(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *config) Stop() {
	a.stop()
	<-a.stopped
}
