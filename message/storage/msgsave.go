package msgsave

import (
	"context"
	"fmt"

	. "github.com/yukongco/msgpush/message/logs"
	pb "github.com/yukongco/msgpush/message/proto"
)

type MsgSaveService struct {
}

// Grpc service msg push
func (c *MsgSaveService) MsgSave(ctx context.Context, in *pb.MsgSaveReq) (*pb.MsgSaveResp, error) {
	if in == nil {
		Log.Error("msg save req is nil")
		return nil, fmt.Errorf("msg save req is nil")
	}

	fmt.Println("key: ", in.DevId+in.Code)

	err := UseStorage.SavePrivate(in)
	if err != nil {
		tmpStr := fmt.Sprintf("save private err: %v", err)
		Log.Errorf(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}

	return &pb.MsgSaveResp{}, nil
}

// Grpc service msg get by key, key=devId + code
func (c *MsgSaveService) GetMsgByKey(ctx context.Context, in *pb.MsgFilter) (*pb.FilterResp, error) {
	if in == nil {
		Log.Error("msg get req is nil")
		return nil, fmt.Errorf("msg get req is nil")
	}

	resp, err := UseStorage.GetMsgByKey(in)
	if err != nil {
		tmpStr := fmt.Sprintf("get msg by key=%v err: %v", in.Key, err)
		Log.Errorf(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}

	return resp, nil
}

// Grpc service msg del by key and member
func (c *MsgSaveService) DelSetByMem(ctx context.Context, in *pb.DelSetReq) (*pb.DelSetResp, error) {
	if in == nil {
		Log.Error("msg del req is nil")
		return nil, fmt.Errorf("msg del req is nil")
	}

	resp, err := UseStorage.DelSetByMem(in)
	if err != nil {
		tmpStr := fmt.Sprintf("del msg by key=%v member=%v err: %v", in.Key, in.Member, err)
		Log.Errorf(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}

	return resp, nil
}

// Grpc service msg register conect register
func (c *MsgSaveService) RegisterCli(ctx context.Context, in *pb.RegCliReq) (*pb.RegCliResp, error) {
	if in == nil {
		Log.Error("msg register cli req is nil")
		return nil, fmt.Errorf("msg register cli req is nil")
	}

	resp, err := UseStorage.RegisterCli(in)
	if err != nil {
		tmpStr := fmt.Sprintf("register cli msg by req err: %v", in, err)
		Log.Errorf(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}

	return resp, nil
}

// Grpc service msg heart
func (c *MsgSaveService) Ping(ctx context.Context, in *pb.MsgPingReq) (*pb.MsgPingResp, error) {
	return &pb.MsgPingResp{}, nil
}

// Grpc service msg del by key, key=devId + code
func (c *MsgSaveService) DelCli(ctx context.Context, in *pb.DelCliReq) (*pb.DelCliResp, error) {
	if in == nil {
		Log.Error("msg del cli req is nil")
		return nil, fmt.Errorf("msg del cli req is nil")
	}

	resp, err := UseStorage.DelCli(in)
	if err != nil {
		tmpStr := fmt.Sprintf("del cli msg by key=%v err: %v", in.Key, err)
		Log.Errorf(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}

	return resp, nil
}

// Grpc service msg get serice addr ip+port
func (c *MsgSaveService) GetSerAddr(ctx context.Context, in *pb.SerType) (*pb.AddrResp, error) {
	if in == nil {
		Log.Error("msg get service is nil")
		return nil, fmt.Errorf("msg get service is nil")
	}

	resp, err := UseStorage.GetSerAddr(in)
	if err != nil {
		tmpStr := fmt.Sprintf("msg get service type=%v err: %v", in.Type, err)
		Log.Errorf(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}

	return resp, nil
}
