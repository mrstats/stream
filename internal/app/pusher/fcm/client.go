package fcm

import (
	"encoding/json"
	"fmt"

	"stream/api/go-spec"

	goFcm "github.com/NaySoftware/go-fcm"
	"github.com/nsqio/go-nsq"

	"stream/internal/pkg/privlib/queue"
)

const (
	prefixTopic   = "feed.user."
	prefixChannel = "pusher.fcm."
)

var (
	clients      = make(map[token]client, 0)
	fcmServerKey = cfg.GetString("fcm.key.server")
)

type client struct {
	token         token
	fcmTokens     []string
	subscribeConn *queue.Consumer
}

func (c *client) update(resp activity.Response) {
	fcmClient := goFcm.NewFcmClient(fcmServerKey)
	fcmClient.NewFcmRegIdsMsg(c.fcmTokens, resp)

	fcmClient.SetNotificationPayload(&goFcm.NotificationPayload{
		Title: "test title",
		Body:  "test body",
		Icon:  "https://firebase.google.com/downloads/brand-guidelines/SVG/logo-logomark.svg",
	})

	status, err := fcmClient.Send()
	if err != nil {
		log.WithError(err).Errorf("fcm sending | resp: %s", resp)
	} else if status.StatusCode != 200 {
		log.Errorf("fcm sending err | resp: %+v", resp)
		status.PrintResults()
	} else {
		status.PrintResults()
	}
}

func (c *client) subscribe() error {
	if !nsq.IsValidTopicName(c.token.String()) {
		return fmt.Errorf("invalid topic name")
	}

	conn, err := queue.GetConsumer(prefixTopic+c.token.String(), prefixChannel+c.token.String())
	if err != nil {
		log.WithError(err).Error("consumer/reader creation error")
	}

	conn.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		resp := activity.Response{}
		json.Unmarshal(message.Body, &resp)
		c.update(resp)
		return nil
	}))

	conn.Connect()
	c.subscribeConn = conn

	return nil
}

type token string

func (t token) String() string {
	return string(t)
}
