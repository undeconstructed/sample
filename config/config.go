package config

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "config")

type Config interface {
	common.Service
}

func New(port int, store string) Config {
	a := &config{
		port:    port,
		store:   store,
		sources: map[string]common.SourceConfig{},
	}

	// dummy date
	a.sources["bbc"] = common.SourceConfig{
		ID:    "bbc",
		URL:   "http://feeds.bbci.co.uk/news/uk/rss.xml",
		Store: store,
	}
	a.sources["itv"] = common.SourceConfig{
		ID:    "itv",
		URL:   "http://itv.thing",
		Store: store,
	}

	return a
}

type config struct {
	port    int
	stopped chan bool
	stop    context.CancelFunc

	srv  *grpc.Server
	hsrv http.Server

	store   string
	sources map[string]common.SourceConfig
}

func (a *config) Start() error {
	log.Info("Starting")

	ctx, cancel := context.WithCancel(context.Background())
	a.stopped = make(chan bool)
	a.stop = cancel

	router := gin.Default()
	router.GET("/sources", a.getSources)
	router.PUT("/sources/:id", a.putSource)
	router.GET("/sources/:id", a.getSource)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}

	a.hsrv = http.Server{
		Handler: router,
	}

	a.srv = grpc.NewServer()
	common.RegisterConfigServer(a.srv, a)

	go func() {
		a.hsrv.Serve(l)
	}()

	go func() {
		a.srv.Serve(l)
	}()

	go func() {
		<-ctx.Done()
		log.Info("Stopping")
		a.srv.GracefulStop()
		a.hsrv.Shutdown(context.Background())
		l.Close()
		close(a.stopped)
	}()

	return nil
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
			Store: a.store,
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
