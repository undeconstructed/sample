package frontend

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

	articles someArticles
}

func makeHSrv(bind string, articles someArticles) (*hsrv, error) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}
	return &hsrv{
		listener: l,
		articles: articles,
	}, nil
}

func (s *hsrv) Start(ctx context.Context) error {
	router := gin.Default()
	router.GET("/feed", s.getFeed)
	router.GET("/items/:id", s.getItem)

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

func (s *hsrv) getFeed(c *gin.Context) {
	query := c.Query("query")
	from := c.Query("from")

	if from == "" && query == "" {
		articles := make([]common.OutputArticle, 0, 10)

		for i := len(s.articles.list) - 1; i >= 0; i-- {
			articles = append(articles, s.articles.list[i])
		}

		out := common.OutputFeed{
			Query:    query,
			Next:     from + "plus",
			Articles: articles,
		}

		c.JSON(http.StatusOK, out)
		return
	}

	c.String(http.StatusNotImplemented, "only root")
}

func (s *hsrv) getItem(c *gin.Context) {
	id := c.Param("id")
	message := "item " + id
	c.String(http.StatusOK, message)
}
