package wsocket

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"stream/api/go-spec"
	"stream/internal/pkg/privlib/queue"
	"sync"
)

const (
	prefixChannel = "pusher.wsocket."
)

type Worker struct {
	update chan<- *Message

	follow   chan *Token
	unfollow chan *Token

	users map[Token]*queue.Consumer

	mux sync.Mutex
}

func NewWorker(notify chan<- *Message) *Worker {
	return &Worker{
		update:   notify,
		follow:   make(chan *Token),
		unfollow: make(chan *Token),
		users:    make(map[Token]*queue.Consumer),
	}
}

func (w *Worker) Run() {
	for {
		t := *<-w.follow
		if c := w.followQueue(t); c != nil {
			w.followUser(t, c)
			w.users[t].Connect()
		}
	}
}

func (w *Worker) GetFollowChan() (follow, unfollow chan<- *Token) {
	return w.follow, w.unfollow
}

func (w *Worker) followQueue(t Token) *queue.Consumer {
	if _, ok := w.users[t]; ok {
		return nil
	}

	if !nsq.IsValidTopicName(t.String()) {
		return nil
	}

	c, err := queue.GetConsumer(t.String(), prefixChannel+t.String())
	if err != nil {
		log.WithError(err).Error("consumer/reader creation error")
	}

	c.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		resp := activity.Response{}
		json.Unmarshal(message.Body, &resp)
		w.update <- &Message{
			ClientID: uuid.Nil,
			Token:    t,
			Context:  resp,
		}

		return nil
	}))

	return c
}

func (w *Worker) followUser(t Token, consumer *queue.Consumer) {
	w.mux.Lock()
	defer w.mux.Unlock()

	w.users[t] = consumer
}
