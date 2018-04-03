package check

import (
	"errors"
)

var (
	// task of cron
	Cron_Push_Err = errors.New("cron push msg")

	User_Not_On_Line = errors.New("user not on line")

	Newline = []byte{'\n'}

	Err_Already_Try = errors.New("Have already try")
)

const (
	Android = "android"
	IOS     = "ios"
	PC      = "pc"

	IOS_Cert_Prex = "./ioscert/"
	Dev           = "dev"

	// Redis
	Redis    = "redis"
	SADD     = "SADD"
	SMEMBERS = "SMEMBERS"
	SREM     = "SREM"

	Comet_Storage   = "cometdbstorage"
	Message_Storage = "cometdbstorage"
	Comet_Addr_Type = "2"
	Msg_Addr_Type   = "3"
)
