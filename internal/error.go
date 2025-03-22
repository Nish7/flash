package server

import (
	"log"
	"net"
)

func (s *Server) ErrorHandler(err error, conn net.Conn) {
	log.Printf("[%s] Error: %s", conn.RemoteAddr().String(), err.Error())
	errMsg := EncodeError(err.Error())
	conn.Write(errMsg)
}
