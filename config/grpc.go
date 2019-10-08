package config

import (
	"context"
	"net"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

type gsrv struct {
	listener net.Listener
	cfgCh    chan *cfg
	cfg      *cfg
	store    *store
	sched    *sched
}

func makeGSrv(bind string) (*gsrv, error) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}
	return &gsrv{
		listener: l,
		cfgCh:    make(chan *cfg),
	}, nil
}

func (s *gsrv) Start(ctx context.Context, store *store, sched *sched) error {
	s.store = store
	s.sched = sched

	select {
	case c := <-s.cfgCh:
		s.cfg = c
	case <-ctx.Done():
		return ctx.Err()
	}

	srv := grpc.NewServer()
	common.RegisterConfigServer(srv, s)

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		for {
			select {
			case c := <-s.cfgCh:
				s.cfg = c
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

func (s *gsrv) GetServeWork(context.Context, *common.ServeWorkRequest) (*common.ServeWork, error) {
	feeds := make([]*common.ServeFeed, 0, len(s.cfg.Sources))

	for i, src := range s.cfg.Sources {
		feeds = append(feeds, &common.ServeFeed{
			ID:    i,
			Store: src.Spec.Store,
		})
	}

	out := &common.ServeWork{
		Feeds: feeds,
	}

	return out, nil
}

func (s *gsrv) GetFetchWork(ctx context.Context, _ *common.FetchWorkRequest) (*common.FetchWork, error) {
	out, err := s.sched.getWork(ctx)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *gsrv) ReportFetchResult(ctx context.Context, req *common.FetchReport) (*common.Empty, error) {
	return nil, s.store.Update([]interface{}{*req})
}
