package dispatch

import (
	"context"
	"fmt"
	"time"

	cometpb "github.com/yukongco/msgpush/comet/proto"
	"github.com/yukongco/msgpush/common/check"
	"github.com/yukongco/msgpush/common/grpcclient"
	. "github.com/yukongco/msgpush/dispatch/logs"
	msgpb "github.com/yukongco/msgpush/message/proto"
)

func HeartRun() {
	conn, err := msgGrpcPool.Get(context.Background())
	if err != nil {
		tmpStr := fmt.Sprintf("cann't get message conn from pool of grpc, err: %v", err)
		Log.Errorf(tmpStr)
		return
	}
	defer conn.Close()

	msgCli := msgpb.NewMsgSaveServiceClient(conn.Client)
	addrReq := &msgpb.SerType{}
	addrReq.Type = check.Comet_Addr_Type

	timer := time.NewTicker(3 * time.Second)
	defer timer.Stop()

	req := &msgpb.SerType{}
	req.Type = check.Comet_Addr_Type

	for {
		select {
		case <-timer.C:
			cometAddrs, err := msgCli.GetSerAddr(context.Background(), req)
			if err != nil {
				fmt.Println("get comet addr err: ", err)
				Log.Errorf("get comet addr err: %v", err)
				break
			}
			for _, v := range cometAddrs.Addrs {
				addr := v
				go PingServices(addr)
			}
		}
	}
}

// ping comet addrs
func PingServices(addr string) error {
	var cli *grpcclient.Pool
	value, ok := cometGrpcPoolMap.Load(addr)
	if !ok {
		// renew register grpcpool
		cli = grpcclient.NewPoolTimeout(addr, 100, 500, 36000*time.Second)
		if msgGrpcPool == nil {
			Log.Errorf("message grpc init pool by ip port %v err", addr)
			return fmt.Errorf("message grpc init pool by ip port %v err", addr)
		}
		go cometGrpcPoolMap.Store(addr, cli)
	} else {
		cli, ok = value.(*grpcclient.Pool)
		if !ok {
			tmpStr := fmt.Sprintf("grpc clinet=%v is now type of *grpcclient.Pool", value)
			Log.Errorf(tmpStr)
			return fmt.Errorf(tmpStr)
		}
		go cometGrpcPoolMap.Store(addr, cli)
	}

	err := PingComet(cli, &cometpb.CometPingReq{})
	if err != nil {
		tmpStr := fmt.Sprintf("ping comet addr=%v err: %v", addr, err)
		fmt.Println(tmpStr)
		Log.Error(tmpStr)
		delReq := &msgpb.DelCliReq{
			Key:        check.Comet_Storage,
			ConectAddr: addr,
		}
		go DelCli(delReq)
		go cometGrpcPoolMap.Delete(addr)
		return fmt.Errorf(tmpStr)
	}

	return nil
}

// comet ping
func PingComet(grpcCli *grpcclient.Pool, req *cometpb.CometPingReq) error {
	getCtx, getCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer getCancel()

	conn, err := grpcCli.Get(getCtx)
	if err != nil {
		tmpStr := fmt.Sprintf("cann't get comet conn from pool of grpc, err: %v", err)
		return fmt.Errorf(tmpStr)
	}
	defer conn.Close()

	requestCtx, requestCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer requestCancel()

	cometCli := cometpb.NewMsgPushServiceClient(conn.Client)
	_, err = cometCli.Ping(requestCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// delete comet addr from message
func DelCli(req *msgpb.DelCliReq) error {
	getCtx, getCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer getCancel()

	conn, err := msgGrpcPool.Get(getCtx)
	if err != nil {
		tmpStr := fmt.Sprintf("cann't get message conn from pool of grpc, err: %v", err)
		Log.Errorf(tmpStr)
		return fmt.Errorf(tmpStr)
	}
	defer conn.Close()

	requestCtx, requestCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer requestCancel()

	msgCli := msgpb.NewMsgSaveServiceClient(conn.Client)
	_, err = msgCli.DelCli(requestCtx, req)
	if err != nil {
		tmpStr := fmt.Sprintf("delet by key=%v member=%v err: %v", req.Key, req.ConectAddr, err)
		Log.Error(tmpStr)
		fmt.Println(tmpStr)
		return fmt.Errorf(tmpStr)
	}

	return nil
}
