package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"slices"
)

func (s *Server) HandleDispatcherReq(conn net.Conn, reader *bufio.Reader, isClientRegistered *bool) {
	defer delete(s.cameras, conn)
	if *isClientRegistered {
		log.Printf("Client is alredy registered on this connection")
		return
	}

	d, err := ParseDispatcherRecord(reader)
	if err != nil {
		log.Printf("Failed to parse request %v", err)
		return
	}

	log.Printf("[%s] Dispatcher Recived %v\n", conn.RemoteAddr().String(), d)
	s.dispatchers[conn] = d
	*isClientRegistered = true

	err = s.checkPendingTickets(conn, d)
	if err != nil {
		log.Printf("Failed to parse request %v", err)
		return
	}

	return
}

// TODO: improve the perfomance
func (s *Server) checkPendingTickets(conn net.Conn, d Dispatcher) error {
	log.Printf("[%s] Checking Pending Tickets [%v]\n", conn.RemoteAddr().String(), s.pending_queue)
	var newQueue []Ticket
	var errors []error

	for _, ticket := range s.pending_queue {
		if slices.Contains(d.Roads, ticket.Road) {
			log.Printf("[%s] Dispatcher [%v] is available; Sending ticket [%v]\n", conn.RemoteAddr().String(), d, ticket)
			if err := s.SendTicket(conn, &ticket); err != nil {
				errors = append(errors, fmt.Errorf("failed to send ticket %v: %w", ticket, err))
				newQueue = append(newQueue, ticket)
				continue
			}
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
