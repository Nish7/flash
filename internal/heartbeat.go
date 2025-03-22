package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

func (s *Server) WantHeatbeatHandler(conn net.Conn, reader *bufio.Reader, isHeartbeatRegistered *bool, client *ClientType) error {
	if *client == UNKNOWN {
		return fmt.Errorf("Client is not registered.")
	}

	if *isHeartbeatRegistered {
		return fmt.Errorf("Heartbeat is already registered.")
	}

	req, err := ParseWantHeartbeat(reader)
	if err != nil {
		return fmt.Errorf("Failed to parse request %v", err)
	}

	*isHeartbeatRegistered = true
	log.Printf("[%s] WantHeartbeat: Recived %v\n", conn.RemoteAddr().String(), req)

	if req.Interval == 0 {
		log.Printf("Recieved 0 inteval req. Heartbeat Disabled")
		return nil
	}

	go s.sendHeartbeat(conn, req.Interval)
	return nil
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
				s.ErrorHandler(fmt.Errorf("Failed to send heartbeat: %v\n", err), conn)
				return
			}

			log.Printf("[%s] Heartbeat sent", conn.RemoteAddr().String())
		}
	}
}
