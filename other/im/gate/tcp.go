package main

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

type serverConfig struct {
	id           string
	needCrypto   bool
	needMAC      bool
	readTimeout  time.Duration
	writeTimeout time.Duration
}

type tcpServer struct {
	config *serverConfig

	l net.Listener

	stopFlag int32
	stopSig  chan struct{}
}

func newTcpServer(config *serverConfig) *tcpServer {
	t := &tcpServer{}

	t.config = config
	t.stopSig = make(chan struct{}, 0)
	return t
}

func (s *tcpServer) start(network, laddr string) error {
	l, err := net.Listen(network, laddr)
	if err != nil {
		return err
	}
	s.l = l

	fmt.Println("listen at:", l.Addr())

	return s.accept()
}

func (s *tcpServer) stop() {
	if atomic.CompareAndSwapInt32(&s.stopFlag, 0, 1) {
		close(s.stopSig)
	}
}

func (s *tcpServer) accept() error {

	for {
		select {
		case <-s.stopSig:
			return fmt.Errorf("stop")
		default:
		}

		nc, err := s.l.Accept()
		if err != nil {
			return err
		}
		log.Debug("new connect: ", nc.RemoteAddr())
		go s.serveConn(nc)
	}
}

func (s *tcpServer) serveConn(nc net.Conn) {
	if tc, ok := nc.(*net.TCPConn); ok {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(time.Minute * 1)
	}

	newSessionTCP(nc, s.config)
}
