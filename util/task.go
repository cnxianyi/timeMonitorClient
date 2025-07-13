package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-toast/toast"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
	"timeMonitorClient/global"
	"timeMonitorClient/types"
)

var (
	OutPutList [][]string
	GlobalTime time.Time = time.Now()
)

func SetTaskPowerShell() {
	ticker := time.Tick(1 * time.Second)
	for range ticker {
		res := PowerShellOutput()
		res = append(res, time.Now().Format(time.RFC3339Nano))
		OutPutList = append(OutPutList, res)
	}
}

func UploadOutPut() {
	ticker := time.Tick(5 * time.Second)

	for range ticker {
		l := len(OutPutList)

		if l > 30 {
			l = 30
		}

		cp := OutPutList[:l]
		var forms []types.UploadForm
		Name := os.Getenv("NAME")
		for i := range cp {
			form := types.UploadForm{
				Process:  cp[i][1],
				Title:    cp[i][0],
				Time:     cp[i][2],
				Username: Name,
			}
			forms = append(forms, form)

		}
		global.InfoJSON("upload:", forms)
		jsonData, err := json.Marshal(forms)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		err, res := Post(jsonData)
		if err != nil {
			fmt.Println("Error posting JSON:", err)
			return
		}

		OutPutList = OutPutList[l:]

		CheckData(res)
	}
}

func Post(jsonData []byte) (error, types.UploadDataResult) {
	client := &http.Client{Timeout: 10 * time.Second}
	baseUrl := os.Getenv("BASE_URL")

	req, err := http.NewRequest("POST", baseUrl+"upload", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err), types.UploadDataResult{}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err), types.UploadDataResult{}
	}
	defer resp.Body.Close() // Ensure the response body is closed.

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err), types.UploadDataResult{}
	}

	var res types.UploadResult

	// Now json.Unmarshal will be recognized
	err = json.Unmarshal(bodyBytes, &res) // Always check the error returned by Unmarshal!
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON response: %v", err), types.UploadDataResult{}
	}

	global.InfoJSON("upload result:", res)
	return nil, res.Data
}

func CheckData(result types.UploadDataResult) {

	if result.Notice != "" {
		Notice(result.Notice)
	}

	// 一小时
	if 60*59 < result.Lave && result.Lave < 60*61 {
		if computedTime(time.Now()) > 3*60 {
			GlobalTime = time.Now()
			Notice("剩余时间： 一小时")
		}
	}

	// 半小时
	if 60*29 < result.Lave && result.Lave < 60*31 {
		if computedTime(time.Now()) > 3*60 {
			GlobalTime = time.Now()
			Notice("剩余时间： 半小时")
		}
	}

	// 十分钟
	if 60*9 < result.Lave && result.Lave < 60*11 {
		if computedTime(time.Now()) > 3*60 {
			GlobalTime = time.Now()
			Notice("剩余时间： 十分钟")
		}
	}

	if result.Lave < 60 {
		Notice("即将关机")
		time.Sleep(1 * time.Second)
		ShotDown()
	}
}

func computedTime(currentTime time.Time) int {

	// 计算时间差：currentTime 减去 globalTime
	// 结果是 time.Duration 类型
	timeDifference := currentTime.Sub(GlobalTime)

	// 将时间差转换为秒
	// time.Duration 类型有一个 Seconds() 方法可以直接返回浮点数秒
	secondsDiff := timeDifference.Seconds()

	return int(secondsDiff)
}

func Notice(str string) {
	fmt.Println(str)

	notification := toast.Notification{
		AppID:   os.Getenv("APPID"), // 你的应用程序ID，这在Windows通知中心显示
		Title:   os.Getenv("Title"), // 通知标题
		Message: str,                // 通知内容
		Icon:    os.Getenv("Icon"),  // (可选) 图标路径，必须是绝对路径
		//Actions: []toast.Action{ // (可选) 添加按钮
		//	{"protocol", "打开", "https://www.google.com"}, // 点击打开网址
		//	{"", "忽略", ""}, // 纯文本按钮
		//},
		//ActivationArguments: "action=view", // (可选) 当点击通知主体时传递的参数
		// Audio: toast.Silent, // (可选) 设置通知声音为静音
		// Duration: toast.Long, // (可选) 设置通知显示时长 (Short 或 Long)
	}

	err := notification.Push()
	if err != nil {
		log.Fatalf("发送通知失败: %v", err)
	}
}

func ShotDown() {

	// 构建 shutdown 命令
	// /s 表示关机
	// /t 0 表示延迟0秒，即立即关机
	cmd := exec.Command("cmd", "/C", "shutdown", "/s", "/t", "0")

	// 执行命令
	err := cmd.Run()

	if err != nil {
		fmt.Printf("关闭电脑失败: %v\n", err)
		fmt.Println("请确保你以管理员权限运行此程序，或者命令有误。")
	} else {
		fmt.Println("关机命令已发送。电脑应该会立即关机。")
	}
}
