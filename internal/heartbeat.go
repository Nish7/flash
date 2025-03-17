package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

func (s *Server) WantHeatbeatHandler(conn net.Conn, reader *bufio.Reader, isHeartbeatRegistered *bool, client *ClientType) {
	if *client == UNKNOWN {
		log.Printf("[%s] Client is not registered.", conn.RemoteAddr().String())
		return
	}

	if *isHeartbeatRegistered {
		log.Printf("[%s] Heartbeat is already registered.", conn.RemoteAddr().String())
		return
	}

	req, err := ParseWantHeartbeat(reader)
	if err != nil {
		log.Printf("Failed to parse request %v", err)
		return
	}

	*isHeartbeatRegistered = true
	log.Printf("[%s] WantHeartbeat: Recived %v\n", conn.RemoteAddr().String(), req)

	if req.Interval == 0 {
		log.Printf("[%s] Recieved 0 inteval req. Heartbeat Disabled", conn.RemoteAddr().String())
		return
	}

	go s.sendHeartbeat(conn, req.Interval)
}

func (s *Server) sendHeartbeat(conn net.Conn, decisecond uint32) {
	interval := time.Duration(decisecond*100) * time.Millisecond
	log.Printf("[%s] Sending heartbeat every %.1f seconds",
		conn.RemoteAddr().String(),
		float64(interval)/float64(time.Second),
	)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	heartbeatMsg := EncodeHeartbeat()
	for {
		select {
		case <-ticker.C:
			_, err := conn.Write(heartbeatMsg)
			if err != nil {
				fmt.Errorf("[%s] Failed to send heartbeat: %v", conn.RemoteAddr().String(), err)
				return
			}
			log.Printf("[%s] Heartbeat sent", conn.RemoteAddr().String())
		}
	}
}
