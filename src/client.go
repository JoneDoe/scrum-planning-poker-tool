package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  maxMessageSize,
	WriteBufferSize: maxMessageSize,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

type vote struct {
	Owner string `json:"owner"`
	Score string `json:"score"`
}

type roomResp struct {
	VotingIsOver bool   `json:"isOver"`
	Scores       []vote `json:"scores"`
}

var roomVotes = make(map[string][]vote)

// readPump pumps messages from the websocket connection to the hub.
func (s subscription) readPump() {
	c := s.conn

	defer func() {
		h.unregister <- s
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))

		return nil
	})

	var data vote

	for {
		err := c.ws.ReadJSON(&data)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		data.Owner = s.subscriber

		var exists bool

		/*for i, voter := range roomVotes[s.room] {
			if voter.Owner == data.Owner {
				roomVotes[s.room][i].Score = data.Score
				exists = true
				break
			}
		}*/

		if !exists {
			roomVotes[s.room] = append(roomVotes[s.room], data)
		}

		fmt.Println(roomVotes)

		fmt.Println(h.countSubscribers(s.room))

		msg, _ := json.Marshal(&roomResp{
			h.countSubscribers(s.room) == len(roomVotes[s.room]),
			roomVotes[s.room]})

		h.broadcast <- message{msg, s.room}
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))

	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request, roomId string, context *gin.Context) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	id, _ := context.Cookie("ROOMSESSID")

	c := &connection{send: make(chan []byte, 256), ws: ws}
	s := subscription{c, roomId, id}

	h.register <- s

	go s.writePump()
	go s.readPump()
}
