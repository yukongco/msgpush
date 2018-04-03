package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yukongco/msgpush/comet/conf"
	. "github.com/yukongco/msgpush/comet/logs"
	pb "github.com/yukongco/msgpush/comet/proto"
	"github.com/yukongco/msgpush/comet/push"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	confInfo *conf.Config
	confPath = flag.String("config", "./conf/app.ini", "comet profilePath")
)

func Init() error {
	flag.Parse()
	var err error
	fmt.Println("main init")

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

	// start cron
	cronRes := push.InitCron()
	cronRes.Start()
	defer cronRes.Stop()

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
	pb.RegisterMsgPushServiceServer(grpcServer, &push.MsgPushService{})

	hub := push.NewHub()
	go hub.Run()

	go func() {
		http.HandleFunc("/", push.ServerHome)
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			push.ServerWs(hub, w, r)
		})

		err := http.ListenAndServe(confInfo.WebsocketConf.HostPort, nil)
		if err != nil {
			Log.Error("ListenAndServe err: ", err)
			panic(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = grpcServer.Serve(grpcLis); err != nil {
			Log.Errorf("grpc error:%v", err)
			panic(err)
		}
	}()

	// register ip to redis
	go push.RegisterCometIp()

	fmt.Println("msg push comet start")
	Log.Info("msg push comet start")
	<-done

	grpcServer.GracefulStop()
	Log.Infof("msg push comet stoped")
}

func InitGrpc() error {
	err := push.InitMsgGrpc()
	if err != nil {
		return err
	}

	return nil
}
