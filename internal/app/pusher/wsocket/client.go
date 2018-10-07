package wsocket

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"stream/internal/pkg/privlib/config"
	"stream/internal/pkg/privlib/logger"
)

var (
	log = logger.GetInstance()
	cfg = config.GetInstance()

	timeoutWrite = time.Second * cfg.GetDuration("pusher.ws.timeout.write")
	timeoutPing  = time.Second * cfg.GetDuration("pusher.ws.timeout.ping")
	timeoutPong  = timeoutPing * 11 / 10

	//tokenTitle = cfg.GetString("pusher.ws.GetToken.title")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

type Client struct {
	hub   *Hub
	conn  *websocket.Conn
	send  chan *Message
	id    uuid.UUID
	token Token
}

func (c *Client) write() {
	ticker := time.NewTicker(timeoutPing)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			message.ClientID = c.id
			b, _ := json.Marshal(message)
			w.Write(b)

			n := len(c.send)
			for i := 0; i < n; i++ {
				if message, ok := <-c.send; ok {
					b, _ := json.Marshal(message)
					w.Write(b)
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(timeoutPong))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(timeoutPong)); return nil })

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.WithError(err).Debug("websocket connection close unexpected")
			}
			break
		}
	}
}

func (c *Client) GetToken() Token {
	return c.token
}

func verifyToken(r *http.Request) (token Token, err error) {
	vars := mux.Vars(r)
	t, ok := vars["token"]
	if !ok {
		return "", errors.New("invalid GetToken key")
	} else if t == "" {
		return "", errors.New("invalid GetToken value")
	}

	return Token(t), nil
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	token, err := verifyToken(r)
	if err != nil {
		log.WithError(err).Debugf("invalid GetToken at websocket upgrade request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	con, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		log.Println(err)
		return
	}

	c := &Client{
		hub:   hub,
		conn:  con,
		send:  make(chan *Message, 256),
		token: token,
	}
	c.hub.register <- c

	go c.write()
	go c.read()
}
