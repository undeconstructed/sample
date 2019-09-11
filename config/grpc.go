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

func (s *gsrv) Start(ctx context.Context, sched *sched) error {
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

func (s *gsrv) GetSources(context.Context, *common.Nil) (*common.ConfigSources, error) {
	sources := make([]*common.ConfigSource, 0)

	for i, src := range s.cfg.Sources {
		sources = append(sources, &common.ConfigSource{
			ID:    i,
			URL:   src.URL,
			Store: src.Store,
		})
	}

	out := &common.ConfigSources{
		Sources: sources,
	}

	return out, nil
}

func (s *gsrv) GetWork(ctx context.Context, _ *common.Nil) (*common.FetchWork, error) {
	out, err := s.sched.getWork(ctx)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
