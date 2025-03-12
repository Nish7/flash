package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func (s *Server) HandleCameraReq(conn net.Conn, reader *bufio.Reader) {
	defer delete(s.cameras, conn)

	camera, err := ParseCameraRequest(reader)
	if err != nil {
		log.Printf("Failed to parse request %v", err)
		return
	}

	// register camera
	log.Printf("[%s] Camera: Recived %v\n", conn.RemoteAddr().String(), camera)
	s.cameras[conn] = camera

	err = s.listenForPlates(conn, reader, camera)
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (s *Server) listenForPlates(conn net.Conn, reader *bufio.Reader, camera Camera) error {
	for {
		msgType, err := ReadMsgType(reader)
		if err != nil {
			return err
		}

		if msgType != PLATE_REQ {
			return fmt.Errorf("Illegal Message type")
		}

		plate, err := ParsePlateRecord(reader)
		if err != nil {
			return fmt.Errorf("Failed to Parse Request  %v", err)
		}

		log.Printf("[%s] Plate Record Receieved: %v from Camera %v\n", conn.RemoteAddr().String(), plate, camera)
		err = s.handlePlateReq(conn, camera, plate)
		if err != nil {
			return fmt.Errorf("Failed to Handle Plate Records: %v", err)
		}
	}
}

func (s *Server) handlePlateReq(conn net.Conn, cam Camera, plate Plate) error {
	observation := Observation{Plate: plate.Plate, Road: cam.Road, Mile: cam.Mile, Timestamp: plate.Timestamp, Limit: cam.Limit}
	s.store.AddObservation(observation)

	err := s.handleSpeedViolations(conn, observation)
	if err != nil {
		return err
	}

	return nil
}
