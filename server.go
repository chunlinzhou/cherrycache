package main

import (
	"bufio"
	"context"
	"io"
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
	ctx, _ := context.WithCancel(context.Background())
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

type BaseHandler struct {
	activeConn sync.Map
	close      int32
}

func (bh *BaseHandler) Handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	if atomic.LoadInt32(&bh.close) == 1 {
		return
	}
	bh.activeConn.Store(conn, 1)
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("connection close")
				bh.activeConn.Delete(conn)
			} else {
				log.Println(err)
			}
			return
		}
		b := []byte(msg)
		conn.Write(b)
	}
}

func (bh *BaseHandler) Close() error {
	atomic.StoreInt32(&bh.close, 1)
	bh.activeConn.Range(func(key, value interface{}) bool {
		err := key.(net.Conn).Close()
		if err != nil {
			log.Fatalln(err)
			return false
		}
		return true
	})
	return nil
}

func main() {
	ListenAndServer(&Config{":8080"}, &BaseHandler{})
}
