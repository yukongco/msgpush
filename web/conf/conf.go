package conf

import (
	ini "gopkg.in/ini.v1"
)

var Conf *Config

type Config struct {
	Base    Base
	Comet   Comet
	LogConf LogConf
}

type Base struct {
	HttpBind string `ini:"HttpBind"`
	Env      string `ini:"Env"`
}

type Comet struct {
	GrpcAddr string `ini:"GrpcAddr"`
}

// Log config
type LogConf struct {
	LogPath  string `ini:"LogPath"`
	LogLevel string `ini:"LogLevel"`
}

func InitConfig(confPath *string) (*Config, error) {
	Conf = new(Config)
	if err := ini.MapTo(Conf, *confPath); err != nil {
		return nil, err
	}

	return Conf, nil
}
