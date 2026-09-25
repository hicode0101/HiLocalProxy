package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var (
	ServerConfig *Config
)

// AppVersion 版本号：必须与发布 tag（如 v1.0.3）保持一致，CI 在打 tag 时会校验
var AppVersion = "v1.0.3"

func main() {

	loadConfig()

	fmt.Println("--------------------------")
	fmt.Println("HiLocalProxy " + AppVersion)
	fmt.Println("--------------------------")
	fmt.Println("Forward to Upstream Socks5 proxy server：", ServerConfig.UpSocks5Server)

	socks5UpProxy := &Socks5UpProxy{
		ListenAddr: ServerConfig.Socks5ListenAddr,
		UpServer:   ServerConfig.UpSocks5Server,
		UpUserName: ServerConfig.UpUserName,
		UpPassword: ServerConfig.UpPassword,
	}

	socks5UpProxy.RunSocks5Proxy()
}

type Config struct {
	AppName          string `json:"AppName"`
	Socks5ListenAddr string `json:"Socks5ListenAddr"`
	UpSocks5Server   string `json:"UpSocks5Server"`
	UpUserName       string `json:"UpUserName"`
	UpPassword       string `json:"UpPassword"`
}

func GetCurrentTime() string {

	return time.Now().Format("2006-01-02 15:04:05")
}

func FromJson(data []byte, t interface{}) error {
	return json.Unmarshal(data, t)
}

func loadConfig() {
	//加载配置文件（程序从工作目录读取 ./config.json，因此必须在配置文件所在目录运行）
	workPath, _ := os.Getwd()

	configData, configErr := os.ReadFile(fmt.Sprint(workPath, "/config.json"))
	if configErr != nil {
		fmt.Println("Read config File err:", configErr)
		os.Exit(1)
	}

	ServerConfig = new(Config)
	if err := FromJson(configData, ServerConfig); err != nil {
		fmt.Println("Parse config File err:", err)
		os.Exit(1)
	}
}
