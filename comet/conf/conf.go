package conf

import (
	ini "gopkg.in/ini.v1"
)

var Conf *Config

type Config struct {
	GrpcConf
	IosConf
	WebsocketConf
	BaseConf
	LogConf
}

// protobuf grpc config
type GrpcConf struct {
	GrpcHost                  string `ini:"GrpcHost"`
	GrpcPort                  string `ini:"GrpcPort"`
	GrpcMaxConnectIdleSec     int    `ini:"GrpcMaxConnectIdleSec"`
	GrpcMaxConnectAgeSec      int    `ini:"GrpcMaxConnectAgeSec"`
	GrpcMaxConnectAgeGraceSec int    `ini:"GrpcMaxConnectAgeGraceSec"`
	GrpcTimeSec               int    `ini:"GrpcTimeSec"`
	GrpcTimeTimeoutSec        int    `ini:"GrpcTimeTimeoutSec"`
}

type IosConf struct {
	Mode     string `ini:"Mode"`
	CertFile string `ini:"CertFile"` // default cert file
	Pwd      string `ini:"Pwd"`
}

type WebsocketConf struct {
	HostPort      string `ini:"HostPort"`      // websocket host and port
	BroadcastMax  int    `ini:"BroadcastMax"`  // max num of broadcast
	RegisterMax   int    `ini:"RegisterMax"`   // max num of register client
	UnregisterMax int    `ini:"UnregisterMax"` // max num of unregister client
	ClientSendMax int    `ini:"ClientSendMax"` // max num of client send msg at the same time
}

type BaseConf struct {
	MsgGrpcAddr string `ini:"MsgGrpcAddr"` // message grpc addr
	ExternAddr  string `ini:"ExternAddr"`  // extern ip register to redis, is host+GrpcPort
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
