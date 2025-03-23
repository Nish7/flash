package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"slices"
)

func (s *Server) HandleDispatcherReq(conn net.Conn, reader *bufio.Reader, client *ClientType) error {
	if *client != UNKNOWN {
		return fmt.Errorf("[dispatcher] Client is alredy registered on this connection")
	}

	d, err := ParseDispatcherRecord(reader)
	if err != nil {
		return fmt.Errorf("Failed to parse request %v", err)
	}

	log.Printf("[%s] Dispatcher Recived %v\n", conn.RemoteAddr().String(), d)

	s.slock.Lock()
	defer s.slock.Unlock()
	s.dispatchers[conn] = d
	*client = DISPATCHER

	err = s.checkPendingTickets(conn, d)
	if err != nil {
		return fmt.Errorf("Failed to parse request %v", err)
	}

	return nil
}

// TODO: improve the perfomance
func (s *Server) checkPendingTickets(conn net.Conn, d Dispatcher) error {
	log.Printf("[%s] Checking Pending Tickets [%v]\n", conn.RemoteAddr().String(), s.pending_queue)
	var newQueue []Ticket
	var errors []error

	for _, ticket := range s.pending_queue {
		if slices.Contains(d.Roads, ticket.Road) {
			log.Printf("[%s] Dispatcher [%v] is available; Sending ticket [%v]\n", conn.RemoteAddr().String(), d, ticket)
			// TODO: Assuming only one dispatcher per road
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
