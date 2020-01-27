package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xingshuo/gonet/src"
	"github.com/xingshuo/gonet/examples/helloworld"
)

type Client struct {
	reqSeq int
}

func (s *Client) OnConnected(sender gonet.Sender) error {
	log.Print("Connected ok!\n")
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for {
			<- ticker.C
			req := fmt.Sprintf("pkg %d", s.reqSeq)
			data := proto.Pack(req)
			sender.Send(data)
			s.reqSeq++
		}
	}()
	return nil
}

func (s *Client) OnMessage(sender gonet.Sender, b []byte) (n int, err error) {
	n, msg := proto.Unpack(b)
	if len(msg) > 0 {
		log.Printf("recv server reply:%s\n",  msg)
	}
	return n, nil
}

func main() {
	d,err := gonet.NewDialer("127.0.0.1:5051", &Client{reqSeq: 1}, gonet.WithMaxRetryTimes(5))
	if err != nil {
		log.Fatalf("dial failed:%v\n", err)
	}
	err = d.Start()
	if err != nil {
		log.Fatalf("dialer start failed:%v\n", err)
	}
	//等待接收信号
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	<-sigs
	err = d.Shutdown()
	if err != nil {
		log.Fatalf("shut down failed:%v\n", err)
	}
	log.Print("recv sig do shutdown!\n")
}
