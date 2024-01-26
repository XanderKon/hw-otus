package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger LoggerConf `mapstructure:"logger"`
	Sender SenderConf `mapstructure:"sender"`
	Rmq    RMQConf    `mapstructure:"rmq"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

type SenderConf struct {
	Threads int `mapstructure:"threads"`
}

type ExchangeConf struct {
	Name       string `mapstructure:"name"`
	Type       string `mapstructure:"type"`
	QueueName  string `mapstructure:"queueName"`
	BindingKey string `mapstructure:"bindingKey"`
}

type RMQConf struct {
	URI             string        `mapstructure:"uri"`
	ConsumerTag     string        `mapstructure:"consumerTag"`
	MaxElapsedTime  string        `mapstructure:"maxElapsedTime"`
	InitialInterval string        `mapstructure:"initialInterval"`
	Multiplier      int           `mapstructure:"multiplier"`
	MaxInterval     time.Duration `mapstructure:"maxInterval"`
	Exchange        ExchangeConf
}

func NewConfig() *Config {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("couldn't load config: %s", err)
		os.Exit(1)
	}
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		fmt.Printf("couldn't read config: %s", err)
	}

	return &config
}
