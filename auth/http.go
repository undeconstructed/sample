package auth

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type hsrv struct {
	listener net.Listener
}

func makeHSrv(bind string) (*hsrv, error) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}
	return &hsrv{
		listener: l,
	}, nil
}

func (s *hsrv) Start(ctx context.Context) error {
	router := gin.Default()
	router.GET("/", s.getRoot)

	srv := http.Server{
		Handler: router,
	}

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		for {
			select {
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

func (s *hsrv) getRoot(c *gin.Context) {
	c.JSON(http.StatusOK, struct{}{})
}
