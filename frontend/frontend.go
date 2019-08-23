package frontend

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	resty "github.com/go-resty/resty/v2"

	"github.com/undeconstructed/sample/common"
)

// Frontend is the frontend microservice
type Frontend interface {
	common.Service
}

// New makes a new Frontend
func New(port int, configURL string) Frontend {
	articles := []common.OutputArticle{
		common.OutputArticle{
			Source: "bbc",
			ID:     "1",
			Date:   "1",
			Body:   "this is a dummy article",
		},
	}

	return &server{
		port:      port,
		configURL: configURL,
		sources:   common.SourcesConfig{},
		articles:  articles,
	}
}

type server struct {
	port       int
	configURL  string
	stopped    chan bool
	stop       context.CancelFunc
	srv        http.Server
	client     *resty.Client
	sources    common.SourcesConfig
	lastUpdate time.Time
	// article index (with full data)
	articles []common.OutputArticle
}

func (a *server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.stopped = make(chan bool)
	a.stop = cancel
	a.client = resty.New()

	router := gin.Default()
	router.GET("/feed", a.getFeed)
	router.GET("/items/:id", a.getItem)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}

	a.srv = http.Server{
		Handler: router,
	}

	// server in new goroutine
	go func() {
		a.srv.Serve(l)
	}()

	// updater in new goroutine
	go func() {
		// could update on request instead, of course
		for {
			now := time.Now()
			a.updateSources()
			a.purgeArticles(now)
			a.updateFeeds(a.lastUpdate)
			a.lastUpdate = now
			t := time.After(10 * time.Second)
			select {
			case <-t:
				continue
			case <-ctx.Done():
				fmt.Println("Fetcher stopping")
				break
			}
		}
	}()

	// server must be stopped from another routine
	go func() {
		<-ctx.Done()
		fmt.Println("Frontend stopping")
		a.srv.Shutdown(context.Background())
		l.Close()
		close(a.stopped)
	}()

	return nil
}

func (a *server) updateSources() {
	// find out what sources exist
	sources := common.SourcesConfig{}

	_, err := a.client.R().
		SetResult(&sources).
		Get("http://" + a.configURL + "/sources/")

	if err != nil {
		fmt.Printf("Error fetching sources list: %v\n", err)
		return
	}

	// atomic replace of whole list
	a.sources = sources
}

func (a *server) purgeArticles(before time.Time) {
	// get rid of old things
}

func (a *server) updateFeeds(since time.Time) {
	// get any new articles from the store

	for _, s := range a.sources.Sources {
		feed := common.StoreFeed{}
		_, err := a.client.R().
			SetQueryParam("since", strconv.FormatInt(since.Unix(), 10)).
			SetResult(&feed).
			Get("http://" + s.Store + "/feeds/" + s.ID)

		if err != nil {
			fmt.Printf("Error updating feed %s: %v\n", s.ID, err)
			continue
		}

		// TODO - create new index
		fmt.Printf("Updated: %v\n", feed)
	}
}

func (a *server) getFeed(c *gin.Context) {
	query := c.Query("query")
	from := c.Query("from")

	// TODO - selecting articles
	out := common.OutputFeed{
		Query:    query,
		Next:     from + "plus",
		Articles: a.articles,
	}

	c.JSON(http.StatusOK, out)
}

func (a *server) getItem(c *gin.Context) {
	id := c.Param("id")
	message := "item " + id
	c.String(http.StatusOK, message)
}

func (a *server) Stop() {
	a.stop()
	<-a.stopped
}
