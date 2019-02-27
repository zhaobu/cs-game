package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type serverConfig struct {
	id string

	tlsConfig    *tls.Config
	readTimeout  time.Duration
	writeTimeout time.Duration
}

type tcpServer struct {
	config *serverConfig

	l net.Listener

	stopFlag int32
	stopSig  chan struct{}
}

func newTCPServer(config *serverConfig) *tcpServer {
	t := &tcpServer{}

	t.config = config
	t.stopSig = make(chan struct{}, 0)
	return t
}

func (s *tcpServer) start(laddr string) error {
	var err error
	var ln net.Listener

	if s.config.tlsConfig == nil {
		ln, err = net.Listen("tcp4", laddr)
	} else {
		ln, err = tls.Listen("tcp4", laddr, s.config.tlsConfig)
	}

	if err != nil {
		return err
	}

	s.l = ln

	logrus.WithFields(logrus.Fields{"listen at:": ln.Addr()}).Info()

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

		go s.serveConn(nc)
	}
}

func (s *tcpServer) serveConn(nc net.Conn) {
	logrus.WithFields(logrus.Fields{"new connect:": nc.RemoteAddr()}).Info()

	if tc, ok := nc.(*net.TCPConn); ok {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(time.Minute * 5)
		tc.SetLinger(0)
		tc.SetNoDelay(false)
		tc.SetReadBuffer(1024 * 4)
		tc.SetWriteBuffer(1024 * 4)
	}

	newSession(nc, s.config)
}
