package config

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type DbConfig struct {
	Host     string `validate:"required"`
	Port     uint   `validate:"gt=0,required"`
	DbName   string `validate:"required"`
	User     string `validate:"required"`
	Password string `validate:"required"`
}

type RestConfig struct {
	Port uint `validate:"gt=0,required"`
}

type DataConfig struct {
	Dir string `validate:"required"`
}

type ConversionConfig struct {
	Parallelism int `validate:"required"`
}

type Config struct {
	Db         DbConfig         `validate:"dive,required"`
	Rest       RestConfig       `validate:"dive,required"`
	Data       DataConfig       `validate:"dive,required"`
	Conversion ConversionConfig `validate:"dive,required"`
}

func loadConfig() (*Config, error) {
	config := Config{}

	configBytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, err
	}

	validatorInstance := validator.New()
	err = validatorInstance.Struct(config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

var Export = fx.Options(fx.Provide(loadConfig))
