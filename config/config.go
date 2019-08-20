package config

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Config interface {
	Start() error
	Stop()
}

func New(port int, store string) Config {
	a := &config{
		port:    port,
		store:   store,
		sources: map[string]sourceCfg{},
	}

	// dummy date
	a.sources["bbc"] = sourceCfg{
		url: "http://bbc.something",
	}

	return a
}

type sourceCfg struct {
	url string
}

type config struct {
	port    int
	stopped chan bool
	stop    context.CancelFunc
	srv     http.Server

	store   string
	sources map[string]sourceCfg
}

func (a *config) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.stopped = make(chan bool)
	a.stop = cancel

	router := gin.Default()
	router.GET("/sources", a.getSites)
	router.GET("/sources/:id", a.getSite)
	router.PUT("/sources/:id", a.putSite)

	router.GET("/work", a.getWork)

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
		fmt.Println("Config stopping")
		a.srv.Shutdown(context.Background())
		l.Close()
		close(a.stopped)
	}()

	return nil
}

func (a *config) getSites(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *config) getSite(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *config) putSite(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *config) getWork(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *config) Stop() {
	a.stop()
	<-a.stopped
}
