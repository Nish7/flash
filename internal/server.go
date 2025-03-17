package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	quitch        chan struct{}
	listener      net.Listener
	addr          string
	store         Store
	cameras       map[net.Conn]Camera
	dispatchers   map[net.Conn]Dispatcher
	pending_queue []Ticket
}

func NewServer(addr string, store Store) *Server {
	return &Server{
		quitch:        make(chan struct{}),
		addr:          addr,
		store:         store,
		cameras:       make(map[net.Conn]Camera),
		dispatchers:   make(map[net.Conn]Dispatcher),
		pending_queue: []Ticket{},
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
	reader := bufio.NewReader(conn)
	heartbeatRegistered := false

	for {
		msgType, err := ReadMsgType(reader)
		if err != nil {
			log.Printf("Connection Error: %v", err)
			return
		}

		switch msgType {
		case IAMCAMERA_REQ:
			s.HandleCameraReq(conn, reader)
		case IAMDISPATCHER_REQ:
			s.HandleDispatcherReq(conn, reader)
		case WANTHEARTBEAT_REQ:
			s.WantHeatbeatHandler(conn, reader, &heartbeatRegistered)
		default:
			fmt.Printf("Unknown message type: %X\n", msgType)
		}
	}
}

func ReadMsgType(reader *bufio.Reader) (MsgType, error) {
	msgType, err := reader.ReadByte()
	if err == io.EOF {
		return 0, fmt.Errorf("Connection closed by remote end")
	}

	if err != nil {
		return 0, fmt.Errorf("Unknown Error %v", err)
	}

	return MsgType(msgType), nil
}
