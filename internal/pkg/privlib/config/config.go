package config

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	configName = "servers"
)

var (
	inst *Config
	once sync.Once
)

type Config struct {
	*viper.Viper

	mode      Mode
	modeTitle string

	updateHandlers []UpdateHandler

	mu sync.Mutex
}

func GetInstance() *Config {
	once.Do(func() {
		inst = &Config{
			Viper:          viper.New(),
			updateHandlers: make([]UpdateHandler, 0, 4),
		}

		inst.SetConfigName(configName)

		inst.AddConfigPath(".")
		inst.AddConfigPath("configs/")
		inst.AddConfigPath("deployments/configs/")
		inst.AddConfigPath("/etc/stream/")
		inst.AddConfigPath("$HOME/.stream")

		err := inst.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err)) //TODO: change it to error and handle it in app
		}

		inst.updateHandlers = make([]UpdateHandler, 0, 4)
		inst.OnConfigChange(func(e fsnotify.Event) {
			fmt.Println("Config changed by", e.Name)
			inst.update()
		})
		inst.WatchConfig()

		inst.update()
	})
	return inst
}

func (c *Config) OnChange(f UpdateHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updateHandlers = append(c.updateHandlers, f)
}

func (c *Config) update() {
	c.parseMode()

	for _, f := range c.updateHandlers {
		go f()
	}
}

func (c *Config) parseMode() {

	modStr := strings.ToLower(c.GetString("mode"))
	switch modStr {
	case "prod":
		if !c.IsSet("prod") {
			log.Fatal("can not find prod key") //TODO: change it to error and handle it in app
		}
		c.mode = ModePro
		c.Viper = c.Sub("prod")

	case "stag":
		if !c.IsSet("stag") {
			log.Fatal("can not find stag key") //TODO: change it to error and handle it in app
		}
		c.mode = ModeStg
		c.Viper = c.Sub("stag")

	case "dev":
		if !c.IsSet("dev") {
			log.Fatal("can not find dev key") //TODO: change it to error and handle it in app
		}
		c.mode = ModeDev
		c.Viper = c.Sub("dev")

	default:
		if !c.IsSet("prod") {
			log.Fatal("can not parse config mode and prod key(default mode) not found") //TODO: change it to error and handle it in app
		}
		c.mode = ModePro
		c.Viper = c.Sub("prod")
	}
}

func (c *Config) Mode() (mod Mode) {
	return c.mode
}
