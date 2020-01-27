//Author: lakefu
//Date:   2020.1.27
//Function: 对外提供接口
package gonet

import "sync"

func NewDialer(address string, r Receiver, opts ...DialOption) (*Dialer, error) {
	d := &Dialer{
		opts:defaultDialOptions(),
		address:address,
		quit:NewEvent("gonet.Dialer.quit"),
		receiver:r,
	}
	//处理参数
	for _, opt := range opts {
		opt.apply(&d.opts)
	}
	return d, nil
}

func NewListener(address string, r Receiver) (*Listener, error) {
	l := &Listener{
		conns:make(map[*Conn]bool),
		address:address,
		quit:NewEvent("gonet.Listener.quit"),
		done:NewEvent("gonet.Listener.done"),
		receiver:r,
	}
	l.cv = sync.NewCond(&l.mu)
	return l, nil
}