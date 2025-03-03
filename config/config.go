package config

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
	"os"
)

type RateLimitConfig struct {
	Requests   int `validate:"required,gt=0" yaml:"requests"`
	IntervalMs int `validate:"required,gt=0" yaml:"intervalMs"`
}

type SearchConfig struct {
	Proxy          string          `yaml:"proxy"`
	RateLimit      RateLimitConfig `yaml:"rateLimit"`
	TimeoutMs      int             `yaml:"timeoutMs"`
	RssIntervalSec int             `yaml:"rssIntervalSec"`
}

type DbConfig struct {
	Host                string `validate:"required" yaml:"host"`
	Port                uint   `validate:"required" yaml:"port"`
	DbName              string `validate:"required" yaml:"dbName"`
	Username            string `validate:"required" yaml:"username"`
	Password            string `validate:"required" yaml:"password"`
	MaxIdleConns        int    `validate:"required" yaml:"maxIdleConns"`
	MaxOpenConns        int    `validate:"required" yaml:"maxOpenConns"`
	ConnMaxLifetimeSecs int    `validate:"required" yaml:"connMaxLifetimeSecs"`
}

type RestConfig struct {
	Port    uint   `validate:"required" yaml:"port"`
	BaseUrl string `validate:"required" yaml:"baseUrl"`
}

type WebSocketConfig struct {
	WriteTimeoutMs        int   `validate:"required,gt=0" yaml:"writeTimeoutMs"`
	PingTimeoutMs         int   `validate:"required,gt=0" yaml:"pingTimeoutMs"`
	PingIntervalMs        int   `validate:"required,gt=0" yaml:"pingIntervalMs"`
	MaxMessageSize        int64 `validate:"required,gt=0" yaml:"maxMessageSize"`
	BufferSize            int   `validate:"required,gt=0" yaml:"bufferSize"`
	MessageChanBufferSize int   `validate:"required,gt=0" yaml:"messageChanBufferSize"`
}

type DataConfig struct {
	Dir              string `validate:"required" yaml:"dir"`
	DownloadBpsLimit int    `validate:"gt=0" yaml:"downloadBpsLimit"`
	UploadBpsLimit   int    `validate:"gt=0" yaml:"uploadBpsLimit"`
	EpisodesPerPage  int    `validate:"gt=0" yaml:"episodesPerPage"`
}

type FFMpegConfig struct {
	StreamSizeArgs string `validate:"required" yaml:"streamSizeArgs"`
	ExtractSubArgs string `validate:"required" yaml:"extractSubArgs"`
	ConvertArgs    string `validate:"required" yaml:"convertArgs"`
	MaxThreads     int    `validate:"required" yaml:"maxThreads"`
}

type ThumbConfig struct {
	Args     string `validate:"required" yaml:"args"`
	Attempts int    `validate:"gt=0" yaml:"attempts"`
}

type UserConfig struct {
	Salt             string `validate:"required" yaml:"salt"`
	CookieHashKey    string `validate:"required" yaml:"cookieHashKey"`
	CookieEncryptKey string `validate:"required" yaml:"cookieEncryptKey"`
	AuthDurationSec  int    `validate:"gt=0" yaml:"authDurationSec"`
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
	Db        DbConfig        `validate:"dive,required" yaml:"db"`
	Rest      RestConfig      `validate:"dive,required" yaml:"rest"`
	WebSocket WebSocketConfig `validate:"dive,required" yaml:"ws"`
	Data      DataConfig      `validate:"dive,required" yaml:"data"`
	FFMpeg    FFMpegConfig    `validate:"dive,required" yaml:"ffmpeg"`
	Search    SearchConfig    `validate:"dive,required" yaml:"search"`
	Thumb     ThumbConfig     `validate:"dive,required" yaml:"thumb"`
	User      UserConfig      `validate:"dive,required" yaml:"user"`
	Admin     AdminConfig     `validate:"dive,required" yaml:"admin"`
	Mail      MailConfig      `validate:"dive,required" yaml:"mail"`
}

func GetDefaultConfig() Config {
	return Config{
		Db: DbConfig{
			Host:                "localhost",
			Port:                5432,
			DbName:              "anileha",
			Username:            "postgres",
			Password:            "postgres",
			MaxIdleConns:        10,
			MaxOpenConns:        30,
			ConnMaxLifetimeSecs: 1800,
		},
		Rest: RestConfig{
			Port:    8080,
			BaseUrl: "http://127.0.0.1:8080",
		},
		WebSocket: WebSocketConfig{
			WriteTimeoutMs:        10000,
			PingTimeoutMs:         30000,
			PingIntervalMs:        20000,
			MaxMessageSize:        1024,
			BufferSize:            1024,
			MessageChanBufferSize: 256,
		},
		Data: DataConfig{
			Dir:              "data",
			DownloadBpsLimit: 5 * 1024 * 1024,
			UploadBpsLimit:   1024 * 1024,
			EpisodesPerPage:  20,
		},
		FFMpeg: FFMpegConfig{
			StreamSizeArgs: "$BASE -analyzeduration $MAX -probesize $MAX -i $INPUT -map $MAP -c copy -f null -",
			ExtractSubArgs: "$BASE -i $INPUT -map $MAP -f srt $OUTPUT",
			ConvertArgs:    "$BASE -i $INPUT -acodec aac -b:a 196k -ac 2 -vcodec libx264 -crf 18 -tune animation -pix_fmt yuv420p -preset slow -f mp4 $FILTER_SUB $FILTER_AUDIO $MAP_SUB $MAP_AUDIO -movflags +faststart -threads $THREADS $OUTPUT",
			MaxThreads:     16,
		},
		Search: SearchConfig{
			RateLimit: RateLimitConfig{
				Requests:   1,
				IntervalMs: 5000,
			},
			TimeoutMs:      10000,
			RssIntervalSec: 1800,
		},
		Thumb: ThumbConfig{
			Args:     "$BASE -ss $SS -i $INPUT -frames:v 1 $OUTPUT",
			Attempts: 5,
		},
		User: UserConfig{
			Salt:             "salt",
			CookieHashKey:    "qwertyuiopasdfghjkl;'zxcvbnm,.qw",
			CookieEncryptKey: "qwertyuiopasdfgh",
			AuthDurationSec:  604800,
		},
		Admin: AdminConfig{
			Username: "admin",
			Password: "admin",
		},
		Mail: MailConfig{
			Server:               "your.smtp.server",
			Port:                 1337,
			From:                 "your@email.com",
			Username:             "user",
			Password:             "pass",
			Subject:              "AniLeha registration",
			RegisterTemplatePath: "register.tmpl",
		},
	}
}

func LoadConfig() (*Config, error) {
	config := GetDefaultConfig()

	configBytes, err := os.ReadFile("config.yaml")
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
