package user

import (
	"context"
	"net"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

type gsrv struct {
	listener net.Listener
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

func (s *gsrv) Start(ctx context.Context) error {
	srv := grpc.NewServer()
	common.RegisterUserServer(srv, s)

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
