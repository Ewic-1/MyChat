package config

import (
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

type MainConfig struct {
	AppName string `toml:"appName"`
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
}

type MysqlConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
	DatabaseName string `toml:"databaseName"`
}

type RedisConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	Db       int    `toml:"db"`
}

type AuthCodeConfig struct {
	AccessKeyID     string `toml:"accessKeyID"`
	AccessKeySecret string `toml:"accessKeySecret"`
	SignName        string `toml:"signName"`
	TemplateCode    string `toml:"templateCode"`
}

type LogConfig struct {
	LogPath string `toml:"logPath"`
}

type KafkaConfig struct {
	MessageMode string        `toml:"messageMode"`
	HostPort    string        `toml:"hostPort"`
	LoginTopic  string        `toml:"loginTopic"`
	LogoutTopic string        `toml:"logoutTopic"`
	ChatTopic   string        `toml:"chatTopic"`
	Partition   int           `toml:"partition"`
	Timeout     time.Duration `toml:"timeout"`
}

type JWTConfig struct {
	// ActiveKid 标识当前签发 token 使用的密钥版本。
	ActiveKid string `toml:"activeKid"`
	// AccessSecret / RefreshSecret 是当前生效密钥，优先级低于环境变量。
	AccessSecret  string `toml:"accessSecret"`
	RefreshSecret string `toml:"refreshSecret"`
	// Previous*Secrets 用于密钥轮换时兼容旧 token 验签。
	PreviousAccessSecrets  []string `toml:"previousAccessSecrets"`
	PreviousRefreshSecrets []string `toml:"previousRefreshSecrets"`
	// AccessTTLSeconds / RefreshTTLSeconds 控制两类 token 过期时间（秒）。
	AccessTTLSeconds  int64 `toml:"accessTTLSeconds"`
	RefreshTTLSeconds int64 `toml:"refreshTTLSeconds"`
}

type StaticSrcConfig struct {
	StaticAvatarPath string `toml:"staticAvatarPath"`
	StaticFilePath   string `toml:"staticFilePath"`
}

type Config struct {
	MainConfig      `toml:"mainConfig"`
	MysqlConfig     `toml:"mysqlConfig"`
	RedisConfig     `toml:"redisConfig"`
	AuthCodeConfig  `toml:"authCodeConfig"`
	LogConfig       `toml:"logConfig"`
	KafkaConfig     `toml:"kafkaConfig"`
	JWTConfig       `toml:"jwtConfig"`
	StaticSrcConfig `toml:"staticSrcConfig"`
}

var config *Config

func LoadConfig() error {
	if _, err := toml.DecodeFile("configs/config.toml", config); err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}

func GetConfig() *Config {
	if config == nil {
		config = new(Config)
		_ = LoadConfig()
	}
	return config
}
