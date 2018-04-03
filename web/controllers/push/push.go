package push

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	pb "github.com/yukongco/msgpush/comet/proto"
	"github.com/yukongco/msgpush/common/check"
	"github.com/yukongco/msgpush/common/grpcclient"
	"github.com/yukongco/msgpush/web/conf"
	"github.com/yukongco/msgpush/web/controllers/base"
	. "github.com/yukongco/msgpush/web/logs"
)

var (
	grpcPool *grpcclient.Pool
)

func InitPushGrpc() error {
	grpcPool = grpcclient.NewPoolTimeout(conf.Conf.Comet.GrpcAddr, 100, 500, 36000*time.Second)
	if grpcPool == nil {
		Log.Errorf("push grpc init pool nil")
		return fmt.Errorf("push grpc init pool nil")
	}

	return nil
}

func PushPrivate(c *gin.Context) {
	getCtx, getCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer getCancel()

	conn, err := grpcPool.Get(getCtx)
	if err != nil {
		Log.Errorf("cann't get conn from pool of grpc, err: %v", err)
		base.WebResp(c, http.StatusBadRequest, nil, err.Error())
		return
	}
	defer conn.Close()

	requestCtx, requestCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer requestCancel()

	pushReq := &pb.PushReq{}
	pushReq.Key = c.Query("key")
	pushReq.Expire = c.Query("expire")
	pushReq.PushUtc = c.Query("push_utc")
	pushReq.Platform = c.Query("platform")
	pushReq.Cert = c.Query("cert")
	pushReq.Phone = c.Query("phone")
	pushReq.Code = c.Query("code")

	msgBytes, err := c.GetRawData()
	if err != nil {
		tmpStr := fmt.Sprintf("fmt msg cotent is: %v", err)
		Log.Errorf(tmpStr)
		base.WebResp(c, http.StatusBadRequest, nil, tmpStr)
		return
	}
	if string(msgBytes) == "" {
		Log.Error("msg content is nil")
		base.WebResp(c, http.StatusBadRequest, nil, "msg content is nil")
		return
	}
	pushReq.Msg = string(msgBytes)

	err = pushParmCheck(pushReq)
	if err != nil {
		tmpStr := fmt.Sprintf("push parm err: %v", err)
		Log.Error(tmpStr)
		base.WebResp(c, http.StatusBadRequest, nil, tmpStr)
		return
	}

	// flag id
	pushReq.PushId = check.GetStandId()

	grpcCli := pb.NewMsgPushServiceClient(conn.Client)
	resp, err := grpcCli.MsgPush(requestCtx, pushReq)
	if err != nil {
		tmpStr := fmt.Sprintf("Single msg push err: %v", err)
		Log.Error(tmpStr)
		fmt.Println(tmpStr)
		base.WebResp(c, http.StatusInternalServerError, nil, tmpStr)
		return
	}

	fmt.Println("single msg push resp: ", resp)

	base.WebResp(c, http.StatusOK, nil, "success")
	return
}

// push parm check
func pushParmCheck(pushReq *pb.PushReq) error {
	err := check.NumLetter(pushReq.Key)
	if err != nil {
		return fmt.Errorf("key=%v invalid: %v", pushReq.Key, err)
	}

	// expire time, unit s
	if pushReq.Expire != "" {
		err := check.NumCheck(pushReq.Expire)
		if err != nil {
			return fmt.Errorf("expire=%v invalid: %v", pushReq.Expire, err)
		}
	}

	// cron time, utc unit s
	if pushReq.PushUtc != "" {
		err := check.NumCheck(pushReq.PushUtc)
		if err != nil {
			return fmt.Errorf("push_utc=%v invalid: %v", pushReq.PushUtc, err)
		}
	}

	// platform just android and ios and pc
	if pushReq.Platform != "" {
		if pushReq.Platform != check.Android &&
			pushReq.Platform != check.IOS &&
			pushReq.Platform != check.PC {
			return fmt.Errorf("platform=%v invalid", pushReq.Platform)
		}

		if pushReq.Platform == check.IOS {
			err := check.NumLetterPointLine(pushReq.Cert)
			if err != nil {
				return fmt.Errorf("cert=%v invalid: %v", pushReq.Cert, err)
			}
		}
	}

	if pushReq.Phone != "" {
		err := check.NumCheck(pushReq.Phone)
		if err != nil {
			return fmt.Errorf("phone=%v invalid: %v", pushReq.Phone, err)
		}
	}

	return nil
}
