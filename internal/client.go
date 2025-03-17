package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type TCPClient struct {
	address string
	Conn    net.Conn
}

func NewTCPClient(address string) *TCPClient {
	return &TCPClient{
		address: address,
	}
}

func (c *TCPClient) Connect() error {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", c.address, err)
	}
	c.Conn = conn
	return nil
}

func (c *TCPClient) Disconnect() {
	c.Conn.Close()
}

func (c *TCPClient) SendWantHeartbeat(wh WantHeartbeat) {
	msg := make([]byte, 1+4)
	msg[0] = byte(WANTHEARTBEAT_REQ)
	binary.BigEndian.PutUint32(msg[1:], wh.Interval)

	_, err := c.Conn.Write(msg)
	if err != nil {
		fmt.Printf("Error: sending WantHeartbeat message: %v\n", err)
		return
	}

	fmt.Printf("Client -> Sent WantHeartbeat: %v - hex[% X]\n", wh, msg)
}

func (c *TCPClient) SendPlateRecord(plate Plate) {
	plateBytes := []byte(plate.Plate)
	plateLen := len(plateBytes)

	msg := make([]byte, 1+1+plateLen+4)

	msg[0] = byte(PLATE_REQ)
	msg[1] = byte(plateLen)

	copy(msg[2:2+plateLen], plateBytes)
	binary.BigEndian.PutUint32(msg[2+plateLen:], plate.Timestamp)

	_, err := c.Conn.Write(msg)
	if err != nil {
		fmt.Printf("Error sending PlateRecord message: %v\n", err)
		return
	}

	fmt.Printf("Client -> Sent PlateRecord: %v - hex[% X]\n", plate, msg)
}

// TODO: return error as well
func (c *TCPClient) SendIAMCamera(cam Camera) {
	msg := make([]byte, 7)
	msg[0] = byte(IAMCAMERA_REQ)
	binary.BigEndian.PutUint16(msg[1:3], cam.Road)
	binary.BigEndian.PutUint16(msg[3:5], cam.Mile)
	binary.BigEndian.PutUint16(msg[5:7], cam.Limit)

	_, err := c.Conn.Write(msg)

	if err != nil {
		fmt.Printf("Error sending IAmCamera message")
		return
	}

	fmt.Printf("Client -> Sent CameraRequest: %v - hex[% X]\n", cam, msg)
}

func (c *TCPClient) SendIAMDispatcher(disp Dispatcher) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, byte(IAMDISPATCHER_REQ)); err != nil {
		fmt.Printf("Error writing request type: %v\n", err)
		return
	}

	// Write length of roads slice
	if err := binary.Write(&buf, binary.BigEndian, uint8(len(disp.Roads))); err != nil {
		fmt.Printf("Error writing roads length: %v\n", err)
		return
	}

	// Write the roads slice directly
	if err := binary.Write(&buf, binary.BigEndian, disp.Roads); err != nil {
		fmt.Printf("Error writing roads: %v\n", err)
		return
	}

	// Send the buffer over the connection
	msg := buf.Bytes()
	_, err := c.Conn.Write(msg)
	if err != nil {
		fmt.Printf("Error sending IAmDispatcher message\n")
		return
	}

	fmt.Printf("Client -> Sent DispatcherRequest: %v - hex[% X]\n", disp, msg)
}
