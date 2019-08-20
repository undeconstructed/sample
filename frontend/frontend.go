package frontend

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Frontend interface {
	Start() error
	Stop()
}

func New(port int, configURL string) Frontend {
	a := &server{port: port, configURL: configURL}
	return a
}

type server struct {
	port      int
	configURL string
	stopped   chan bool
	stop      context.CancelFunc
	srv       http.Server
}

func (a *server) Start() error {
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

	go func() {
		a.srv.Serve(l)
	}()

	go func() {
		<-ctx.Done()
		fmt.Println("Frontend stopping")
		a.srv.Shutdown(context.Background())
		l.Close()
		close(a.stopped)
	}()

	return nil
}

func (a *server) getFeed(c *gin.Context) {
	query := c.Param("query")
	message := "Feed " + query
	c.String(http.StatusOK, message)
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
