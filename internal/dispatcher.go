package server

import (
	"bufio"
	"log"
	"net"
)

func (s *Server) HandleDispatcherReq(conn net.Conn, reader *bufio.Reader) {
	defer delete(s.cameras, conn)

	d, err := ParseDispatcherRecord(reader)
	if err != nil {
		log.Printf("Failed to parse request %v", err)
		return
	}

	log.Printf("[%s] Dispatcher Recived %v\n", conn.RemoteAddr().String(), d)
	s.dispatchers[conn] = d
	return
}
