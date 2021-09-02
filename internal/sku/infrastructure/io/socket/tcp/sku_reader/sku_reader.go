package sku_reader

import (
	"bufio"
	"errors"
	"net"
	"strings"
	"time"
)
//go:generate mockgen -destination=mock/sku_reader_mockgen_mock.go -package=mock . SkuReader
type SkuReader interface {
	Read(deadline time.Time)(string, error)
}

type SkuReaderImpl struct {
	listener net.Listener
}

func New(addr string) (*SkuReaderImpl, error){
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &SkuReaderImpl{listener: listener}, nil
}

func (h *SkuReaderImpl) Read(deadline time.Time)(string, error) {
	conn, err := h.connect(deadline)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	line, _, err := bufio.NewReader(conn).ReadLine()
	if err != nil {
		return "", err
	}

	return strings.TrimLeft(string(line), "0"), nil
}

func (h *SkuReaderImpl) connect(deadline time.Time) (net.Conn, error) {
	c := make(chan net.Conn)
	e := make(chan error)
	go func() {
		conn, err := h.listener.Accept()
		if err != nil {
			e <- err
		} else {
			c <- conn
		}
	}()

	select {
	case conn := <-c:
		return conn, nil
	case err := <- e:
		return nil, err
	case <-time.After(deadline.Sub(time.Now())):
		return nil, errors.New("deadline exceeded")
	}
}