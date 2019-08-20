package store

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Store interface {
	Start() error
	Stop()
}

func New(port int) Store {
	return &store{
		port: port,
	}
}

type store struct {
	port    int
	stopped chan bool
	stop    context.CancelFunc
	srv     http.Server
}

func (a *store) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.stopped = make(chan bool)
	a.stop = cancel

	router := gin.Default()
	router.GET("/", a.get1)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}

	a.srv = http.Server{
		Handler: router,
	}

	go func() {
		a.srv.Serve(l)
	}()

	go func() {
		<-ctx.Done()
		fmt.Println("Store stopping")
		a.srv.Shutdown(context.Background())
		l.Close()
		close(a.stopped)
	}()

	return nil
}

func (a *store) get1(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *store) Stop() {
	a.stop()
	<-a.stopped
}
