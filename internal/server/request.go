package server

import (
	"bufio"
	"encoding/binary"
	"fmt"
	repo "github.com/nish7/flash/internal/repository"
	"io"
)

type MsgType byte

const (
	IAMCAMERA_REQ     MsgType = 0x80
	IAMDISPATCHER_REQ MsgType = 0x81
	PLATE_REQ         MsgType = 0x20
	TICKET_RESP       MsgType = 0x21
)

type Client int

const (
	CAMERA = iota
	DISPATCHER
)

type Camera struct {
	Road  uint16
	Mile  uint16
	Limit uint16
}

type Plate struct {
	Plate     string
	Timestamp uint32
}

type Dispatcher struct {
	Roads []uint16
}

func ParseDispatcherRecord(reader *bufio.Reader) (Dispatcher, error) {
	dispatcher := Dispatcher{}
	// read the plate value
	roads, err := ParseIntSlice(reader)
	if err != nil {
		return dispatcher, fmt.Errorf("error reading dispatcher roads %w", err)
	}

	dispatcher.Roads = roads
	return dispatcher, nil
}

func ParseIntSlice(reader *bufio.Reader) ([]uint16, error) {
	// Read length byte
	var lengthByte byte
	if err := binary.Read(reader, binary.BigEndian, &lengthByte); err != nil {
		return nil, fmt.Errorf("error reading length: %w", err)
	}
	sliceLen := int(lengthByte)

	// Read the uint16 slice directly
	result := make([]uint16, sliceLen)
	if err := binary.Read(reader, binary.BigEndian, result); err != nil {
		return nil, fmt.Errorf("error reading uint16 slice: %w", err)
	}

	return result, nil
}

func ParseString(reader *bufio.Reader) (string, error) {
	// read the length byte
	lengthByte, err := reader.ReadByte()
	if err != nil {
		return "", fmt.Errorf("error reading plate length: %w", err)
	}
	stringLen := int(lengthByte)

	// read the string into the buffer
	stringBytes := make([]byte, stringLen)
	n, err := io.ReadFull(reader, stringBytes)
	if err != nil {
		return "", fmt.Errorf("error reading plate string: %w", err)
	}
	if n != stringLen {
		return "", fmt.Errorf("incomplete plate string: expected %d bytes, got %d", stringLen, n)
	}

	return string(stringBytes), nil
}

func ParseCameraRequest(reader *bufio.Reader) (Camera, error) {
	var data Camera
	err := binary.Read(reader, binary.BigEndian, &data)

	if err != nil {
		return data, fmt.Errorf("Error Parsing CameraRequest %x", err)
	}

	return data, nil
}

// used by integration_testing
func ParseTicket(reader *bufio.Reader) (repo.Ticket, error) {
	ticket := repo.Ticket{}

	// read the plate value
	plate, err := ParseString(reader)
	if err != err {
		return ticket, fmt.Errorf("error reading plate (str): %w", err)
	}
	ticket.Plate = plate

	// Read all the fields
	err = binary.Read(reader, binary.BigEndian, &ticket.Road)
	err = binary.Read(reader, binary.BigEndian, &ticket.Mile1)
	err = binary.Read(reader, binary.BigEndian, &ticket.Timestamp1)
	err = binary.Read(reader, binary.BigEndian, &ticket.Mile2)
	err = binary.Read(reader, binary.BigEndian, &ticket.Timestamp2)
	err = binary.Read(reader, binary.BigEndian, &ticket.Speed)

	if err != nil {
		return ticket, fmt.Errorf("error reading timestamp: %w", err)
	}

	return ticket, nil
}

func ParsePlateRecord(reader *bufio.Reader) (Plate, error) {
	plateRecord := Plate{}
	// read the plate value
	plate, err := ParseString(reader)
	if err != err {
		return plateRecord, fmt.Errorf("error reading plate (str): %w", err)
	}
	plateRecord.Plate = plate

	// Read the timestamp (4 bytes).
	err = binary.Read(reader, binary.BigEndian, &plateRecord.Timestamp)
	if err != nil {
		return plateRecord, fmt.Errorf("error reading timestamp: %w", err)
	}

	return plateRecord, nil
}
