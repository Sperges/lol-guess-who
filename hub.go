package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/google/uuid"
)

const (
	TOTAL_CHAMPS = 165
	MAX_PICKS    = 24
)

type Hub struct {
	id         string
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return newHubWithId(uuid.NewString())
}

func newHubWithId(id string) *Hub {
	return &Hub{
		id:         id,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) stop() {

	log.Printf("Stopping room %s\n", h.id)
	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
}

func (h *Hub) run(owner map[string]*Hub) {
	log.Printf("Starting room %s\n", h.id)
loop:
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
			if len(h.clients) == 0 {
				h.stop()
				delete(owner, h.id)
				break loop
			}
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	clients := len(h.clients)
	if clients < 2 {
		h.clients[client] = true
		h.broadcastMessage([]byte(fmt.Sprintf("%s connected", client.id)))
		if clients+1 == 2 {
			h.broadcastStartMessage()
		}
	} else {
		client.send <- []byte("room full")
	}
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
}

func (h *Hub) broadcastStartMessage() {
	message, _ := json.Marshal(getStartmessage())

	h.broadcastMessage(message)
}

func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

type StartMessage struct {
	Type   string `json:"type"`
	Champs []int  `json:"champs"`
}

func getStartmessage() *StartMessage {
	champs := []int{}

	for i := 0; i < MAX_PICKS; i++ {
		champs = append(champs, rand.Int()%TOTAL_CHAMPS)
	}

	return &StartMessage{
		Type:   "start",
		Champs: champs,
	}
}
