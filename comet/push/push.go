package push

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/yukongco/msgpush/comet/conf"
	. "github.com/yukongco/msgpush/comet/logs"
	pb "github.com/yukongco/msgpush/comet/proto"
	"github.com/yukongco/msgpush/common/check"
	"github.com/yukongco/msgpush/common/grpcclient"
	msgpb "github.com/yukongco/msgpush/message/proto"
)

type MsgPushService struct {
}

var (
	msgGrpcPool *grpcclient.Pool
)

func InitMsgGrpc() error {
	msgGrpcPool = grpcclient.NewPoolTimeout(conf.Conf.BaseConf.MsgGrpcAddr, 100, 500, 36000*time.Second)
	if msgGrpcPool == nil {
		Log.Errorf("message grpc init pool nil")
		return fmt.Errorf("message grpc init pool nil")
	}

	return nil
}

// Grpc service msg push
func (c *MsgPushService) MsgPush(ctx context.Context, in *pb.PushReq) (*pb.PushResp, error) {
	fmt.Println("msg push is start")

	if in == nil {
		Log.Error("msg push req is nil")
		return &pb.PushResp{}, fmt.Errorf("msg push req is nil")
	}

	var pushUinx int64
	var err error
	if in.PushUtc != "" {
		pushUinx, err = strconv.ParseInt(in.PushUtc, 10, 64)
		if err != nil {
			tmpStr := fmt.Sprintf("push utc fmt to unix err: %v", err)
			Log.Error(tmpStr)
			return &pb.PushResp{}, fmt.Errorf(tmpStr)
		}
	}

	// judge whether need to delay send
	if pushUinx > time.Now().Unix() {
		err = DelaySend(in)
		// task of cron
		if err != nil {
			tmpStr := fmt.Sprintf("delay send platform=%v, key=%v pushUtc=%v err: %v", in.Platform, in.Key, in.PushUtc, err)
			Log.Errorf(tmpStr)
			return &pb.PushResp{}, fmt.Errorf(tmpStr)
		}

		return &pb.PushResp{}, nil
	}

	resp, err := SendMsg(in)
	if err != nil {
		tmpStr := fmt.Sprintf("Push msg err: %v", err)
		Log.Error(tmpStr)
		return &pb.PushResp{}, nil
	}
	fmt.Println("resp: ", resp)

	return &pb.PushResp{}, nil
}

// Push msg to device
func SendMsg(in *pb.PushReq) (*pb.PushResp, error) {
	var err error

	switch in.Platform {
	case check.IOS:
		err = PushToIOS(in.Key, []byte(in.Msg), in.Cert, in.Topic)
		if err != nil {
			tmpStr := fmt.Sprintf("push msg to ios err=%v", err)
			return nil, fmt.Errorf(tmpStr)
		}
		fmt.Println("push ios success")
	case check.Android, check.PC:
		err = AndroidPcPush(in.Key, []byte(in.Msg))
		if err != nil {
			go MsgSave(in)
			tmpStr := fmt.Sprintf("push msg to %v err=%v", in.Platform, err)
			return nil, fmt.Errorf(tmpStr)
		}
	default:
	}

	return nil, nil
}

// delay send msg
func DelaySend(in *pb.PushReq) error {
	cronExpress, err := GetCronExpress(in.PushUtc)
	if err != nil {
		return fmt.Errorf("get cron express err: %v", err)
	}

	err = CronRes.AddFuncWithName(cronExpress, func() {
		go SendMsg(in)
	}, in.PushId)
	if err != nil {
		return fmt.Errorf("add cron job err: %v", err)
	}

	return nil
}

// message storage
func MsgSave(in *pb.PushReq) error {
	getCtx, getCancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	expireStr := ""
	if in.Expire != "" {
		expireTmp, err := strconv.ParseInt(in.Expire, 10, 64)
		if err != nil {
			tmpStr := fmt.Sprintf("expire=%v fmt to int err: %v", err)
			Log.Errorf(tmpStr)
			return fmt.Errorf(tmpStr)
		}
		expireTmp = time.Now().Add(time.Duration(expireTmp) * time.Second).Unix()
		expireStr = strconv.FormatInt(expireTmp, 10)
	}

	req := msgpb.MsgSaveReq{}
	req.Callback = in.Callback
	req.Code = in.Code
	req.DevId = in.Key
	req.Expire = expireStr
	req.Msg = in.Msg
	req.Phone = in.Phone
	req.Platform = in.Platform
	req.PushId = in.PushId

	msgCli := msgpb.NewMsgSaveServiceClient(conn.Client)
	_, err = msgCli.MsgSave(requestCtx, &req)
	if err != nil {
		tmpStr := fmt.Sprintf("Single msg save err: %v", err)
		Log.Error(tmpStr)
		fmt.Println(tmpStr)
		return fmt.Errorf(tmpStr)
	}

	return nil
}

// dispatch service push msg
func (c *MsgPushService) WRRMsgPush(ctx context.Context, in *pb.PushReq) (*pb.PushResp, error) {
	key := in.Key + in.Code

	client, exist := HubOrg.IsOnline(key)
	if !exist {
		return nil, check.User_Not_On_Line
	}

	client.send <- []byte(in.Msg)

	return &pb.PushResp{}, nil
}

// ping heart
func (c *MsgPushService) Ping(ctx context.Context, in *pb.CometPingReq) (*pb.CometPingResp, error) {

	return &pb.CometPingResp{}, nil
}

// register ip to redis
func RegisterCometIp() error {
	ip := ""
	var err error

	if conf.Conf.BaseConf.ExternAddr != "" {
		ip = conf.Conf.BaseConf.ExternAddr
	} else {
		ip, err = check.GetLoacalIp()
		if err != nil {
			Log.Errorf("get local ip is err: ", err)
			return fmt.Errorf("get local ip is err: ", err)
		}
	}

	conn, err := msgGrpcPool.Get(context.Background())
	if err != nil {
		tmpStr := fmt.Sprintf("cann't get message conn from pool of grpc, err: %v", err)
		Log.Errorf(tmpStr)
		return fmt.Errorf(tmpStr)
	}
	defer conn.Close()

	msgCli := msgpb.NewMsgSaveServiceClient(conn.Client)
	req := &msgpb.RegCliReq{}
	req.Key = check.Comet_Storage
	req.ConectAddr = ip + conf.Conf.GrpcConf.GrpcPort

	timer := time.NewTicker(5 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			_, err = msgCli.RegisterCli(context.Background(), req)
			if err != nil {
				fmt.Printf("register ip=%v to db err: %v\n", ip, err)
			} else {
				return nil
			}
		}
	}

	return nil
}
