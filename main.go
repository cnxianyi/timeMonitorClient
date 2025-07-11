package main

import "timeMonitorClient/util"

func main() {

	util.InitEnv()

	go util.SetTaskPowerShell()

	go util.UploadOutPut()

	select {}
}
