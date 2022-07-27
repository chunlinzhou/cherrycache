package handle

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

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