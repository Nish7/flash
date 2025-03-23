package server

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

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
		return data, fmt.Errorf("Error Parsing CameraRequest %v", err)
	}

	return data, nil
}

func ParseWantHeartbeat(reader *bufio.Reader) (WantHeartbeat, error) {
	var data WantHeartbeat
	err := binary.Read(reader, binary.BigEndian, &data)

	if err != nil {
		return data, fmt.Errorf("Error Parsing WantHeartbeat %x", err)
	}

	return data, nil
}

// used by integration_testing
func ParseTicket(reader *bufio.Reader) (Ticket, error) {
	ticket := Ticket{}

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

func EncodeTicket(ticket *Ticket) []byte {
	plateLen := len(ticket.Plate)
	msg := make([]byte, 1+1+plateLen+16)

	msg[0] = byte(TICKET_RESP)
	msg[1] = byte(plateLen)
	copy(msg[2:2+plateLen], ticket.Plate)
	binary.BigEndian.PutUint16(msg[2+plateLen:4+plateLen], ticket.Road)
	binary.BigEndian.PutUint16(msg[4+plateLen:6+plateLen], ticket.Mile1)
	binary.BigEndian.PutUint32(msg[6+plateLen:10+plateLen], ticket.Timestamp1)
	binary.BigEndian.PutUint16(msg[10+plateLen:12+plateLen], ticket.Mile2)
	binary.BigEndian.PutUint32(msg[12+plateLen:16+plateLen], ticket.Timestamp2)
	binary.BigEndian.PutUint16(msg[16+plateLen:18+plateLen], ticket.Speed)

	return msg
}

func EncodeHeartbeat() []byte {
	return []byte{byte(HEARTBEAT_RESP)}
}

func EncodeError(errMsg string) []byte {
	errLen := len(errMsg)
	msg := make([]byte, 1+1+errLen)

	msg[0] = byte(ERROR_RESP)
	msg[1] = byte(errLen)
	copy(msg[2:2+errLen], errMsg)

	return msg
}

func ParseError(reader *bufio.Reader) (ErrorResp, error) {
	errResp := ErrorResp{}
	errMsg, err := ParseString(reader)
	if err != err {
		return errResp, fmt.Errorf("error reading plate (str): %w", err)
	}
	errResp.Msg = errMsg
	return errResp, nil
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
