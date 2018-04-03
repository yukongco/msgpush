package libredis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/yukongco/msgpush/common/check"
	. "github.com/yukongco/msgpush/message/logs"
	pb "github.com/yukongco/msgpush/message/proto"
)

// save single of msg
func (c *RedisPool) SavePrivate(in *pb.MsgSaveReq) error {
	key := in.DevId + in.Code

	inBytes, err := json.Marshal(in)
	if err != nil {
		Log.Errorf("json fmt bytes err: %v", err)
		return fmt.Errorf("json fmt bytes err: %v", err)
	}

	err = AddSet(key, inBytes)
	if err != nil {
		Log.Errorf("save redis msg key=%v err: %v", key, err)
		return fmt.Errorf("save redis msg key=%v err: %v", key, err)
	}

	return nil
}

// delete set by key and member, 以及判断该消息是否过期
func (c *RedisPool) GetMsgByKey(in *pb.MsgFilter) (*pb.FilterResp, error) {
	smeValues, err := GetSets(in.Key)
	if err != nil {
		Log.Errorf("get redis msg key=%v err: %v", err)
		return nil, fmt.Errorf("del redis msg key=%v err: %v", err)
	}

	resp := &pb.FilterResp{}
	arr := []*pb.MsgSaveReq{}
	nowTimeUnix := time.Now().Unix()

	for _, v := range smeValues {
		tmp := pb.MsgSaveReq{}
		vBytes, ok := v.([]byte)
		if !ok {
			Log.Errorf("key =%v  value=%v is not byte", in.Key, v)
			continue
		}

		fmt.Println("vBytes: ", string(vBytes))

		err = json.Unmarshal(vBytes, &tmp)
		if err != nil {
			fmt.Println("err========", err)
			Log.Errorf("json unmarshal err: %v", err)
			continue
		}

		if tmp.Expire == "" || tmp.Expire == "0" {
			go DelSetMember(in.Key, vBytes)
			continue
		}

		expire, err := strconv.ParseInt(tmp.Expire, 10, 64)
		if err != nil {
			Log.Errorf("json unmarshal err: %v", err)
			continue
		}

		// 消息已经过期
		if expire < nowTimeUnix {
			go DelSetMember(in.Key, vBytes)
			continue
		}

		arr = append(arr, &tmp)
	}
	resp.Content = arr

	return resp, nil
}

// delete set by key and member
func (c *RedisPool) DelSetByMem(in *pb.DelSetReq) (*pb.DelSetResp, error) {
	err := DelSetMember(in.Key, in.Member)
	if err != nil {
		tmpStr := fmt.Sprintf("del redis msg key=%v meber=%v err: %v", in.Key, in.Member, err)
		Log.Errorf(tmpStr)
		return nil, fmt.Errorf(tmpStr)
	}

	return &pb.DelSetResp{}, nil
}

// close pool of redis
func (c *RedisPool) Close() error {
	redisPool.pool.Close()
	return nil
}

// get conect host and port, key=devId+code
func (c *RedisPool) DelCli(in *pb.DelCliReq) (*pb.DelCliResp, error) {
	err := DelSetMember(in.Key, in.ConectAddr)
	if err != nil {
		return nil, err
	}

	return &pb.DelCliResp{}, nil
}

// register host and addr to db
func (c *RedisPool) RegisterCli(in *pb.RegCliReq) (*pb.RegCliResp, error) {
	err := AddSet(in.Key, in.ConectAddr)
	if err != nil {
		return nil, err
	}

	return &pb.RegCliResp{}, nil
}

// get service addr
func (c *RedisPool) GetSerAddr(in *pb.SerType) (*pb.AddrResp, error) {
	var err error
	addrs := []string{}

	switch in.Type {
	case "1":
		addrs, err = GetAllSerAddr()
	case "2":
		addrs, err = GetCometAllAddr()
	case "3":
		addrs, err = GetMsgAllAddr()
	default:
		return nil, fmt.Errorf("type=%v is not exist", in.Type)
	}

	if err != nil {
		return nil, err
	}
	resp := &pb.AddrResp{}
	resp.Addrs = addrs

	return resp, nil
}

func GetAllSerAddr() ([]string, error) {
	addrs := []string{}
	cometFace, err := GetSets(check.Comet_Storage)
	if err != nil {
		return nil, fmt.Errorf("get comet addr err: %v", err)
	}

	msgFace, err := GetSets(check.Message_Storage)
	if err != nil {
		return nil, fmt.Errorf("get comet addr err: %v", err)
	}

	for _, v := range cometFace {
		value, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("comet addr=%v is not byte", v)
		}
		addrs = append(addrs, string(value))
	}

	for _, v := range msgFace {
		value, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("message addr=%v is not byte", v)
		}
		addrs = append(addrs, string(value))
	}

	return addrs, nil
}

func GetCometAllAddr() ([]string, error) {
	cometFace, err := GetSets(check.Comet_Storage)
	if err != nil {
		return nil, fmt.Errorf("get comet addr err: %v", err)
	}

	addrs := []string{}
	for _, v := range cometFace {
		value, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("comet addr=%v is not byte", v)
		}
		addrs = append(addrs, string(value))
	}

	return addrs, nil
}

func GetMsgAllAddr() ([]string, error) {
	msgFace, err := GetSets(check.Message_Storage)
	if err != nil {
		return nil, fmt.Errorf("get comet addr err: %v", err)
	}

	addrs := []string{}
	for _, v := range msgFace {
		value, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("message addr=%v is not byte", v)
		}
		addrs = append(addrs, string(value))
	}

	return addrs, nil
}
