package config

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type DbConfig struct {
	Host     string `validate:"required" yaml:"host"`
	Port     uint   `validate:"gt=0,required" yaml:"port"`
	DbName   string `validate:"required" yaml:"dbName"`
	Username string `validate:"required" yaml:"username"`
	Password string `validate:"required" yaml:"password"`
}

type RestConfig struct {
	Port uint `validate:"gt=0,required" yaml:"port"`
}

type DataConfig struct {
	Dir string `validate:"required" yaml:"dir"`
}

type ConversionConfig struct {
	Parallelism   int    `validate:"required" yaml:"parallelism"`
	PrefAudioLang string `validate:"required" yaml:"prefAudioLang"`
	PrefSubLang   string `validate:"required" yaml:"prefSubLang"`
	WordsPath     string `validate:"required" yaml:"wordsPath"`
}

type Config struct {
	Db         DbConfig         `validate:"dive,required" yaml:"db"`
	Rest       RestConfig       `validate:"dive,required" yaml:"rest"`
	Data       DataConfig       `validate:"dive,required" yaml:"data"`
	Conversion ConversionConfig `validate:"dive,required" yaml:"conversion"`
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
