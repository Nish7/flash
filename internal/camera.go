package server

import (
	"bufio"
	"log"
	"net"
)

func (s *Server) HandleCameraReq(conn net.Conn, reader *bufio.Reader, isClientRegistered *bool) {
	defer delete(s.cameras, conn) // fix this
	if *isClientRegistered {
		return
	}

	camera, err := ParseCameraRequest(reader)
	if err != nil {
		log.Printf("Failed to parse request %v", err)
		return
	}

	// register camera
	log.Printf("[%s] Camera: Recived %v\n", conn.RemoteAddr().String(), camera)
	s.cameras[conn] = camera
	*isClientRegistered = true

	return
}

func (s *Server) HandlePlateReq(conn net.Conn, reader *bufio.Reader, isClientRegistered *bool) {
	if !*isClientRegistered {
		log.Printf("Client not registered yet for plate request")
		return
	}

	plate, err := ParsePlateRecord(reader)
	if err != nil {
		log.Printf("Failed to Parse Request  %v", err)
		return
	}

	cam, ok := s.cameras[conn]
	if !ok {
		log.Printf("Camera not found\n")
		return
	}

	log.Printf("[%s] Plate Record Receieved: %v from Camera %v\n", conn.RemoteAddr().String(), plate, cam)

	observation := Observation{Plate: plate.Plate, Road: cam.Road, Mile: cam.Mile, Timestamp: plate.Timestamp, Limit: cam.Limit}
	s.store.AddObservation(observation)

	err = s.handleSpeedViolations(conn, observation)
	if err != nil {
		log.Printf("Failed to Handle Plate Records: %v", err)
		return
	}
}
