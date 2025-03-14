package test

import (
	"bufio"
	"testing"

	srv "github.com/nish7/flash/internal"
)

func SetupDispatchers(t *testing.T, addr string, dispatchers ...srv.Dispatcher) []*srv.TCPClient {
	t.Helper()

	clients := make([]*srv.TCPClient, len(dispatchers))
	for i, dispatcher := range dispatchers {
		client := srv.NewTCPClient(addr)
		if err := client.Connect(); err != nil {
			t.Fatalf("Failed to connect dispatcher %d: %v", i, err)
		}

		client.SendIAMDispatcher(dispatcher)
		clients[i] = client
	}
	return clients
}

func SetupCameras(t *testing.T, addr string, cameras ...srv.Camera) []*srv.TCPClient {
	t.Helper()

	clients := make([]*srv.TCPClient, len(cameras))
	for i, camera := range cameras {
		client := srv.NewTCPClient(addr)
		if err := client.Connect(); err != nil {
			t.Fatalf("Failed to connect camera %d: %v", i, err)
		}

		client.SendIAMCamera(camera)
		clients[i] = client
	}
	return clients
}

func ClientCleanUp(t *testing.T, clients ...*srv.TCPClient) {
	t.Helper()
	for _, client := range clients {
		if client != nil {
			client.Disconnect()
		}
	}
}

func AssertTicket(t *testing.T, reader *bufio.Reader, expectedTicket srv.Ticket) {
	t.Helper()

	msgType, _ := reader.ReadByte()
	if msgType != byte(srv.TICKET_RESP) {
		t.Fatalf("Illegal Message Type/Code")
	}

	recievedTicket, err := srv.ParseTicket(reader)
	if err != nil {
		t.Fatalf("Error Parsing Ticket. Wrong Message")
	}

	if recievedTicket != expectedTicket {
		t.Fatalf("ExpectedTicket %v and Got %v", expectedTicket, recievedTicket)
	}
}
