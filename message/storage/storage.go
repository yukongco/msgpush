package msgsave

import (
	"fmt"

	"github.com/yukongco/msgpush/common/check"
	"github.com/yukongco/msgpush/message/conf"
	pb "github.com/yukongco/msgpush/message/proto"
	"github.com/yukongco/msgpush/message/redis"
)

var (
	UseStorage Storage
)

type Storage interface {
	// SavePrivate Save single private msg.
	SavePrivate(in *pb.MsgSaveReq) error

	// get msg by key
	GetMsgByKey(in *pb.MsgFilter) (*pb.FilterResp, error)

	// del msg by key and member
	DelSetByMem(in *pb.DelSetReq) (*pb.DelSetResp, error)

	// register connet client host and port, key=devId + code
	RegisterCli(in *pb.RegCliReq) (*pb.RegCliResp, error)

	// register connet client host and port, key=devId + code
	DelCli(in *pb.DelCliReq) (*pb.DelCliResp, error)

	// get service addrs
	GetSerAddr(in *pb.SerType) (*pb.AddrResp, error)

	Close() error
}

// InitStorage init the storage type(mysql or redis).
func InitStorage() error {
	var err error

	if conf.Conf.BaseConf.StorageType == check.Redis {
		UseStorage, err = libredis.InitRedis()
		if err != nil {
			return fmt.Errorf("init redis err: %v", err)
		}
	} else {
		return fmt.Errorf("unknown storage type: %v", conf.Conf.BaseConf.StorageType)

	}

	return nil
}
