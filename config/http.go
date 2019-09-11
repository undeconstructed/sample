package config

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/undeconstructed/sample/common"
)

type hsrv struct {
	listener net.Listener
	cfgCh    chan *cfg
	store    *store
	cfg      *cfg
}

func makeHSrv(bind string) (*hsrv, error) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}
	return &hsrv{
		listener: l,
		cfgCh:    make(chan *cfg),
	}, nil
}

func (s *hsrv) Start(ctx context.Context, store *store) error {
	s.store = store

	select {
	case c := <-s.cfgCh:
		s.cfg = c
	case <-ctx.Done():
		return ctx.Err()
	}

	router := gin.Default()
	router.GET("/sources", s.getSources)
	router.PUT("/sources/:id", s.putSource)
	router.GET("/sources/:id", s.getSource)
	router.DELETE("/sources/:id", s.deleteSource)

	srv := http.Server{
		Handler: router,
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		for {
			select {
			case c := <-s.cfgCh:
				s.cfg = c
			case <-gctx.Done():
				return srv.Shutdown(context.Background())
			}
		}
	})
	grp.Go(func() error {
		return srv.Serve(s.listener)
	})

	return grp.Wait()
}

func (s *hsrv) getSources(c *gin.Context) {
	sources := []common.SourceConfig{}

	for i, src := range s.cfg.Sources {
		// XXX - currently this just copies the slice for no reason
		sources = append(sources, common.SourceConfig{
			ID:    i,
			URL:   src.URL,
			Store: src.Store,
		})
	}

	c.JSON(http.StatusOK, common.SourcesConfig{
		Sources: sources,
	})
}

func (s *hsrv) putSource(c *gin.Context) {
	in := common.SourceConfig{}
	sid := c.Param("id")
	err := c.Bind(&in)
	if err != nil {
		c.String(http.StatusBadRequest, "bad request")
		return
	}

	in.ID = sid

	err = s.store.Update([]interface{}{in})
	if err != nil {
		c.String(http.StatusBadRequest, "couldn't save")
		return
	}

	c.String(http.StatusOK, "ok")
}

func (s *hsrv) getSource(c *gin.Context) {
	sid := c.Param("id")
	out, exists := s.cfg.Sources[sid]
	if !exists {
		c.String(http.StatusNotFound, "not found")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *hsrv) deleteSource(c *gin.Context) {
	sid := c.Param("id")
	in := common.SourceConfig{ID: sid}

	err := s.store.Update([]interface{}{in})
	if err != nil {
		c.String(http.StatusBadRequest, "couldn't save")
		return
	}

	c.String(http.StatusOK, "ok")
}
