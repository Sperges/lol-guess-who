package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	session "github.com/go-session/session/v3"
	"github.com/google/uuid"
)

const TOTAL_CHAMPS = 165
const BOARD_SIZE = 25

type GameState int

const (
	LOBBY GameState = iota
)

// Game maintains the set of active clients and broadcasts messages to the
// clients.
type Game struct {
	id string

	champs []int

	currentHandler func() bool

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	owner map[string]*Game
}

func newGame(owner map[string]*Game) *Game {
	game := &Game{
		id:         uuid.NewString(),
		champs:     generateChamps(),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		owner:      owner,
	}
	game.currentHandler = game.lobby
	owner[game.id] = game
	go game.run()
	return game
}

func generateChamps() []int {
	champs := []int{}
	for i := 0; i < BOARD_SIZE; i++ {
		champs = append(champs, rand.Int()%TOTAL_CHAMPS)
	}
	return champs
}

func (game *Game) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		id:   uuid.NewString(),
		game: game,
		conn: conn,
		send: make(chan []byte, 256),
	}

	game.storeSession(w, r, client)

	client.game.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (game *Game) storeSession(w http.ResponseWriter, r *http.Request, client *Client) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Println("Session creation err:", err)
		return
	}

	store.Set(client.game.id+"-clientID", client.id)
	err = store.Save()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
}

func (game *Game) broadcastMessage(message *Message) {
	for client := range game.clients {
		select {
		case client.send <- message.data:
		default:
			close(client.send)
			delete(game.clients, client)
		}
	}
}

func (game *Game) run() {
	log.Println("Starting game", game.id)
	for {
		if ok := game.currentHandler(); !ok {
			break
		}
	}
	log.Println("Ending game", game.id)
}

func (game *Game) lobby() bool {
	select {
	case client := <-game.register:
		game.lobbyRegister(client)
	case client := <-game.unregister:
		game.lobbyUnregister(client)
	case message := <-game.broadcast:
		log.Println("Message from", message.client.id, message.data)
		game.broadcastMessage(message)
	}
	return true
}

func (game *Game) lobbyRegister(client *Client) bool {
	log.Println("Registering", client.id)
	game.clients[client] = true
	if len(game.clients) == 2 {
		log.Println("Game", game.id, "starting champ select")
		game.currentHandler = game.champSelect
	}
	return true
}

func (game *Game) lobbyUnregister(client *Client) bool {
	log.Println("Unregistering", client.id)
	if _, ok := game.clients[client]; ok {
		delete(game.clients, client)
		close(client.send)
	}
	if len(game.clients) == 0 {
		delete(game.owner, game.id)
		return false
	}
	return true
}

func (game *Game) champSelect() bool {
	select {
	case <-game.register:
	case <-game.unregister:
	case <-game.broadcast:
	}
	return true
}
