package util

import "net"

func AllocListenAddr() (addr *net.TCPAddr, err error) {
	ln, err := net.Listen("tcp4", "localhost:")
	if err != nil {
		return nil, err
	}
	addr = ln.Addr().(*net.TCPAddr)
	err = ln.Close()
	return
}
