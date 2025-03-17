package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"slices"
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

	err = s.checkPendingTicket(conn, d)
	if err != nil {
		log.Printf("Failed to parse request %v", err)
		return
	}
	return
}

// TODO: improve the perfomance
func (s *Server) checkPendingTicket(conn net.Conn, d Dispatcher) error {
	var newQueue []Ticket
	var errors []error

	for _, ticket := range s.pending_queue {
		if slices.Contains(d.Roads, ticket.Road) {
			if err := s.SendTicket(conn, &ticket); err != nil {
				errors = append(errors, fmt.Errorf("failed to send ticket %v: %w", ticket, err))
				newQueue = append(newQueue, ticket)
				continue
			}

			addr := conn.RemoteAddr().String()
			fmt.Printf("[%s] Dispatcher [%v] is available; Sending ticket [%v]\n", addr, d, ticket)
		} else {
			newQueue = append(newQueue, ticket)
		}
	}

	s.pending_queue = newQueue
	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors: %v", len(errors), errors)
	}

	return nil
}
