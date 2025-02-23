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

func main() {

	loadConfig()

	fmt.Println("HiLocalProxy")
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
	//加载配置文件
	workPath, _ := os.Getwd()

	configData, configErr := os.ReadFile(fmt.Sprint(workPath, "/config.json"))
	if configErr != nil {
		fmt.Println("Read config File err:", configErr)
		return
	}

	ServerConfig = new(Config)
	FromJson(configData, ServerConfig)
	//fmt.Println(AppConfigS.AppName)
}
