package wsocket

import (
	"github.com/google/uuid"
	"stream/api/go-spec"
)

type Tokener interface {
	GetToken() Token
}

type Token string

func (t Token) String() string {
	return string(t)
}

type Message struct {
	ClientID uuid.UUID         `json:"client_id,string"`
	Token    Token             `json:"token"`
	Context  activity.Response `json:"context"`
}

func (m *Message) GetToken() Token {
	return m.Token
}
