package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"log"
	"time"

	gonet "github.com/xingshuo/gonet/src"
	"github.com/xingshuo/gonet/examples/helloworld"
)

type Server struct {
	rspSeq int
}

//固定4字节包头长度 + 内容
func (s *Server) OnConnected(sender gonet.Sender) error {
	log.Printf("New Connection From %s\n", sender.PeerAddr())
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for {
			<-ticker.C
			reply := fmt.Sprintf("ack %d", s.rspSeq)
			data := proto.Pack(reply)
			sender.Send(data)
			s.rspSeq++
		}

	}()
	return nil
}

func (s *Server) OnMessage(sender gonet.Sender, b []byte) (n int, err error) {
	n, msg := proto.Unpack(b)
	if len(msg) > 0 {
		log.Printf("recv client request: %s from %s\n",  msg, sender.PeerAddr())
	}
	return n, nil
}

func handleSignal(l *gonet.Listener) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	<-sigs
	l.GracefulStop()
	log.Print("stop ok!")
}

func main() {
	l,err := gonet.NewListener(":5051", &Server{rspSeq: 1})
	if err != nil {
		log.Fatalf("new listener failed:%v\n", err)
	}
	go handleSignal(l)
	l.Serve()
	log.Print("serve exit!\n")
}
