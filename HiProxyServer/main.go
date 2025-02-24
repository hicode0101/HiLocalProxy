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

	fmt.Println("--------------------------")
	fmt.Println("HiProxyServer v1.0.0")
	fmt.Println("--------------------------")
	fmt.Println("NeedAuth：", ServerConfig.NeedAuth)

	socks5ProxyServer := &Socks5ProxyServer{
		NeedAuth:   ServerConfig.NeedAuth,
		UserName:   ServerConfig.UserName,
		Password:   ServerConfig.Password,
		ListenAddr: ServerConfig.Socks5ListenAddr,
	}

	go socks5ProxyServer.RunSocks5Proxy()

	httpProxyServer := &HttpProxyServer{
		CertFile:   "cert.pem",
		KeyFile:    "key.pem",
		NeedAuth:   ServerConfig.NeedAuth,
		UserName:   ServerConfig.UserName,
		Password:   ServerConfig.Password,
		ListenAddr: ServerConfig.HttpListenAddr,
	}

	httpProxyServer.RunHttpProxy()

}

type Config struct {
	AppName          string `json:"AppName"`
	NeedAuth         bool   `json:"NeedAuth"`
	UserName         string `json:"UserName"`
	Password         string `json:"Password"`
	HttpListenAddr   string `json:"HttpListenAddr"`
	Socks5ListenAddr string `json:"Socks5ListenAddr"`
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
