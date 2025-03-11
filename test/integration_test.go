package test

import (
	"bufio"
	"log"
	"os"
	"testing"
	"time"

	fc "github.com/nish7/flash/internal/client"
	repo "github.com/nish7/flash/internal/repository"
	srv "github.com/nish7/flash/internal/server"
)

var testServer *srv.Server
var addr string = ":8080"

func TestMain(t *testing.M) {
	// setup the server
	store := repo.NewInMemoryStore()
	testServer = srv.NewServer(addr, store)

	go func() {
		err := testServer.Start()
		if err != nil {
			log.Fatalf("Error: Starting the server %v", err)
		}
	}()

	time.Sleep(1000 * time.Millisecond) // give some time to start the server
	code := t.Run()
	os.Exit(code)
}

func TestSimpleTicketGeneration(t *testing.T) {
	// setup clients
	dispatchers := []srv.Dispatcher{{Roads: []uint16{123, 2}}}
	dispatcherClients := SetupDispatchers(t, dispatchers...)
	d1 := dispatcherClients[0]

	cameras := []srv.Camera{
		{Road: 123, Mile: 8, Limit: 60},
		{Road: 123, Mile: 9, Limit: 60},
	}
	cameraClients := SetupCameras(t, cameras...)
	c1, c2 := cameraClients[0], cameraClients[1]

	defer ClientCleanUp(t, cameraClients...)
	defer ClientCleanUp(t, dispatcherClients...)

	// Send Plate Observations
	c1.SendPlateRecord(srv.Plate{Plate: "UN1X", Timestamp: 0})
	c2.SendPlateRecord(srv.Plate{Plate: "UN1X", Timestamp: 45})

	time.Sleep(2000 * time.Millisecond)

	// assert the ticket
	reader := bufio.NewReader(d1.Conn)
	expectedTicket := repo.Ticket{Plate: "UN1X", Road: 123, Mile1: 8, Mile2: 9, Timestamp1: 0, Timestamp2: 45, Speed: 8000}
	AssertTicket(t, reader, expectedTicket)
}

func TestPlateRequest(t *testing.T) {
	client := fc.NewTCPClient(addr)
	client.Connect()
	defer client.Disconnect()

	client.SendIAMCamera(srv.Camera{Road: 20, Mile: 80, Limit: 100})
	client.SendPlateRecord(srv.Plate{Plate: "UN1X", Timestamp: 1000})

	time.Sleep(500 * time.Millisecond) // test ended before verifying
}

func TestDispatcherRequest(t *testing.T) {
	client := fc.NewTCPClient(addr)
	client.Connect()
	defer client.Disconnect()

	client.SendIAMDispatcher(srv.Dispatcher{Roads: []uint16{66}})
	time.Sleep(500 * time.Millisecond) // test ended before verifying
}

func TestCameraRequest(t *testing.T) {
	client := fc.NewTCPClient(addr)
	client.Connect()
	defer client.Disconnect()

	client.SendIAMCamera(srv.Camera{Road: 66, Mile: 100, Limit: 60})
	time.Sleep(500 * time.Millisecond) // test ended before verifying
}

func SetupDispatchers(t *testing.T, dispatchers ...srv.Dispatcher) []*fc.TCPClient {
	t.Helper()

	clients := make([]*fc.TCPClient, len(dispatchers))
	for i, dispatcher := range dispatchers {
		client := fc.NewTCPClient(addr)
		if err := client.Connect(); err != nil {
			t.Fatalf("Failed to connect dispatcher %d: %v", i, err)
		}

		client.SendIAMDispatcher(dispatcher)
		clients[i] = client
	}
	return clients
}

func SetupCameras(t *testing.T, cameras ...srv.Camera) []*fc.TCPClient {
	t.Helper()

	clients := make([]*fc.TCPClient, len(cameras))
	for i, camera := range cameras {
		client := fc.NewTCPClient(addr)
		if err := client.Connect(); err != nil {
			t.Fatalf("Failed to connect camera %d: %v", i, err)
		}

		client.SendIAMCamera(camera)
		clients[i] = client
	}
	return clients
}

func ClientCleanUp(t *testing.T, clients ...*fc.TCPClient) {
	t.Helper()
	for _, client := range clients {
		if client != nil {
			client.Disconnect()
		}
	}
}

func AssertTicket(t *testing.T, reader *bufio.Reader, expectedTicket repo.Ticket) {
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
