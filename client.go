package main

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	Sender             *Client             `json:"sender,omitempty"`
	Flip               *Flip               `json:"flip,omitempty"`
	Reveal             *Reveal             `json:"reveal,omitempty"`
	RequestSelectChamp *RequestSelectChamp `json:"requestSelectChamp,omitempty"`
	ChampSelected      *ChampSelected      `json:"champSelected,omitempty"`
	Chat               *Chat               `json:"chat,omitempty"`
	InitialMessage     *InitialMessage     `json:"initialMessage,omitempty"`
	GameStarted        *GameStarted        `json:"gameStarted,omitempty"`
	RequestBoardUpdate *RequestBoardUpdate `json:"requestBoardUpdate,omitempty"`
	BoardUpdate        *BoardUpdate        `json:"boardUpdate,omitempty"`
}

type RequestSelectChamp struct {
	Index int `json:"index"`
}

type ChampSelected struct{}

type Flip struct {
	Index int  `json:"index"`
	Down  bool `json:"down"`
}

type Chat struct {
	Text string `json:"text"`
}

type InitialMessage struct {
	PlayerId string `json:"playerId"`
	Champs   []int  `json:"champs"`
}

type GameStarted struct{}

type RequestBoardUpdate struct{}

type BoardUpdate struct {
	SenderBoard []bool `json:"senderBoard"`
	OtherBoard  []bool `json:"otherBoard"`
}

type Reveal struct {
	Index int `json:"index"`
}

type Client struct {
	id            string
	selectedChamp int
	board         []bool
	game          *Game
	conn          *websocket.Conn
	send          chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.game.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.parseMessage(bytes.TrimSpace(bytes.Replace(message, newline, space, -1)))
	}
}

func (c *Client) parseMessage(data []byte) {
	var message Message
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Println("json error:", err)
		return
	}
	message.Sender = c
	c.game.broadcast <- &message
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
