package server

import (
	"encoding/binary"
	"fmt"
	repo "github.com/nish7/flash/internal/repository"
	"log"
	"net"
	"slices"
)

func (s *Server) HandleSpeedViolations(obs repo.Observation, conn net.Conn) {
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

		ticket := &repo.Ticket{
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
		if !CheckTicketLimit(ticket, priorPlateTickets, conn) {
			continue
		}

		s.DispatchTicket(ticket, conn)
	}
}

func (s *Server) DispatchTicket(ticket *repo.Ticket, conn net.Conn) {
	for c, disp := range s.dispatchers {
		if slices.Contains(disp.Roads, ticket.Road) {
			s.store.AddTicket(*ticket)
			err := s.SendTicket(c, ticket)
			if err != nil {
				fmt.Println("Error sending ticket:", err)
			} else {
				log.Printf("[%s] Ticket sent for %s on road %d [%v]\n", conn.RemoteAddr().String(), ticket.Plate, ticket.Road, ticket)
			}
			return
		}
	}

	log.Printf("No Dispatcher Found")
}

func (s *Server) SendTicket(conn net.Conn, ticket *repo.Ticket) error {
	plateLen := len(ticket.Plate)
	msg := make([]byte, 1+1+plateLen+16)

	// TODO: to be improved
	msg[0] = byte(TICKET_RESP)
	msg[1] = byte(plateLen)
	copy(msg[2:2+plateLen], ticket.Plate)
	binary.BigEndian.PutUint16(msg[2+plateLen:4+plateLen], ticket.Road)
	binary.BigEndian.PutUint16(msg[4+plateLen:6+plateLen], ticket.Mile1)
	binary.BigEndian.PutUint32(msg[6+plateLen:10+plateLen], ticket.Timestamp1)
	binary.BigEndian.PutUint16(msg[10+plateLen:12+plateLen], ticket.Mile2)
	binary.BigEndian.PutUint32(msg[12+plateLen:16+plateLen], ticket.Timestamp2)
	binary.BigEndian.PutUint16(msg[16+plateLen:18+plateLen], ticket.Speed)

	_, err := conn.Write(msg)
	return err
}

func isSpeedViolation(obs1, obs2 repo.Observation) (bool, uint16) {
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
func CheckTicketLimit(ticket *repo.Ticket, plateTickets []repo.Ticket, conn net.Conn) bool {
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
