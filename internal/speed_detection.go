package server

import (
	"fmt"
	"log"
	"net"
	"slices"
)

func (s *Server) handleSpeedViolations(conn net.Conn, obs Observation) error {
	log.Printf("[%s] Prior Observations [%s]: %v", conn.RemoteAddr().String(), obs.Plate, s.store.GetObservations(obs.Plate))

	// for all prior observation check any speed violations
	for _, preObs := range s.store.GetObservations(obs.Plate) {
		if preObs.Road != obs.Road || preObs.Timestamp == obs.Timestamp {
			continue
		}

		obs1 := preObs
		obs2 := obs
		if obs1.Timestamp > obs2.Timestamp {
			obs1, obs2 = obs2, obs1
		}

		isSpeedViolation, speed := isSpeedViolation(obs1, obs2)
		log.Printf("[%s] isSpeedViolation[%v] - %v\n", conn.RemoteAddr().String(), isSpeedViolation, speed)

		if !isSpeedViolation {
			continue
		}

		ticket := &Ticket{
			Plate:      obs1.Plate,
			Road:       obs1.Road,
			Mile1:      obs1.Mile,
			Timestamp1: obs1.Timestamp,
			Mile2:      obs2.Mile,
			Timestamp2: obs2.Timestamp,
			Speed:      speed,
		}

		priorPlateTickets := s.store.GetTickets(obs.Plate)
		log.Printf("[%s] Prior Plate Tickets [%s]: %v", conn.RemoteAddr().String(), obs.Plate, priorPlateTickets)
		if !CheckTicketLimit(conn, ticket, priorPlateTickets) {
			continue
		}

		err := s.DispatchTicket(conn, ticket)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) DispatchTicket(conn net.Conn, ticket *Ticket) error {
	for c, disp := range s.dispatchers {
		if slices.Contains(disp.Roads, ticket.Road) {
			s.store.AddTicket(*ticket)
			err := s.SendTicket(c, ticket)
			if err != nil {
				return err
			} else {
				log.Printf("[%s] Ticket sent for %s on road %d [%v]\n", conn.RemoteAddr().String(), ticket.Plate, ticket.Road, ticket)
				return nil
			}
		}
	}

	return fmt.Errorf("No Dispatcher Found")
}

func (s *Server) SendTicket(conn net.Conn, ticket *Ticket) error {
	_, err := conn.Write(EncodeTicket(ticket))
	return err
}

func isSpeedViolation(obs1, obs2 Observation) (bool, uint16) {
	distance := uint32(obs2.Mile - obs1.Mile)
	time := obs2.Timestamp - obs1.Timestamp // unix timestamp -> seconds
	if time == 0 {
		return false, 0
	}

	speed := uint16((distance * 3600 * 100) / uint32(time))
	limit := obs1.Limit

	if speed < limit*100+50 {
		return false, speed
	}

	return true, speed
}

// implementing multi-day limit and with one limit per day
func CheckTicketLimit(conn net.Conn, ticket *Ticket, plateTickets []Ticket) bool {
	day1 := ticket.Timestamp1 / 86400
	day2 := ticket.Timestamp2 / 86400

	// check one ticket per day
	for _, t := range plateTickets {
		if t.Timestamp1 == day1 || day1 == t.Timestamp2 || day2 == t.Timestamp1 || day2 == t.Timestamp2 {
			log.Printf("[%s] Ticket Already Exist for Timestamp [%d or %d]\n", conn.RemoteAddr().String(), day1, day2)
			return false
		}
	}

	return true
}
