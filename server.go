package main

import (
	"cherrycache/internal/handle"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

type Config struct {
	Addr string
}

func ListenAndServer(cfg *Config, handler Handler) {
	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatalf("listen err: %v", err)
	}
	connClose := func() {
		_ = listener.Close()
		_ = handler.Close()
	}
	defer connClose()

	var exitFlag int32
	singleCh := make(chan os.Signal, 1)
	signal.Notify(singleCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-singleCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			atomic.StoreInt32(&exitFlag, 1)
			connClose()
		}
	}()
	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()
	var waiter sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			if atomic.LoadInt32(&exitFlag) == 1 {
				waiter.Wait()
				return
			}
			log.Fatalf("accept err: %v", err)
			continue
		}

		go func() {
			defer waiter.Done()
			waiter.Add(1)
			handler.Handle(ctx, conn)
		}()
	}

}

func main() {
	ListenAndServer(&Config{":8080"}, &handle.BaseHandler{})
}
