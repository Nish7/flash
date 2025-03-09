package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	quitch      chan struct{}
	listener    net.Listener
	addr        string
	store       Store
	cameras     map[net.Conn]Camera
	dispatchers map[net.Conn]Dispatcher
}

func NewServer(addr string, store Store) *Server {
	return &Server{
		quitch:      make(chan struct{}),
		addr:        addr,
		store:       store,
		cameras:     make(map[net.Conn]Camera),
		dispatchers: make(map[net.Conn]Dispatcher),
	}
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)

	if err != nil {
		return err
	}

	log.Printf("Server Listening on Port %s", s.addr)
	s.listener = l
	go s.Accept()

	<-s.quitch
	defer l.Close()
	return nil
}

func (s *Server) Accept() {
	for {
		conn, err := s.listener.Accept()
		log.Printf("New Connection :%s\n", conn.RemoteAddr().String())

		if err != nil {
			log.Printf("Error: Connection error %e\n", err)
			return
		}

		go s.HandleConnection(conn)
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()
	var client Client = -1
	reader := bufio.NewReader(conn)

	for {
		msgType, err := reader.ReadByte()
		if err == io.EOF {
			log.Println("Connection closed by remote end")
			return
		}

		if err != nil {
			log.Printf("[%s] Error Reading Connection %v", conn.RemoteAddr().String(), err)
			return
		}

		// request handler
		switch MsgType(msgType) {
		case IAMCAMERA_REQ:
			if client != -1 {
				fmt.Printf("Connection already setup")
				continue
			}

			d, err := ParseCameraRequest(reader)
			if err != nil {
				log.Printf("Failed to parse request %v", err)
				return
			}

			client = CAMERA
			s.HandleCameraReq(conn, d)
			defer s.Cleanup(conn, client)
		case IAMDISPATCHER_REQ:
			if client != -1 {
				fmt.Printf("Connection already setup")
				continue
			}

			d, err := ParseDispatcherRecord(reader)
			if err != nil {
				log.Printf("Failed to parse request %v", err)
				return
			}

			client = DISPATCHER
			s.HandleDispatcherReq(conn, d)
			defer s.Cleanup(conn, client)

		case PLATE_REQ:
			if client != CAMERA {
				log.Printf("Invalid Client. Expected Camera")
			}

			d, err := ParsePlateRecord(reader)
			if err != nil {
				log.Printf("Failed to parse Request %v", err)
				return
			}

			s.HandlePlateReq(conn, d)
		default:
			fmt.Printf("Unknown message type: %X\n", msgType)
		}
	}
}

func (s *Server) HandleDispatcherReq(conn net.Conn, req Dispatcher) error {
	log.Printf("[%s] IAMDISPATCHER_REQ: Recived %v\n", conn.RemoteAddr().String(), req)
	s.dispatchers[conn] = req
	return nil
}

func (s *Server) HandleCameraReq(conn net.Conn, req Camera) error {
	log.Printf("[%s] IAMCAMERA: Recived %v\n", conn.RemoteAddr().String(), req)
	s.cameras[conn] = req
	return nil
}

func (s *Server) HandlePlateReq(conn net.Conn, plate Plate) error {
	cam := s.cameras[conn]
	log.Printf("[%s] Plate Record Receieved: %v from Camera %v\n", conn.RemoteAddr().String(), plate, cam)

	observation := Observation{Plate: plate.Plate, Road: cam.Road, Mile: cam.Mile, Timestamp: plate.Timestamp, Limit: cam.Limit}

	s.store.AddObservation(observation)
	s.HandleSpeedViolations(observation, conn)
	return nil
}

func (s *Server) Cleanup(conn net.Conn, client Client) error {
	switch client {
	case CAMERA:
		delete(s.cameras, conn)
	case DISPATCHER:
		delete(s.dispatchers, conn)
	default:
		return fmt.Errorf("Invalid Client type")
	}
	return nil
}
