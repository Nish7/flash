package server

// consts
type MsgType byte

const (
	IAMCAMERA_REQ     MsgType = 0x80
	IAMDISPATCHER_REQ MsgType = 0x81
	PLATE_REQ         MsgType = 0x20
	WANTHEARTBEAT_REQ MsgType = 0x40
	TICKET_RESP       MsgType = 0x21
	HEARTBEAT_RESP    MsgType = 0x41
)

// model

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

type Observation struct {
	Plate     string
	Road      uint16
	Mile      uint16
	Timestamp uint32
	Limit     uint16
}

type WantHeartbeat struct {
	interval uint32
}

type Ticket struct {
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 uint32
	Mile2      uint16
	Timestamp2 uint32
	Speed      uint16
}
