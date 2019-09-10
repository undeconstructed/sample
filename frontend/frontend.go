package frontend

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "frontend")

// Frontend is the frontend microservice
type Frontend interface {
	common.Service
}

// New makes a new Frontend
func New(httpBind string, configURL string) Frontend {
	articles := []common.OutputArticle{}

	return &frontend{
		httpBind:  httpBind,
		configURL: configURL,
		sources:   []*common.ConfigSource{},
		articles:  articles,
	}
}

type frontend struct {
	httpBind  string
	configURL string

	stopped chan bool
	stop    context.CancelFunc

	sources    []*common.ConfigSource
	lastUpdate time.Time
	// article index (with full data)
	articles []common.OutputArticle
}

func (a *frontend) Start() error {
	log.Info("Starting")

	ctx, cancel := context.WithCancel(context.Background())
	a.stopped = make(chan bool)
	a.stop = cancel

	grp, gctx := errgroup.WithContext(ctx)

	l, err := net.Listen("tcp", a.httpBind)
	if err != nil {
		return err
	}

	grp.Go(func() error {
		return a.startHTTP(gctx, l)
	})
	grp.Go(func() error {
		return a.startUpdater(gctx)
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

func (a *frontend) startHTTP(ctx context.Context, l net.Listener) error {
	router := gin.Default()
	router.GET("/feed", a.getFeed)
	router.GET("/items/:id", a.getItem)

	srv := http.Server{
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	return srv.Serve(l)
}

func (a *frontend) startUpdater(ctx context.Context) error {
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
			return ctx.Err()
		}
	}
}

// find out what sources exist
func (a *frontend) updateSources() {
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
func (a *frontend) purgeArticles(before time.Time) {
}

// get any new articles from the store
func (a *frontend) updateFeeds(since time.Time) {
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

func (a *frontend) getFeed(c *gin.Context) {
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

func (a *frontend) getItem(c *gin.Context) {
	id := c.Param("id")
	message := "item " + id
	c.String(http.StatusOK, message)
}

func (a *frontend) Stop() error {
	a.stop()
	<-a.stopped
	return nil
}
