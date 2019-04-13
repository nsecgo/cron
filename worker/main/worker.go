package main

import (
	"flag"
	"fmt"
	"github.com/nsecgo/crontab/worker"
	"os"
	"os/signal"
)

func main() {
	var (
		c        chan os.Signal
		confFile *string
		err      error
	)

	// 初始化命令行参数
	confFile = flag.String("config", "./worker.json", "worker.json")
	flag.Parse()

	// 加载配置
	if err = worker.InitConfig(*confFile); err != nil {
		goto ERR
	}

	// 服务注册
	if err = worker.InitRegister(); err != nil {
		goto ERR
	}

	// 启动日志协程
	if err = worker.InitLogSink(); err != nil {
		goto ERR
	}

	// 启动执行器
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}

	// 启动调度器
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}

	// 初始化任务管理器
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	return
ERR:
	fmt.Println(err)
}
