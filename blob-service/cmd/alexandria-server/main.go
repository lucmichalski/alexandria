package main

import (
	"context"
	"fmt"
	"github.com/maestre3d/alexandria/blob-service/pkg/dep"
	"github.com/oklog/run"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Root context, enable complete context shutdown
	ctx, cancel := context.WithCancel(context.Background())
	// Inject root context with cancel inside DI container
	dep.Ctx = ctx

	transport, cleanup, err := dep.InjectTransportService()
	if err != nil {
		panic(err)
	}
	defer func() {
		log.Print("stopping services")
		cleanup()
	}()

	var g run.Group
	{
		l, err := net.Listen("tcp", transport.HTTPProxy.Server.Addr)
		if err != nil {
			log.Fatalf("failed to start http server\nerror: %v", err)
		}

		g.Add(func() error {
			log.Print("starting http service")
			return http.Serve(l, transport.HTTPProxy.Server.Handler)
		}, func(err error) {
			_ = l.Close()
		})
	}
	{
		g.Add(func() error {
			log.Print("starting event service")
			return transport.EventProxy.Server.Serve()
		}, func(error) {
			transport.EventProxy.Server.Close()
		})
	}
	{
		// Set up signal bind
		var (
			cancelInterrupt = make(chan struct{})
			c               = make(chan os.Signal, 2)
		)
		defer close(c)

		g.Add(func() error {
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			// Cancel root context, propagate cancellation
			cancel()
			close(cancelInterrupt)
		})
	}

	_ = g.Run()
}
