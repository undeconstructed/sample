package store

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/undeconstructed/sample/common"
)

type Store interface {
	common.Service
}

func New(port int) Store {
	feeds := map[string]common.StoreFeed{}

	// dummy data
	feed1 := common.StoreFeed{
		ID: "bbc",
		Articles: []common.StoreArticle{
			{
				ID:   "1",
				Date: "1",
				Body: "this bbc is article 1",
			},
		},
	}
	feeds[feed1.ID] = feed1
	feed2 := common.StoreFeed{
		ID: "itv",
		Articles: []common.StoreArticle{
			{
				ID:   "1",
				Date: "1",
				Body: "this is itv article 1",
			},
		},
	}
	feeds[feed2.ID] = feed2

	return &store{
		port:  port,
		feeds: feeds,
	}
}

type store struct {
	port    int
	stopped chan bool
	stop    context.CancelFunc
	srv     http.Server
	feeds   map[string]common.StoreFeed
}

func (a *store) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.stopped = make(chan bool)
	a.stop = cancel

	router := gin.Default()
	router.POST("/feeds/:fid", a.postFeed)
	router.GET("/feeds/:fid", a.getFeed)
	router.PUT("/feeds/:fid/articles/:aid", a.putArticle)
	router.GET("/feeds/:fid/articles/:aid", a.getArticle)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}

	a.srv = http.Server{
		Handler: router,
	}

	go func() {
		a.srv.Serve(l)
	}()

	go func() {
		<-ctx.Done()
		fmt.Println("Store stopping")
		a.srv.Shutdown(context.Background())
		l.Close()
		close(a.stopped)
	}()

	return nil
}

func (a *store) postFeed(c *gin.Context) {
	in := common.InputFeed{}
	err := c.Bind(&in)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, in)
}

func (a *store) getFeed(c *gin.Context) {
	fid := c.Param("fid")
	// since := c.Query("since")
	feed := a.feeds[fid]
	c.JSON(http.StatusOK, feed)
}

func (a *store) putArticle(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *store) getArticle(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *store) Stop() {
	a.stop()
	<-a.stopped
}
