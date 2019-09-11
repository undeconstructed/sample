package store

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

type gsrv struct {
	listener net.Listener

	feeds someFeeds
}

func makeGSrv(bind string, feeds someFeeds) (*gsrv, error) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}
	return &gsrv{
		listener: l,
		feeds:    feeds,
	}, nil
}

func (s *gsrv) Start(ctx context.Context) error {
	srv := grpc.NewServer()
	common.RegisterStoreServer(srv, s)

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		for {
			select {
			case <-gctx.Done():
				srv.GracefulStop()
				return nil
			}
		}
	})
	grp.Go(func() error {
		return srv.Serve(s.listener)
	})

	return grp.Wait()
}

func (s *gsrv) GetFeed(_ context.Context, req *common.StoreGetFeedRequest) (*common.StoreGetFeedResponse, error) {
	// XXX nothing threadsafe
	fid := req.FeedID
	// since := ...

	// TODO - selective fetching
	feed, exists := s.feeds[fid]
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

func (s *gsrv) PostFeed(_ context.Context, req *common.StorePostFeedRequest) (*common.StorePostFeedResponse, error) {
	// XXX nothing threadsafe
	fid := req.FeedID

	feed, exists := s.feeds[fid]
	if !exists {
		feed = newFeedHolder(fid)
		s.feeds[feed.id] = feed
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
