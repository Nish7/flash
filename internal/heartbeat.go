package server

import (
	"bufio"
	"log"
	"net"
	"time"
)

func (s *Server) WantHeatbeatHandler(conn net.Conn, reader *bufio.Reader, isHeartbeatRegistered *bool) {
	if *isHeartbeatRegistered {
		log.Printf("[%s] Heartbeat is registed.", conn.RemoteAddr().String())
		return
	}

	req, err := ParseWantHeartbeat(reader)
	if err != nil {
		log.Printf("Failed to parse request %v", err)
		return
	}

	log.Printf("[%s] WantHeartbeat: Recived %v\n", conn.RemoteAddr().String(), req)
	if req.interval == 0 {
		log.Printf("[%s] Recieved 0 inteval req. Heartbeat Disabled", conn.RemoteAddr().String())
		return
	}

	go s.sendHeartbeat(conn, req.interval)
}

func (s *Server) sendHeartbeat(conn net.Conn, decisecond uint32) error {
	interval := time.Duration(decisecond*100) * time.Millisecond
	log.Printf("[%s] Sending %d heartbeats every second", conn.RemoteAddr().String(), interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	heartbeatMsg := EncodeHeartbeat()
	for {
		select {
		case <-ticker.C:
			_, err := conn.Write(heartbeatMsg)
			if err != nil {
				log.Printf("[%s] Failed to send heartbeat: %v", conn.RemoteAddr().String(), err)
				return err
			}
			log.Printf("[%s] Heartbeat sent", conn.RemoteAddr().String())
		}
	}
}
