package tcp

import "net"

type ConnectionHandler struct {

}

func NewConnectionHandler() *ConnectionHandler {
	return &ConnectionHandler{}
}

func (h *ConnectionHandler) Handle(conn net.Conn) error {
	return nil
}

