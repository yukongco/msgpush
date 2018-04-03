package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/yukongco/msgpush/web/conf"
	"github.com/yukongco/msgpush/web/controllers/push"
	. "github.com/yukongco/msgpush/web/logs"
	"github.com/yukongco/msgpush/web/router"
)

var (
	confInfo *conf.Config
	confPath = flag.String("config", "./conf/app.ini", "web profilePath")
)

func Init() error {
	flag.Parse()
	var err error
	confInfo, err = conf.InitConfig(confPath)
	if err != nil {
		fmt.Println("init config err: ", err)
		return fmt.Errorf("init config err: %v", err)
	}

	err = InitLog(confInfo.LogConf.LogPath, confInfo.LogConf.LogLevel)
	if err != nil {
		fmt.Println("Init msgpush web log is err: %v", err)
		return fmt.Errorf("Init msgpush web log is err: %v", err)
	}

	InitGrpc()

	return nil
}

func handleSignals(c chan os.Signal) {
	switch <-c {
	case syscall.SIGINT, syscall.SIGTERM:
		fmt.Println("Shutdown quickly, bye...")
	case syscall.SIGQUIT:
		fmt.Println("Shutdown gracefully, bye...")
		// do graceful shutdown
	}
	os.Exit(0)
}

func main() {
	//catch global panic
	defer func() {
		if err := recover(); err != nil {
			Log.Errorf("panic err: %v", err)
			fmt.Println("panic err: ", err)
		}
	}()

	err := Init()
	if err != nil {
		fmt.Println("main init err: %v", err)
		return
	}

	// graceful
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go handleSignals(sigs)

	if confInfo.Base.Env == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	Log.Info("msg push server is start")

	x := gin.New()
	router.ApiRouter(x)
	x.Run(confInfo.Base.HttpBind)
}

func InitGrpc() {
	err := push.InitPushGrpc()
	if err != nil {
		Log.Errorf("Init push grpc err: %v", err)
	}

	return
}
