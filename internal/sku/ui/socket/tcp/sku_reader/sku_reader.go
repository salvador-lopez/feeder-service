package sku_reader

import (
	"bufio"
	"net"
)
//go:generate mockgen -destination=mock/sku_reader_mockgen_mock.go -package=mock . SkuReader
type SkuReader interface {
	Read()(string, error)
}

type SkuReaderImpl struct {
	addr string
}

func New(addr string) *SkuReaderImpl {
	return &SkuReaderImpl{addr: addr}
}



func (h *SkuReaderImpl) Read()(string, error) {
	conn, err := h.connect()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	line, _, err := bufio.NewReader(conn).ReadLine()
	if err != nil {
		return "", err
	}

	return string(line), nil
}

func (h *SkuReaderImpl) connect() (net.Conn, error) {
	listener, _ := net.Listen("tcp", h.addr)
	c, err := listener.Accept()
	if err != nil {
		return nil, err
	}
	return c, nil
}