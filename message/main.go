package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yukongco/msgpush/message/conf"
	. "github.com/yukongco/msgpush/message/logs"
	pb "github.com/yukongco/msgpush/message/proto"
	"github.com/yukongco/msgpush/message/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	confInfo *conf.Config
	confPath = flag.String("config", "./conf/app.ini", "message profilePath")
)

func Init() error {
	flag.Parse()
	var err error
	confInfo, err = conf.InitConfig(confPath)
	if err != nil {
		fmt.Println("init config err: ", err)
		return fmt.Errorf("Init config is err: %v", err)
	}

	err = InitLog(confInfo.LogConf.LogPath, confInfo.LogConf.LogLevel)
	if err != nil {
		fmt.Println("init log is err: ", err)
		return fmt.Errorf("init log is err: %v", err)
	}

	return nil
}

func main() {
	//catch global panic
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Panic error: ", err)
			Log.Info("Panic error: %v", err)
		}
	}()

	err := Init()
	if err != nil {
		fmt.Println("main init err: ", err)
		return
	}
	// init storage
	err = msgsave.InitStorage()
	if err != nil {
		Log.Errorf("init storage err: %v", err)
		return
	}
	defer msgsave.UseStorage.Close()

	grpcLis, err := net.Listen("tcp", confInfo.GrpcConf.GrpcHost+":"+confInfo.GrpcConf.GrpcPort)
	if err != nil {
		Log.Errorf("net listen failed: %v", err)
		return
	}
	grpcServer := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(confInfo.GrpcConf.GrpcMaxConnectIdleSec) * time.Second,
		MaxConnectionAge:      time.Duration(confInfo.GrpcConf.GrpcMaxConnectAgeSec) * time.Second,
		MaxConnectionAgeGrace: time.Duration(confInfo.GrpcConf.GrpcMaxConnectAgeGraceSec) * time.Second,
		Time:    time.Duration(confInfo.GrpcConf.GrpcTimeSec) * time.Second,
		Timeout: time.Duration(confInfo.GrpcConf.GrpcTimeTimeoutSec) * time.Second,
	}))

	// register grpc service
	pb.RegisterMsgSaveServiceServer(grpcServer, &msgsave.MsgSaveService{})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = grpcServer.Serve(grpcLis); err != nil {
			Log.Errorf("grpc error:%v", err)
			panic(err)
		}
	}()

	fmt.Println("msg push message start")
	Log.Info("msg push comet start")
	<-done

	grpcServer.GracefulStop()
	Log.Infof("msg push message stoped")
}
