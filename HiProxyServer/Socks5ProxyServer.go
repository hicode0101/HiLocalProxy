package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	NO_AUTH       = []byte{0x05, 0x00}
	USERPASS_AUTH = []byte{0x05, 0x02}

	AUTH_SUCCESS = []byte{0x05, 0x00}
	AUTH_FAILED  = []byte{0x05, 0x01}

	CONNECT_SUCCESS = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

type Socks5ProxyServer struct {
	ListenAddr string
	NeedAuth   bool
	UserName   string
	Password   string
}

func (_self *Socks5ProxyServer) proxyHandler(conn net.Conn) {

	headBuf := make([]byte, 2)
	_, err := conn.Read(headBuf)

	if err != nil {
		fmt.Println(conn.RemoteAddr().String(), " Read error: ", err)
		return
	}

	//VER
	if headBuf[0] != 0x05 {
		fmt.Println("只支持Socks5代理")
		conn.Close()
		return
	}

	//NMETHODS
	nMethods := headBuf[1]
	//METHODS
	methods := make([]byte, nMethods)
	if n, err := conn.Read(methods); n != int(nMethods) || err != nil {
		fmt.Println("Get methods error", err)
		conn.Close()
		return
	}

	if !_self.NeedAuth {
		//回复2个byte，表示 VER=0x05 METHOD=0x00
		conn.Write(NO_AUTH)
	} else {
		//告诉客户端，需要用户名密码验证
		conn.Write(USERPASS_AUTH)

		authBuf := make([]byte, 520)
		_, err := conn.Read(authBuf)
		if err != nil {
			fmt.Println(conn.RemoteAddr().String(), " Read error: ", err)
			return
		}

		//fmt.Println("读取到验证信息：", n, authBuf)

		//ver := uint8(authBuf[0])
		uLen := uint8(authBuf[1])

		//fmt.Println("Auth ver：", ver, "uLen：", uLen, "authBuf：", authBuf)

		uname := string(authBuf[2 : 2+uLen])
		pLen := uint8(authBuf[2+uLen])
		pass := string(authBuf[2+1+uLen : 2+1+uLen+pLen])

		//fmt.Println("Request Auth u：", uname, " P：", pass)
		//校验用户名和密码
		if _self.UserName == uname && _self.Password == pass {
			//fmt.Println("AUTH_SUCCESS")
			conn.Write(AUTH_SUCCESS)
		} else {
			//用户名密码不对，验证失败关闭连接
			conn.Write(AUTH_FAILED)
			conn.Close()
			fmt.Println("AUTH_FAILED，Closed conn")
			return
		}

	}

	b := make([]byte, 256)
	n, err := conn.Read(b)
	//fmt.Println(conn.RemoteAddr().String(), "len：", n, "reqBuf b：", b)

	var host string
	switch b[3] {
	case 0x01: //IP V4
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
	case 0x03: //domain
		host = string(b[5 : n-2]) //b[4] length of domain
	case 0x04: //IP V6
		host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
	default:
		return
	}
	port := strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
	targetAddr := net.JoinHostPort(host, port)

	fmt.Println(GetCurrentTime(), " Socks5 Proxy to：", targetAddr)

	network := "tcp"
	if b[1] == 0x03 {
		network = "udp"
	}
	//reqServer, err := net.Dial(network, targetAddr)
	reqServer, err := net.DialTimeout(network, targetAddr, 60*time.Second)
	if err != nil {
		fmt.Println("reqServer err:", err)
		return
	}
	conn.Write(CONNECT_SUCCESS)

	go func() {
		_, err := io.Copy(reqServer, conn)
		if err != nil {
			//logger.Errors("数据转发到目标端口异常：", err)
			conn.Close()
		}
	}()

	go func() {
		_, err := io.Copy(conn, reqServer)
		if err != nil {
			//logger.Errors("返回响应数据异常：", err)
			conn.Close()
		}
	}()

}

func (_self *Socks5ProxyServer) RunSocks5Proxy() {
	// 监听本地端口
	listener, err := net.Listen("tcp", _self.ListenAddr)
	if err != nil {
		log.Fatalf("Listen fail: %v", err)
	}

	fmt.Println("Starting Socks5 Proxy server on ", _self.ListenAddr)
	for {
		// 接受客户端连接
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept conn fail: %v", err)
			continue
		}
		// 处理客户端连接
		go _self.proxyHandler(conn)
	}
}
