package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConf   `mapstructure:"server"`
	Database DatabaseConf `mapstructure:"database"`
	Log      LogConf      `mapstructure:"log"`
}

type ServerConf struct {
	Port    string `mapstructure:"port"`
	Timeout string `mapstructure:"timeout"`
}

type DatabaseConf struct {
	Sqlite SqliteNode `mapstructure:"sqlite"`
	Mysql  MysqlNode  `mapstructure:"mysql"`
	Redis  RedisNode  `mapstructure:"redis"`
	Mongo  MongoNode  `mapstructure:"mongo"`
	Rabbit RabbitNode `mapstructure:"rabbit"`
}

type LogConf struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

type MongoNode struct {
	First MongoConf `mapstructure:"first"`
}

type RabbitNode struct {
	First RabbitConf `mapstructure:"first"`
}

type RabbitConf struct {
	Uri string `mapstructure:"uri"`
}

type MongoConf struct {
	Uri string `mapstructure:"uri"`
}

type RedisConf struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

type MysqlConf struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	Database    string `mapstructure:"database"`
	MaxOpenConn int    `mapstructure:"max_open_conn"`
	MaxIdleConn int    `mapstructure:"max_idle_conn"`
}

type SqliteConf struct {
	Path string `mapstructure:"path"`
}

type RedisNode struct {
	First RedisConf `mapstructure:"first"`
}

type SqliteNode struct {
	First SqliteConf `mapstructure:"first"`
}

type MysqlNode struct {
	First  MysqlConf `mapstructure:"first"`
	Second MysqlConf `mapstructure:"second"`
	Third  MysqlConf `mapstructure:"third"`
	Fourth MysqlConf `mapstructure:"fourth"`
}

var AppConfig Config

func LoadConfig(path string) error {
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
