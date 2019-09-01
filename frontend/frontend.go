package frontend

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "frontend")

// Frontend is the frontend microservice
type Frontend interface {
	common.Service
}

// New makes a new Frontend
func New(port int, configURL string) Frontend {
	articles := []common.OutputArticle{}

	return &server{
		port:      port,
		configURL: configURL,
		sources:   []*common.ConfigSource{},
		articles:  articles,
	}
}

type server struct {
	port       int
	configURL  string
	stopped    chan bool
	stop       context.CancelFunc
	srv        http.Server
	sources    []*common.ConfigSource
	lastUpdate time.Time
	// article index (with full data)
	articles []common.OutputArticle
}

func (a *server) Start() error {
	log.Info("Starting")

	ctx, cancel := context.WithCancel(context.Background())
	a.stopped = make(chan bool)
	a.stop = cancel

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
				return
			}
		}
	}()

	// server must be stopped from another routine
	go func() {
		<-ctx.Done()
		log.Info("Stopping")
		a.srv.Shutdown(context.Background())
		l.Close()
		close(a.stopped)
	}()

	return nil
}

// find out what sources exist
func (a *server) updateSources() {
	conn, err := grpc.Dial(a.configURL, grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Error("Error getting source list")
		return
	}
	defer conn.Close()
	c := common.NewConfigClient(conn)

	sources, err := c.GetSources(context.Background(), &common.Nil{})
	if err != nil {
		log.WithError(err).Error("Error getting source list")
		return
	}

	// atomic replace of whole list
	a.sources = sources.Sources
}

// get rid of old things
func (a *server) purgeArticles(before time.Time) {
}

// get any new articles from the store
func (a *server) updateFeeds(since time.Time) {
	for _, s := range a.sources {

		conn, err := grpc.Dial(s.Store, grpc.WithInsecure())
		if err != nil {
			log.WithError(err).Error("Error connecting to store")
			return
		}
		defer conn.Close()
		c := common.NewStoreClient(conn)

		req := &common.StoreGetFeedRequest{
			FeedID: s.ID,
		}

		res, err := c.GetFeed(context.Background(), req)
		if err != nil {
			log.WithError(err).Error("Error updating feed")
			return
		}

		newArticles := make([]common.OutputArticle, 0, len(res.Articles))
		for _, a := range res.Articles {
			newArticles = append(newArticles, common.OutputArticle{
				Source: s.ID,
				ID:     a.ID,
				Title:  a.Title,
				Date:   time.Unix(a.Date, 0),
				Body:   a.Body,
			})
		}

		// TODO - merge, not replace
		a.articles = newArticles
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
