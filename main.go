package main

import (
	"fmt"
	"os"
	"timeMonitorClient/global"
	"timeMonitorClient/util"
)

func main() {
	// 初始化环境变量
	util.InitEnv()

	// 初始化日志系统
	global.InitLogger()

	s := fmt.Sprintf("程序启动成功,用户:%s", os.Getenv("USER"))
	global.Info(s)

	go util.SetTaskPowerShell()

	go util.UploadOutPut()

	select {}
}
