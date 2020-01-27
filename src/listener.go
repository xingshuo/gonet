//Author: lakefu
//Date:   2020.1.27
//Function: tcp监听端口流程的封装,主要提供多网络连接并发以及优雅退出流程
package gonet

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Listener struct {
	lis      net.Listener
	conns    map[*Conn]bool
	address  string
	cv       *sync.Cond
	mu       sync.Mutex
	quit     *Event
	done     *Event
	serveWG  sync.WaitGroup
	receiver Receiver
}

func (l *Listener) addConn(conn *Conn) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.conns == nil {
		conn.Close()
		return false
	}
	l.conns[conn] = true
	return true
}

func (l *Listener) serveConn(conn *Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err := conn.loopRead()
		if err != nil {
			conn.Close()
			log.Printf("loop read err:%v", err)
		}
		wg.Done()
	}()
	go func() {
		err := conn.loopWrite()
		if err != nil {
			conn.Close()
			log.Printf("loop write err:%v", err)
		}
		wg.Done()
	}()
	wg.Wait()
}

func (l *Listener) removeConn(conn *Conn) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.conns != nil {
		delete(l.conns, conn)
		l.cv.Broadcast()
	}
	log.Printf("remove Conn From %s\n", conn.PeerAddr())
}

func (l *Listener) handleRawConn(rawConn net.Conn) {
	if l.quit.HasFired() {
		rawConn.Close()
		return
	}
	conn := new(Conn)
	err := conn.Init(rawConn, l.receiver)
	if err != nil {
		log.Printf("OnConnected err:%v\n", err)
		return
	}
	if !l.addConn(conn) {
		return
	}
	go func() {
		l.serveConn(conn)
		l.removeConn(conn)
	}()
}

//以下为对外提供接口
func (l *Listener) Serve() error {
	if l.quit.HasFired() {
		return fmt.Errorf("serve repeated.")
	}
	l.mu.Lock()
	if l.lis != nil {
		l.mu.Unlock()
		return fmt.Errorf("serve repeated.")
	}
	lis, err := net.Listen("tcp", l.address)
	if err != nil {
		l.mu.Unlock()
		log.Println("Error listening:", err)
		return err
	}

	l.serveWG.Add(1)
	defer func() {
		l.serveWG.Done()
		if l.quit.HasFired() {
			//已经开始Stop流程的, 需要等Stop流程结束再退出Serve循环
			<-l.done.Done()
		}
	}()

	l.lis = lis
	l.mu.Unlock()

	defer func() {
		l.mu.Lock()
		if l.lis != nil {
			l.lis.Close()
			l.lis = nil
		}
		l.mu.Unlock()
	}()
	log.Println("Listening on "  + ":" + l.address)
	for {
		rawConn, err := lis.Accept()
		if err != nil {
			log.Println("Error accepting: ", err)
			if l.quit.HasFired() { //Stop中调用的lis.Close
				return nil
			}
			return err
		}
		//logs an incoming message
		log.Printf("Received message %s -> %s \n", rawConn.RemoteAddr(), rawConn.LocalAddr())
		// Handle connections in a new goroutine.
		l.serveWG.Add(1)
		go func() {
			l.handleRawConn(rawConn)
			l.serveWG.Done()
		}()
	}
}

func (l *Listener) GracefulStop() {
	l.quit.Fire()       //标识开始走正常退出流程,为了不让Serve循环提前退出
	defer l.done.Fire() //已经走完正常退出流程,通知退出Serve循环
	l.mu.Lock()
	if l.lis == nil { //没调用Serve或Serve流程意外退出了
		l.mu.Unlock()
		return
	}
	l.lis.Close()
	l.lis = nil
	l.mu.Unlock()

	l.serveWG.Wait()
	l.mu.Lock()
	for c := range l.conns {
		c.Close()
	}
	//等待所有连接读写循环退出
	for len(l.conns) != 0 {
		l.cv.Wait()
	}
	l.conns = nil
	l.mu.Unlock()
}