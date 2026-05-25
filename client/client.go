package client

import (
	"context"
	"log/slog"
	"time"

	"github.com/yadukrishnan2004/antrelay-sdk/executor"
	"github.com/yadukrishnan2004/antrelay-sdk/poller"
	"github.com/yadukrishnan2004/antrelay-sdk/registry"
	"github.com/yadukrishnan2004/antrelay-sdk/reporter"
	"github.com/yadukrishnan2004/antrelay-sdk/task"
	"github.com/yadukrishnan2004/antrelay-sdk/worker"
)

// Config holds the required settings every client needs.
type Config struct{
	ServerURL string
	Queue string
}

// options holds everything that is optional.
// Every field here has a smart default.

type options struct {
	pollInterval time.Duration
	handlerTimeout time.Duration
	maxPollRetries int
}

// Option is a function that modifies options.
// This is the functional options pattern.
type Option func(*options)

func defaultOptions() *options{
	return &options{
		pollInterval: 2 * time.Second,
		handlerTimeout: 30 * time.Second,
		maxPollRetries: 5,
	}
}

// WithPollInterval sets how long the worker waits between polls
func WithPollInterval(d time.Duration) Option {
	return func(o *options){
		o.pollInterval = d
	}
}

// WithHandlerTimeout sets the maximum time a handler can run
func WithHandlerTimeout(d time.Duration) Option {
	return func(o *options) {
		o.handlerTimeout = d
	}
}


// WithMaxPollRetries sets how many times the worker retries
func WithMaxPollRetries(n int) Option {
	return func(o *options){
		o.maxPollRetries = n 
	}
}

//======================================================================

// Client is the main entry point of the SDK.
type Client struct{
	config Config
	opts     *options
	registry *registry.Registry
	cancel  context.CancelFunc
}

func New(cfg Config, optFns ...Option) *Client {
	opts := defaultOptions()

	for _, fn := range optFns {
		fn(opts)
	}

	return &Client{
		config: cfg,
		opts:     opts,
		registry: registry.New(),
	}
}

func (c *Client) Register(name string,handler task.HandlerFunc) error{
	return c.registry.Register(name,handler)
}

func (c *Client) Start(){
	ctx,cancel:=context.WithCancel(context.Background())
	c.cancel=cancel

//------------------------------------------------------------
	//p:=poller.NewMock()
	//rep := reporter.NewMock()
	p := poller.NewHTTP(c.config.ServerURL)
    rep := reporter.NewHTTP(c.config.ServerURL)
	ex := executor.New(c.registry).WithTimeout(c.opts.handlerTimeout)
	w := worker.New(p, ex, rep, c.config.Queue, c.opts.pollInterval)

//-------------------------------------------------------------	

   slog.Info("client starting",
        "queue", c.config.Queue,
        "server", c.config.ServerURL,
        "poll_interval", c.opts.pollInterval,
        "handler_timeout", c.opts.handlerTimeout,
    )
	go w.Run(ctx)
}

func (c *Client) Stop() {
	if c.cancel != nil {
		slog.Info("client stopping", "queue", c.config.Queue)
		c.cancel()
	}
}