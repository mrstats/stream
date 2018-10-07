package logger

import (
	"log"
	"sync"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"stream/internal/pkg/privlib/config"
)

var (
	inst *Logger
	once sync.Once
)

type Logger struct {
	*logrus.Logger
}

func GetInstance() *Logger {
	once.Do(func() {

		Formatter := new(prefixed.TextFormatter)
		Formatter.TimestampFormat = "02-01-2006 15:04:05"
		Formatter.FullTimestamp = true
		logrus.SetFormatter(Formatter)

		inst = &Logger{
			logrus.New(),
		}

		// Use logrus for standard log output
		log.SetOutput(inst.Writer())

		config.GetInstance().OnChange(inst.setLevel)
		inst.setLevel()
	})

	return inst
}

func (l *Logger) setLevel() {
	cfg := config.GetInstance()

	lvl, err := logrus.ParseLevel(cfg.GetString("log.level"))
	if err != nil {
		switch cfg.Mode() {
		case config.ModePro:
			lvl = logrus.ErrorLevel

		case config.ModeStg:
			lvl = logrus.WarnLevel

		case config.ModeDev:
			lvl = logrus.DebugLevel

		default:
			lvl = logrus.ErrorLevel
		}
	}
	l.SetLevel(lvl)
}
