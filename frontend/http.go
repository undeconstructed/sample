package frontend

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

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

func (s *hsrv) start(ctx context.Context) error {
	router := gin.Default()
	router.GET("/feed", s.getFeed)
	router.GET("/items/:id", s.getItem)

	srv := http.Server{
		Handler: router,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				srv.Shutdown(context.Background())
				return
			}
		}
	}()

	return srv.Serve(s.listener)
}

func (s *hsrv) getFeed(c *gin.Context) {
	query := c.Query("query")
	from := c.Query("from")

	// TODO - selecting articles
	out := common.OutputFeed{
		Query:    query,
		Next:     from + "plus",
		Articles: s.articles.list,
	}

	c.JSON(http.StatusOK, out)
}

func (s *hsrv) getItem(c *gin.Context) {
	id := c.Param("id")
	message := "item " + id
	c.String(http.StatusOK, message)
}
