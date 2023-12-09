package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger     LoggerConf     `mapstructure:"logger"`
	Storage    StorageConf    `mapstructure:"storage"`
	DB         DBConf         `mapstructure:"db"`
	HTTPServer HTTPServerConf `mapstructure:"http"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

type StorageConf struct {
	Driver string `mapstructure:"driver"`
}

type DBConf struct {
	DBHost     string `mapstructure:"host"`
	DBPort     int    `mapstructure:"port"`
	DBName     string `mapstructure:"name"`
	DBUsername string `mapstructure:"username"`
	DBPassword string `mapstructure:"password"`
}

type HTTPServerConf struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func NewConfig() *Config {
	v := viper.New()
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
