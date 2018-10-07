package queue

import (
	stdLog "log"

	"stream/internal/pkg/privlib/config"
	"stream/internal/pkg/privlib/logger"

	"github.com/nsqio/go-nsq"
)

var (
	//pCfg = nsq.NewConfig()
	cCfg = nsq.NewConfig()

	addrNsqd       = "127.0.0.1: 4150"
	addrNsqlookupd = "127.0.0.1:4171"

	//userAgent = fmt.Sprintf(".queue.stream@go-nsq_v%s", nsq.VERSION)

	logLevel = nsq.LogLevelInfo

	cfg = config.GetInstance()
	log = logger.GetInstance()
)

func init() {
	cfg.OnChange(setConfig)
	setConfig()
}

type Producer struct {
	*nsq.Producer
}

func GetProducer() (p *Producer, err error) {
	producer, err := nsq.NewProducer(addrNsqd, cCfg)
	if err != nil {
		return
	}

	p = &Producer{
		producer,
	}
	p.SetLogger(stdLog.New(log.Writer(), "nsq ", 0), logLevel)

	return
}

type Consumer struct {
	*nsq.Consumer
}

func GetConsumer(topic string, channel string) (c *Consumer, err error) {
	consumer, err := nsq.NewConsumer(topic, channel, cCfg)
	if err != nil {
		return
	}

	c = &Consumer{
		consumer,
	}
	c.SetLogger(stdLog.New(log.Writer(), "nsq ", 0), logLevel)

	return
}

func (c *Consumer) Connect() error {
	err := c.ConnectToNSQD()
	if err != nil {
		log.WithError(err).Error("consumer/reader can not connect to queue daemon")
	}
	//err := c.ConnectToLookupd()
	//if err != nil {
	//	log.WithError(err).Error("consumer/reader can not connect to queue lookup daemon")
	//}

	return err
}

func (c *Consumer) ConnectToNSQD() error {
	return c.Consumer.ConnectToNSQD(addrNsqd)
}

func (c *Consumer) ConnectToLookupd() error {
	return c.Consumer.ConnectToNSQLookupd(addrNsqlookupd)
}

func setConfig() {
	addrNsqd = cfg.GetString("queue.addr.nsqd")
	addrNsqlookupd = cfg.GetString("queue.addr.nsqlookupd")

	switch log.Level.String() {
	case "panic", "fatal", "error":
		logLevel = nsq.LogLevelError
	case "warning":
		logLevel = nsq.LogLevelWarning
	case "info":
		logLevel = nsq.LogLevelInfo
	case "debug":
		logLevel = nsq.LogLevelDebug

	}
}
