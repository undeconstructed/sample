package store

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/undeconstructed/sample/common"
)

type Store interface {
	common.Service
}

func New(port int) Store {
	feeds := map[string]*feedHolder{}

	return &store{
		port:  port,
		feeds: feeds,
	}
}

type feedSorter []common.StoreArticle

func (a feedSorter) Len() int           { return len(a) }
func (a feedSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a feedSorter) Less(i, j int) bool { return a[i].Date.Before(a[j].Date) }

type feedHolder struct {
	id       string
	articles map[string]common.StoreArticle
}

func newFeedHolder(id string) *feedHolder {
	return &feedHolder{
		id:       id,
		articles: map[string]common.StoreArticle{},
	}
}

func (f *feedHolder) add(in []common.StoreArticle) {
	for _, a := range in {
		f.articles[a.ID] = a
	}
}

// XXX
func (f *feedHolder) getSomeArticles() []common.StoreArticle {
	articles := make(feedSorter, len(f.articles))
	for _, a := range f.articles {
		articles = append(articles, a)
		sort.Sort(articles)
	}
	return articles
}

type store struct {
	port    int
	stopped chan bool
	stop    context.CancelFunc
	srv     http.Server
	feeds   map[string]*feedHolder
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
	fid := c.Param("fid")
	in := common.InputFeed{}
	err := c.Bind(&in)
	if err != nil {
		return
	}

	// XXX nothing threadsafe

	feed, exists := a.feeds[fid]
	if !exists {
		feed = newFeedHolder(fid)
		a.feeds[feed.id] = feed
	}

	feed.add(in.Articles)

	c.JSON(http.StatusOK, in)
}

func (a *store) getFeed(c *gin.Context) {
	fid := c.Param("fid")

	// since := c.Query("since")
	// TODO - selective fetching
	feed, exists := a.feeds[fid]
	if !exists {
		c.String(http.StatusNotFound, "not found")
		return
	}

	articles := feed.getSomeArticles()

	out := common.StoreFeed{
		ID:       feed.id,
		Articles: articles,
	}

	c.JSON(http.StatusOK, out)
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
