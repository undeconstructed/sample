package store

import (
	"context"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

type gsrv struct {
	listener net.Listener

	bend *backend
}

func makeGSrv(bind string) (*gsrv, error) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}
	return &gsrv{
		listener: l,
	}, nil
}

func (s *gsrv) Start(ctx context.Context, bend *backend) error {
	s.bend = bend

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

func (s *gsrv) GetFeed(ctx context.Context, req *common.StoreGetFeedRequest) (*common.StoreGetFeedResponse, error) {
	fid := req.FeedID

	articles1, err := s.bend.Query(ctx, fid)
	if err != nil {
		return nil, err
	}

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

func (s *gsrv) PostFeed(ctx context.Context, req *common.StorePostFeedRequest) (*common.StorePostFeedResponse, error) {
	// XXX nothing threadsafe
	fid := req.FeedID

	for _, a := range req.Articles {
		s.bend.Put(ctx, fid, storeArticle{
			ID:    a.ID,
			Date:  time.Unix(a.Date, 0),
			Title: a.Title,
			Body:  a.Body,
		})
	}

	out := &common.StorePostFeedResponse{}

	return out, nil
}
