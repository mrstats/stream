package wsocket

import (
	"github.com/google/uuid"
	"stream/api/go-spec"
)

type Hub struct {
	broadcast chan *Message

	register   chan *Client
	unregister chan *Client

	follow   chan<- *Token
	unfollow chan<- *Token

	tokenClients map[Token][]*Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:    make(chan *Message),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		tokenClients: make(map[Token][]*Client),
	}
}

func (h *Hub) GetBroadcastChan() (broadcast chan<- *Message) {
	return h.broadcast
}

func (h *Hub) SetFallowChan(fallow, unfollow chan<- *Token) {
	h.follow = fallow
	h.unfollow = unfollow
}

func (h *Hub) Run() {

	for {
		select {
		case client := <-h.register:
			t := client.token
			c := h.tokenClients[t]
			client.id = uuid.New()
			h.tokenClients[t] = append(c, client)

			h.follow <- &t

			client.send <- &Message{
				ClientID: client.id,
				Token:    client.token,
				Context:  activity.Response{},
			}

		case client := <-h.unregister:
			t := client.token
			close(client.send)
			h.removeClientByClient(t, client)

		case message := <-h.broadcast:
			t := message.GetToken()
			if clients, ok := h.tokenClients[t]; ok {
				for i, client := range clients {
					select {
					case client.send <- message:
						log.
							WithField("GetToken", client.token.String()).
							WithField("id", client.id.String()).
							Debug("wsockt hub broadcast message to websocket client.")
					default:
						close(client.send)
						h.removeClientByIndex(t, i)
					}
				}
			}

			if len(h.tokenClients[t]) == 0 {
				h.unfollow <- &t
			}
		}
	}
}

func (h *Hub) removeClientByClient(t Token, c *Client) {
	if clients, ok := h.tokenClients[t]; ok {
		for i, client := range clients {
			if client.id == c.id {
				h.removeClientByIndex(t, i)
				break
			}
		}
	}
}

func (h *Hub) removeClientByIndex(t Token, i int) {
	if c, ok := h.tokenClients[t]; ok {
		lMinus := len(c) - 1
		if i <= lMinus {
			c[lMinus], c[i] = c[i], c[lMinus]
			h.tokenClients[t] = c[:lMinus]
		}
	}
}
