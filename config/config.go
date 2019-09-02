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
		config:  makeStore(),
	}

	// dummy date
	// a.sources["bbc"] = common.SourceConfig{
	// 	ID:    "bbc",
	// 	URL:   "http://feeds.bbci.co.uk/news/uk/rss.xml",
	// 	Store: storeURL,
	// }
	// a.sources["itv"] = common.SourceConfig{
	// 	ID:    "itv",
	// 	URL:   "http://itv.thing",
	// 	Store: storeURL,
	// }

	return a
}

type config struct {
	grpcBind string
	httpBind string
	storeURL string

	stopped chan bool
	stop    context.CancelFunc

	config *store
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
		return a.startStore(gctx)
	})
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

func (a *config) startStore(ctx context.Context) error {
	return nil
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

	a.config.read(func (cfg *cfg) {
		for i, s := range cfg.Sources {
			sources = append(sources, &common.ConfigSource{
				ID:    i,
				URL:   s.URL,
				Store: s.Store,
			})
		}
	})

	out := &common.ConfigSources{
		Sources: sources,
	}

	return out, nil
}

func (a *config) GetWork(context.Context, *common.Nil) (*common.FetchWork, error) {
	jobs := []*common.FetchJob{}

	a.config.read(func (cfg *cfg) {
		for i, s := range cfg.Sources {
			jobs = append(jobs, &common.FetchJob{
				ID:    i,
				URL:   s.URL,
				Store: a.storeURL,
			})
		}
	})

	out := &common.FetchWork{
		Jobs: jobs,
	}

	return out, nil
}

func (a *config) getSources(c *gin.Context) {
	sources := []common.SourceConfig{}

		a.config.read(func (cfg *cfg) {
		for i, s := range cfg.Sources {
			// XXX - currently this just copies the slice for no reason
			sources = append(sources, common.SourceConfig{
				ID:    i,
				URL:   s.URL,
				Store: s.Store,
			})
		}
	})

	c.JSON(http.StatusOK, common.SourcesConfig{
		Sources: sources,
	})
}

func (a *config) putSource(c *gin.Context) {
	in := common.SourceConfig{}
	sid := c.Param("id")
	err := c.Bind(&in)
	if err != nil {
		c.String(http.StatusBadRequest, "bad request")
		return
	}

	in.ID = sid
	if in.Store == "" {
		in.Store = a.storeURL
	}

	a.config.write(func (cfg *cfg) {
		cfg.Sources[sid] = in
	})

	c.String(http.StatusOK, "ok")
}

func (a *config) getSource(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *config) Stop() {
	a.stop()
	<-a.stopped
}
