package fcm

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"stream/internal/pkg/privlib/config"
	"stream/internal/pkg/privlib/logger"

	"github.com/nsqio/go-nsq"
)

var (
	log = logger.GetInstance()
	cfg = config.GetInstance()
)

type request struct {
	Token    string `json:"token"`
	FCMToken string `json:"fcm_token"`
}

func RegisterClient(w http.ResponseWriter, r *http.Request) {
	req, err := reqFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	feedID := "feed.user." + req.Token
	if !nsq.IsValidTopicName(feedID) {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	reqToken := token(req.Token)
	if c, ok := clients[reqToken]; ok {
		for _, fcmToken := range c.fcmTokens {
			if fcmToken == req.FCMToken {
				errMsg := "fcm token registered before"
				log.Errorf(errMsg)
				http.Error(w, errMsg, http.StatusConflict)
				return
			}
		}

		c.fcmTokens = append(c.fcmTokens, req.FCMToken)
		clients[reqToken] = c
	} else {
		c := client{
			token:     reqToken,
			fcmTokens: []string{req.FCMToken},
		}

		err := c.subscribe()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		clients[reqToken] = c
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("registered"))
}

func UnregisterClient(w http.ResponseWriter, r *http.Request) {
	req, err := reqFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reqToken := token(req.Token)
	if c, ok := clients[reqToken]; ok {
		for k, fcmToken := range c.fcmTokens {
			if fcmToken == req.FCMToken {
				c.fcmTokens = append(c.fcmTokens[:k], c.fcmTokens[k+1:]...)
				clients[reqToken] = c

				w.Write([]byte("unregistered"))
				return
			}
		}

		http.Error(w, "fcm token not found", http.StatusBadRequest)
	}

	http.Error(w, "token not found", http.StatusBadRequest)
}

func reqFromBody(r *http.Request) (req request, err error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Errorf("can not read body")
		return
	}
	r.Body.Close()

	err = json.Unmarshal(b, &req)
	if err != nil {
		log.WithError(err).Errorf("bad request: %s", string(b))
		return
	}

	return
}

func UnsubscribeAll() {
	for _, c := range clients {
		c.subscribeConn.Stop()
	}
}
