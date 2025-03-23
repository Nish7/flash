package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func (s *Server) HandleCameraReq(conn net.Conn, reader *bufio.Reader, clientType *ClientType) error {
	if *clientType != UNKNOWN {
		return fmt.Errorf("Client is already registered.")
	}

	camera, err := ParseCameraRequest(reader)
	if err != nil {
		return fmt.Errorf("Failed to parse request %v", err)
	}

	// register camera
	log.Printf("[%s] Camera: Recived %v\n", conn.RemoteAddr().String(), camera)
	s.slock.Lock()
	s.cameras[conn] = camera
	s.slock.Unlock()
	*clientType = CAMERA

	return nil
}

func (s *Server) HandlePlateReq(conn net.Conn, reader *bufio.Reader, client *ClientType) error {
	if *client != CAMERA {
		return fmt.Errorf("Camera not registered yet for plate request")
	}

	plate, err := ParsePlateRecord(reader)
	if err != nil {
		return err
	}

	s.slock.Lock()
	cam, ok := s.cameras[conn]
	if !ok {
		s.slock.Unlock()
		return err
	}
	s.slock.Unlock()

	log.Printf("[%s] Plate Request Recieved %v on Road %d Mile %d Limit %d", conn.RemoteAddr().String(), plate, cam.Road, cam.Mile, cam.Limit)

	observation := Observation{Plate: plate.Plate, Road: cam.Road, Mile: cam.Mile, Timestamp: plate.Timestamp, Limit: cam.Limit}
	s.store.AddObservation(observation)

	s.slock.Lock()
	defer s.slock.Unlock()
	err = s.handleSpeedViolations(conn, observation)
	if err != nil {
		return fmt.Errorf("Failed to Handle Plate Records: %v", err)
	}

	return nil
}
