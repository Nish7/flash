package main

import (
	"sync"
)

type Store interface {
	AddObservation(observation Observation)
	GetObservations(plate string) []Observation
	GetTickets(plate string) []Ticket
	AddTicket(ticket Ticket)
}

type Observation struct {
	Plate     string
	Road      uint16
	Mile      uint16
	Timestamp uint32
	Limit     uint16
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

type InMemoryStore struct {
	observations map[string][]Observation // plate (str) -> Observations[]
	tickets      map[string][]Ticket      // plate -> day -> tickets
	mu           sync.Mutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		observations: make(map[string][]Observation),
		tickets:      make(map[string][]Ticket),
	}
}

func (db *InMemoryStore) AddObservation(observation Observation) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.observations[observation.Plate] = append(db.observations[observation.Plate], observation)
}

func (db *InMemoryStore) AddTicket(ticket Ticket) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.tickets[ticket.Plate] = append(db.tickets[ticket.Plate], ticket)
}

func (db *InMemoryStore) GetTickets(plate string) []Ticket {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.tickets[plate]
}

func (db *InMemoryStore) GetObservations(plate string) []Observation {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.observations[plate]
}
