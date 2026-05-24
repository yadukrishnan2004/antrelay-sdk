package client

import (
	"context"
	"log"

	"github.com/yadukrishnan2004/antrelay-sdk/executor"
	"github.com/yadukrishnan2004/antrelay-sdk/poller"
	"github.com/yadukrishnan2004/antrelay-sdk/registry"
	"github.com/yadukrishnan2004/antrelay-sdk/reporter"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
	"github.com/yadukrishnan2004/antrelay-sdk/worker"
)

type Config struct{
	ServerURL string
	Queue string
}

type Client struct{
	config Config
	registry *registry.Registry
	cancel  context.CancelFunc
}

func New(cfg Config)*Client{
	return &Client{
		config: cfg,
		registry: registry.New(),
	}
}

func (c *Client) Register(name string,handler task.HandlerFunc) error{
	return c.registry.Register(name,handler)
}

func (c *Client) Start(){
	ctx,cancel:=context.WithCancel(context.Background())
	c.cancel=cancel




	//p:=poller.NewMock()
	//rep := reporter.NewMock()
	p := poller.NewHTTP(c.config.ServerURL)
    rep := reporter.NewHTTP(c.config.ServerURL)
	ex := executor.New(c.registry)
	w := worker.New(p, ex, rep, c.config.Queue)

	

	log.Printf("client: starting worker on queue %q", c.config.Queue)
	go w.Run(ctx)
}

func (c *Client) Stop() {
	if c.cancel != nil {
		log.Printf("client: stopping worker")
		c.cancel()
	}
}