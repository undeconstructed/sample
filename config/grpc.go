package config

import (
	"context"
	"net"

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

func (s *gsrv) start(ctx context.Context, sched *sched) error {
	s.sched = sched

	select {
	case c := <-s.cfgCh:
		s.cfg = c
	case <-ctx.Done():
		return nil
	}

	srv := grpc.NewServer()
	common.RegisterConfigServer(srv, s)

	go func() {
		for {
			select {
			case c := <-s.cfgCh:
				s.cfg = c
			case <-ctx.Done():
				srv.GracefulStop()
				return
			}
		}
	}()

	return srv.Serve(s.listener)
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

func (s *gsrv) GetWork(context.Context, *common.Nil) (*common.FetchWork, error) {
	out := s.sched.getWork()
	return &out, nil
}
