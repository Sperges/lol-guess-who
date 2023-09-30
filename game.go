package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

const TOTAL_CHAMPS = 165
const BOARD_SIZE = 25

type GameState int

const (
	LOBBY GameState = iota
	PLAYING
	POST_GAME
)

// Game maintains the set of active clients and broadcasts messages to the
// clients.
type Game struct {
	id string

	champs []int

	// Registered clients.
	clients map[*Client]bool

	state GameState

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
		state:      LOBBY,
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		owner:      owner,
	}
	owner[game.id] = game
	go game.run()
	return game
}

func generateChamps() []int {
	champs := []int{}
	for len(champs) < BOARD_SIZE {
		champ := rand.Int() % TOTAL_CHAMPS
		if !slices.Contains(champs, champ) {
			champs = append(champs, champ)
		}
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
		id:            uuid.NewString(),
		selectedChamp: -1,
		board:         make([]bool, BOARD_SIZE),
		game:          game,
		conn:          conn,
		send:          make(chan []byte, 256),
	}

	client.game.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (game *Game) run() {
	log.Println("Starting game", game.id)
loop:
	for {
		select {
		case client := <-game.register:
			game.registerClient(client)
			game.sendMessage(client, &Message{
				InitialMessage: &InitialMessage{
					PlayerId: client.id,
					Champs:   shuffleSlice(game.champs),
				},
			})
		case client := <-game.unregister:
			game.unregisterClient(client)
			if game.isEmpty() {
				break loop
			}
		case message := <-game.broadcast:
			log.Println(message)
			if err := game.handleIncomingMessage(message); err != nil {
				break loop
			}
		}
	}
	log.Println("Ending game", game.id)
	delete(game.owner, game.id)
}

func (game *Game) handleIncomingMessage(message *Message) error {
	if message.Chat != nil {
		game.broadcastMessage(&Message{Chat: message.Chat})
	}

	if message.RequestSelectChamp != nil && game.state == LOBBY {
		message.Sender.selectedChamp = message.RequestSelectChamp.Index
		game.sendMessageToOthers(message.Sender.id, &Message{ChampSelected: &ChampSelected{}})
		if game.shouldStart() {
			game.broadcastMessage(&Message{
				GameStarted: &GameStarted{},
			})
			game.state = PLAYING
			log.Println(game.id, "now playing")
		}
	}

	if message.Reveal != nil && game.state == PLAYING {
		game.sendMessageToOthers(message.Sender.id, &Message{
			Reveal: &Reveal{
				Index: message.Sender.selectedChamp,
			},
		})
		game.state = POST_GAME
		log.Println(game.id, "finished playing")
	}

	if message.Flip != nil && game.state != LOBBY {
		if message.Flip.Index > 0 && message.Flip.Index < BOARD_SIZE {
			message.Sender.board[message.Flip.Index] = message.Flip.Down
		}
		game.sendMessageToOthers(message.Sender.id, &Message{
			Flip: message.Flip,
		})
	}

	if message.RequestBoardUpdate != nil && game.state != LOBBY {
		var senderBoard []bool
		var otherBoard []bool
		for client := range game.clients {
			if client.id == message.Sender.id {
				senderBoard = client.board
			} else {
				otherBoard = client.board
			}
		}
		game.sendMessage(message.Sender, &Message{
			BoardUpdate: &BoardUpdate{
				SenderBoard: senderBoard,
				OtherBoard:  otherBoard,
			},
		})
	}

	return nil
}

func (game *Game) registerClient(client *Client) bool {
	log.Println("Registering", client.id)
	if !game.isFull() {
		game.clients[client] = true
	}
	return true
}

func (game *Game) unregisterClient(client *Client) bool {
	log.Println("Unregistering", client.id)
	if _, ok := game.clients[client]; ok {
		delete(game.clients, client)
		close(client.send)
	}
	return true
}

func (game *Game) sendMessageToOthers(id string, message any) error {
	for client := range game.clients {
		if id != client.id {
			if err := game.sendMessage(client, message); err != nil {
				return err
			}
		}
	}
	return nil
}

func (game *Game) sendMessage(client *Client, message any) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println("json error:", err)
		return err
	}
	game.sendBytes(client, bytes)
	return nil
}

func (game *Game) broadcastMessage(message any) error {
	for client := range game.clients {
		if err := game.sendMessage(client, message); err != nil {
			return err
		}
	}
	return nil
}

func (game *Game) sendBytes(client *Client, message []byte) {
	select {
	case client.send <- message:
	default:
		close(client.send)
		delete(game.clients, client)
	}
}

// func (game *Game) broadcastBytes(message []byte) {
// 	for client := range game.clients {
// 		game.sendBytes(client, message)
// 	}
// }

func (game *Game) isEmpty() bool {
	return len(game.clients) == 0
}

func (game *Game) isFull() bool {
	return len(game.clients) == 2
}

func (game *Game) shouldStart() bool {
	isFull := game.isFull()
	allSelected := game.allClientsSelected()
	log.Println(game.id, "should start:", isFull, allSelected)
	return isFull && allSelected
}

func (game *Game) allClientsSelected() bool {
	for client := range game.clients {
		if client.selectedChamp == -1 {
			return false
		}
	}
	return true
}

func shuffleSlice(slice []int) []int {
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}
