package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	ErrReadConfig       = fmt.Errorf("unable to read config")
	ErrUnmarshallConfig = fmt.Errorf("unable to unmarshall config")
)

type Config struct {
	Env        string `mapstructure:"APP_ENV"`
	Port       string `mapstructure:"APP_PORT"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBUsername string `mapstructure:"DB_USERNAME"`
	DBDatabase string `mapstructure:"DB_DATABASE"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
}

func LoadConfig(path string) (c *Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()

	if err != nil {
		return nil, ErrReadConfig
	}

	err = viper.Unmarshal(&c)

	if err != nil {
		return nil, ErrUnmarshallConfig
	}

	return
}
