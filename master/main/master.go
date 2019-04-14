package main

import (
	"flag"
	"github.com/nsecgo/cron/master"
	"log"
	"os"
	"os/signal"
)

func main() {
	var (
		c        chan os.Signal
		confFile *string
		err      error
	)

	confFile = flag.String("config", "./master.json", "指定master.json")
	flag.Parse()

	// 加载配置
	if err = master.InitConfig(*confFile); err != nil {
		goto ERR
	}

	// 初始化服务发现模块
	if err = master.InitWorkerMgr(); err != nil {
		goto ERR
	}

	// 日志管理器
	if err = master.InitLogMgr(); err != nil {
		goto ERR
	}

	//  任务管理器
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}

	// 启动Api HTTP服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}
	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	return

ERR:
	log.Println(err)
}
