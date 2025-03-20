package server

import (
	"bufio"
	"log"
	"net"
)

func (s *Server) HandleCameraReq(conn net.Conn, reader *bufio.Reader, clientType *ClientType) {
	if *clientType != UNKNOWN {
		log.Printf("[%s] Client is already registered.", conn.RemoteAddr().String())
		return
	}

	camera, err := ParseCameraRequest(reader)
	if err != nil {
		log.Printf("Failed to parse request %v", err)
		return
	}

	// register camera
	log.Printf("[%s] Camera: Recived %v\n", conn.RemoteAddr().String(), camera)
	s.slock.Lock()
	s.cameras[conn] = camera
	s.slock.Unlock()
	*clientType = CAMERA

	return
}

func (s *Server) HandlePlateReq(conn net.Conn, reader *bufio.Reader, client *ClientType) {
	if *client != CAMERA {
		log.Printf("Camera not registered yet for plate request")
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
