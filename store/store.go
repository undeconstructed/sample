package store

import (
	"context"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "store")

// Store stores
type Store interface {
	common.Service
}

// New makes
func New(httpBind string) Store {
	feeds := map[string]*feedHolder{}

	return &store{
		httpBind: httpBind,
		feeds:    feeds,
	}
}

// storeArticle is an article as stored in the store.
type storeArticle struct {
	ID    string    `json:"id"`
	Title string    `json:"title"`
	Date  time.Time `json:"date"`
	Body  string    `json:"body"`
}

type feedSorter []storeArticle

func (a feedSorter) Len() int           { return len(a) }
func (a feedSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a feedSorter) Less(i, j int) bool { return a[i].Date.Before(a[j].Date) }

type feedHolder struct {
	id       string
	articles map[string]storeArticle
}

func newFeedHolder(id string) *feedHolder {
	return &feedHolder{
		id:       id,
		articles: map[string]storeArticle{},
	}
}

func (f *feedHolder) add(a storeArticle) {
	if _, exists := f.articles[a.ID]; exists {
		return
	}
	// log.WithField("article", a).Info("Storing article")
	f.articles[a.ID] = a
}

// XXX
func (f *feedHolder) getSomeArticles() []storeArticle {
	articles := make(feedSorter, 0, len(f.articles))
	for _, a := range f.articles {
		articles = append(articles, a)
		sort.Sort(articles)
	}
	return articles
}

type store struct {
	httpBind string

	stopped chan bool
	stop    context.CancelFunc

	feeds map[string]*feedHolder
}

func (a *store) Start() error {
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
		return a.startGRPC(gctx, l)
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

func (a *store) startGRPC(ctx context.Context, l net.Listener) error {
	srv := grpc.NewServer()
	common.RegisterStoreServer(srv, a)

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	return srv.Serve(l)
}

func (a *store) PostFeed(_ context.Context, req *common.StorePostFeedRequest) (*common.StorePostFeedResponse, error) {
	// XXX nothing threadsafe
	fid := req.FeedID

	feed, exists := a.feeds[fid]
	if !exists {
		feed = newFeedHolder(fid)
		a.feeds[feed.id] = feed
	}

	for _, a := range req.Articles {
		feed.add(storeArticle{
			ID:    a.ID,
			Date:  time.Unix(a.Date, 0),
			Title: a.Title,
			Body:  a.Body,
		})
	}

	out := &common.StorePostFeedResponse{}

	return out, nil
}

func (a *store) GetFeed(_ context.Context, req *common.StoreGetFeedRequest) (*common.StoreGetFeedResponse, error) {
	// XXX nothing threadsafe
	fid := req.FeedID
	// since := ...

	// TODO - selective fetching
	feed, exists := a.feeds[fid]
	if !exists {
		return nil, fmt.Errorf("no feed: %s", fid)
	}

	articles1 := feed.getSomeArticles()
	articles := make([]*common.StoreArticle, 0, len(articles1))
	for _, a := range articles1 {
		articles = append(articles, &common.StoreArticle{
			ID:    a.ID,
			Date:  a.Date.Unix(),
			Title: a.Title,
			Body:  a.Body,
		})
	}

	out := &common.StoreGetFeedResponse{
		Articles: articles,
	}

	return out, nil
}

func (a *store) Stop() error {
	a.stop()
	<-a.stopped
	return nil
}
