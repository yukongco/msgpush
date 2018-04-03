package dispatch

import (
	"context"
	"fmt"
	"sync"
	"time"

	cometpb "github.com/yukongco/msgpush/comet/proto"
	"github.com/yukongco/msgpush/common/check"
	"github.com/yukongco/msgpush/common/grpcclient"
	"github.com/yukongco/msgpush/dispatch/conf"
	. "github.com/yukongco/msgpush/dispatch/logs"
	pb "github.com/yukongco/msgpush/dispatch/proto"
	msgpb "github.com/yukongco/msgpush/message/proto"
	"github.com/yukongco/grpool"
)

var (
	msgGrpcPool *grpcclient.Pool

	// storage comet grpc pool, communicate to websocket, key= comet addr, value= *grpcclient.Pool
	cometGrpcPoolMap *sync.Map
)

type DispatchSer struct {
}

func init() {
	cometGrpcPoolMap = &sync.Map{}
}

func InitMsgGrpc() error {
	msgGrpcPool = grpcclient.NewPoolTimeout(conf.Conf.BaseConf.MsgGrpcAddr, 100, 500, 36000*time.Second)
	if msgGrpcPool == nil {
		Log.Errorf("message grpc init pool nil")
		return fmt.Errorf("message grpc init pool nil")
	}

	return nil
}

// msg push
func (c *DispatchSer) DispatchMsg(ctx context.Context, in *pb.DspMsgReq) (*pb.DspMsgResp, error) {
	if in == nil {
		Log.Error("msg get cli ip port req is nil")
		return nil, fmt.Errorf("msg get cli ip port is nil")
	}

	addrs, err := GetSerAddr(check.Comet_Addr_Type)
	if err != nil {
		return nil, fmt.Errorf("get comet service ip err:%v", err)
	}

	if len(addrs) < 1 {
		tmpStr := fmt.Sprintf("service addr is nil")
		Log.Error(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}

	err = DspMsgPush(addrs, in)
	if err != nil {
		return nil, fmt.Errorf("push msg err: %v", err)
	}

	return &pb.DspMsgResp{}, nil
}

// get grpc service host and port, 1== all, 2==comet 3==message
func GetSerAddr(typeStr string) ([]string, error) {
	// get the connet ip and port from redis
	getCtx, getCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer getCancel()

	conn, err := msgGrpcPool.Get(getCtx)
	if err != nil {
		tmpStr := fmt.Sprintf("cann't get message conn from pool of grpc, err: %v", err)
		Log.Errorf(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}
	defer conn.Close()

	requestCtx, requestCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer requestCancel()

	req := &msgpb.SerType{}
	// 1== all, 2==comet 3==message addr, get comet addr
	req.Type = typeStr
	msgCli := msgpb.NewMsgSaveServiceClient(conn.Client)
	addrResp, err := msgCli.GetSerAddr(requestCtx, req)
	if err != nil {
		tmpStr := fmt.Sprintf("Query conet addr by key=%v err: %v", req.Type, err)
		Log.Error(tmpStr)
		fmt.Println(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}

	return addrResp.Addrs, nil
}

// msg push to comet
func DspMsgPush(cometAddrs []string, in *pb.DspMsgReq) error {
	totalNum := len(cometAddrs)

	// 协程池, 任务量, 总量, 超时时间
	pool := grpool.NewPool(totalNum, totalNum, 30*time.Second)
	defer pool.Release()
	pool.WaitCount(totalNum)

	for index := 0; index < totalNum; index++ {
		i := index
		pool.JobQueue <- grpool.Job{
			Jobid: i,
			Jobfunc: func() (interface{}, error) {
				// say that job is done, so we can know how many jobs are finished
				var cli *grpcclient.Pool
				value, ok := cometGrpcPoolMap.Load(cometAddrs[i])
				if !ok {
					// renew register grpcpool
					cli = grpcclient.NewPoolTimeout(cometAddrs[i], 100, 500, 36000*time.Second)
					if msgGrpcPool == nil {
						Log.Errorf("message grpc init pool by ip port %v err", cometAddrs[i])
						return nil, fmt.Errorf("message grpc init pool by ip port %v err", cometAddrs[i])
					}
					go cometGrpcPoolMap.Store(cometAddrs[i], cli)
					if cometAddrs[i] == in.LocalAddr {
						return nil, check.Err_Already_Try
					}
				} else {
					cli, ok = value.(*grpcclient.Pool)
					if !ok {
						tmpStr := fmt.Sprintf("grpc clinet=%v is now type of *grpcclient.Pool", value)
						Log.Errorf(tmpStr)
						return nil, fmt.Errorf(tmpStr)
					}
				}

				err := DspSingleMsgPush(cli, in)
				if err != nil {
					tmpStr := fmt.Sprintf("dispatch comet msg push err: %v", err)
					fmt.Println(tmpStr)
					return nil, fmt.Errorf(tmpStr)
				}
				return nil, nil
			},
		}
	}
	pool.WaitAll()

	pushFlag := false
	for res := range pool.Jobresult {
		if res.Err == nil {
			pushFlag = true
			break
		}
	}

	if pushFlag != true {
		Log.Error("dispatch msg push not success, comet addr: ", cometAddrs)
		return fmt.Errorf("dispatch msg push not success")
	}

	return nil
}

// push sinagle msg push
func DspSingleMsgPush(grpcCli *grpcclient.Pool, in *pb.DspMsgReq) error {
	// get the connet ip and port from redis
	getCtx, getCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer getCancel()

	conn, err := grpcCli.Get(getCtx)
	if err != nil {
		tmpStr := fmt.Sprintf("cann't get comet conn from pool of grpc, err: %v", err)
		Log.Errorf(tmpStr)
		return fmt.Errorf(tmpStr)
	}
	defer conn.Close()

	requestCtx, requestCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer requestCancel()

	req := &cometpb.PushReq{}
	req.Callback = in.Callback
	req.Cert = in.Cert
	req.Code = in.Code
	req.Expire = in.Expire
	req.Key = in.Key
	req.Msg = in.Msg
	req.Phone = in.Phone
	req.Platform = in.Platform
	req.PushId = in.PushId
	req.PushUtc = in.PushUtc
	req.Topic = in.Topic

	cometCli := cometpb.NewMsgPushServiceClient(conn.Client)
	_, err = cometCli.WRRMsgPush(requestCtx, req)
	if err != nil {
		tmpStr := fmt.Sprintf("Comet msg push err: %v", err)
		fmt.Println(tmpStr)
		return fmt.Errorf(tmpStr)
	}

	return nil
}
