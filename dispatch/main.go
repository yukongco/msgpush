package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yukongco/msgpush/dispatch/conf"
	. "github.com/yukongco/msgpush/dispatch/logs"
	pb "github.com/yukongco/msgpush/dispatch/proto"
	"github.com/yukongco/msgpush/dispatch/schedule"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	confInfo *conf.Config
	confPath = flag.String("config", "./conf/app.ini", "dispatch profilePath")
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

	err = InitGrpc()
	if err != nil {
		fmt.Println("init grpc err: ", err)
		return fmt.Errorf("init grpc err: %v", err)
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
	pb.RegisterDispatchServiceServer(grpcServer, &dispatch.DispatchSer{})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = grpcServer.Serve(grpcLis); err != nil {
			Log.Errorf("grpc error:%v", err)
			panic(err)
		}
	}()

	go dispatch.HeartRun()

	fmt.Println("msg push dispatch start")
	Log.Info("msg push dispatch start")
	<-done

	grpcServer.GracefulStop()
	Log.Infof("msg push dispatch stoped")
}

func InitGrpc() error {
	err := dispatch.InitMsgGrpc()
	if err != nil {
		return err
	}

	return nil
}
