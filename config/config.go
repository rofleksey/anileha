package config

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type DbConfig struct {
	Host     string `validate:"required" yaml:"host"`
	Port     uint   `validate:"required" yaml:"port"`
	DbName   string `validate:"required" yaml:"dbName"`
	Username string `validate:"required" yaml:"username"`
	Password string `validate:"required" yaml:"password"`
}

type RestConfig struct {
	Port    uint   `validate:"required" yaml:"port"`
	BaseUrl string `validate:"required" yaml:"baseUrl"`
}

type DataConfig struct {
	Dir string `validate:"required" yaml:"dir"`
}

type UserConfig struct {
	Salt             string `validate:"required" yaml:"salt"`
	CookieHashKey    string `validate:"required" yaml:"cookieHashKey"`
	CookieEncryptKey string `validate:"required" yaml:"cookieEncryptKey"`
}

type AdminConfig struct {
	Username string `validate:"required" yaml:"username"`
	Password string `validate:"required" yaml:"password"`
}

type MailConfig struct {
	Server               string `validate:"required" yaml:"server"`
	Port                 uint   `validate:"required" yaml:"port"`
	From                 string `validate:"required" yaml:"from"`
	Username             string `validate:"required" yaml:"username"`
	Password             string `validate:"required" yaml:"password"`
	Subject              string `validate:"required" yaml:"subject"`
	RegisterTemplatePath string `validate:"required" yaml:"registerTemplatePath"`
}

type Config struct {
	Db    DbConfig    `validate:"dive,required" yaml:"db"`
	Rest  RestConfig  `validate:"dive,required" yaml:"rest"`
	Data  DataConfig  `validate:"dive,required" yaml:"data"`
	User  UserConfig  `validate:"dive,required" yaml:"user"`
	Admin AdminConfig `validate:"dive,required" yaml:"admin"`
	Mail  MailConfig  `validate:"dive,required" yaml:"mail"`
}

func LoadConfig() (*Config, error) {
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

var Export = fx.Options(fx.Provide(LoadConfig))
