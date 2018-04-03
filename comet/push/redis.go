package push

import (
	"context"
	"fmt"
	"time"

	. "github.com/yukongco/msgpush/comet/logs"
	msgpb "github.com/yukongco/msgpush/message/proto"
)

// notify grpc service message, client is on line
func (c *Client) NotifyMsg() error {
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

	req := &msgpb.MsgFilter{}
	req.Key = c.User.Key + c.User.Code

	msgCli := msgpb.NewMsgSaveServiceClient(conn.Client)
	resp, err := msgCli.GetMsgByKey(requestCtx, req)
	if err != nil {
		tmpStr := fmt.Sprintf("Query msg by key=%v err: %v", req.Key)
		Log.Error(tmpStr)
		fmt.Println(tmpStr)
		return fmt.Errorf(tmpStr)
	}
	// 判断用户是否还在线
	_, ok := HubOrg.Clients.Load(c)
	if !ok {
		tmpStr := fmt.Sprintf("the client devId=%v, code=%v is not online", c.User.Key, c.User.Code)
		Log.Errorf(tmpStr)
		return fmt.Errorf(tmpStr)
	}

	// push the msg
	for _, v := range resp.Content {
		c.send <- []byte(v.Msg)
	}

	return nil
}
