package queuer

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"

	"github.com/gorilla/mux"
	"stream/api/go-spec"
	"stream/internal/pkg/privlib/logger"
	"stream/internal/pkg/privlib/queue"
)

var (
	log = logger.GetInstance()
	//cfg = config.GetInstance()
)

type Queuer struct {
	router   *mux.Router
	producer *queue.Producer
}

func NewQueuer() (q *Queuer) {
	p, err := queue.GetProducer()
	if err != nil {
		log.WithError(err).Fatal("can not get new producer/writer to queue")
	}
	p.Ping()

	return &Queuer{
		router:   mux.NewRouter(),
		producer: p,
	}
}

func (q *Queuer) Route() *mux.Router {
	q.router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, r.Method+" method not allowed", http.StatusMethodNotAllowed)
	})
	q.router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad url", http.StatusNotFound)
	})

	api := q.router.
		PathPrefix("/api")

	v1 := api.
		Methods(http.MethodGet, http.MethodPost).
		Headers("Content-Type", "application/json").
		Queries("api_key", "{api_key}").
		PathPrefix("/v1.0").
		Name("v1").
		Subrouter()

	v1Feed := v1.
		PathPrefix("/feed/{feed_slug}/{feed_entity_id}").
		Name("v1Feed").
		Subrouter()

	v1Feed.
		Methods(http.MethodGet).
		HandlerFunc(q.feedGet).
		Queries(
			"limit", "{limit}",
			"offset", "{offset}",
		).
		Name("v1FeedGet")

	v1Feed.
		Methods(http.MethodPost).
		HandlerFunc(q.feedPost).
		Name("v1FeedPost")

	return q.router
}

func (q *Queuer) feedGet(w http.ResponseWriter, r *http.Request) {
	q.queue("websocket", []byte("hi nsq"))
}

func (q *Queuer) feedPost(w http.ResponseWriter, r *http.Request) {
	req := activity.Request{}
	b, err := q.reqBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(b, &req)
	if err != nil {
		log.WithError(err).Errorf("bad request: %s", string(b))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	feedSlug := vars["feed_slug"]
	feedEntityID := vars["feed_entity_id"]
	feedID := "feed." + feedSlug + "." + feedEntityID
	if !nsq.IsValidTopicName(feedID) {
		log.WithError(err).Errorf("invalid 'slug' or 'entity' name")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Time.IsZero() {
		req.Time = time.Now()
	}
	id, _ := uuid.NewRandom()
	res := activity.Response{
		ID:      id,
		Request: req,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

	j, _ := json.Marshal(res)
	q.queue(feedID, j)
}

func (q *Queuer) queue(topic string, msg []byte) {
	q.producer.Publish(topic, msg)
}

func (q *Queuer) reqBody(r *http.Request) (b []byte, err error) {
	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Errorf("can not read body")
		return
	}
	r.Body.Close()
	return
}

func (q *Queuer) Shutdown() {

}
