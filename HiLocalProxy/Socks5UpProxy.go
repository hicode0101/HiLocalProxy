package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

type Socks5UpProxy struct {
	ListenAddr string
	UpServer   string
	UpUserName string
	UpPassword string
}

func (_self *Socks5UpProxy) proxyHandler(conn net.Conn) {

	headBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, headBuf); err != nil {
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
	if _, err := io.ReadFull(conn, methods); err != nil {
		fmt.Println("Get methods error", err)
		conn.Close()
		return
	}

	//按客户端提供的认证方法选择本地认证方式：优先无认证；仅提供用户名密码时也接受。
	//本地凭据不做校验，真实的用户名密码认证发生在上游服务器
	methodChosen := byte(0xFF)
	for _, m := range methods {
		if m == 0x00 {
			methodChosen = 0x00
			break
		}
		if m == 0x02 {
			methodChosen = 0x02
		}
	}
	if methodChosen == 0xFF {
		conn.Write([]byte{0x05, 0xFF})
		conn.Close()
		return
	}
	conn.Write([]byte{0x05, methodChosen})

	if methodChosen == 0x02 {
		if err := _self.readDiscardLocalAuth(conn); err != nil {
			fmt.Println("本地用户名密码子协商读取失败: ", err)
			conn.Close()
			return
		}
	}

	// 连接到目标Socks5代理服务器
	proxyConn, err := net.Dial("tcp", _self.UpServer)
	if err != nil {
		log.Printf("无法连接到目标Socks5代理服务器: %v", err)
		conn.Close()
		return
	}

	// 向目标Socks5代理服务器进行认证
	if err = _self.authenticateWithProxy(proxyConn); err != nil {
		log.Printf("目标Socks5代理服务器认证失败: %v", err)
		proxyConn.Close()
		conn.Close()
		return
	}

	//fmt.Println("auth success.")

	// 开始双向数据转发
	go func() {
		_, _ = io.Copy(conn, proxyConn)
		conn.Close()
		proxyConn.Close()
	}()

	go func() {
		_, _ = io.Copy(proxyConn, conn)
		conn.Close()
		proxyConn.Close()
	}()

}

// readDiscardLocalAuth 读取并丢弃本地子协商的用户名密码（RFC 1929），本地不做校验
func (_self *Socks5UpProxy) readDiscardLocalAuth(conn net.Conn) error {
	head := make([]byte, 2) // VER ULEN
	if _, err := io.ReadFull(conn, head); err != nil {
		return err
	}
	uname := make([]byte, head[1])
	if _, err := io.ReadFull(conn, uname); err != nil {
		return err
	}
	plenBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, plenBuf); err != nil {
		return err
	}
	passwd := make([]byte, plenBuf[0])
	if _, err := io.ReadFull(conn, passwd); err != nil {
		return err
	}
	_, err := conn.Write([]byte{0x01, 0x00})
	return err
}

func (_self *Socks5UpProxy) authenticateWithProxy(proxyConn net.Conn) error {
	// 发送支持的认证方法，这里只支持用户名密码认证
	handshakeRequest := []byte{
		0x05, // Socks5 版本
		0x01, // 支持的认证方法数量
		0x02, // 用户名密码认证
	}
	_, err := proxyConn.Write(handshakeRequest)
	if err != nil {
		return err
	}

	// 读取服务器响应
	response := make([]byte, 2)
	_, err = io.ReadFull(proxyConn, response)
	if err != nil {
		return err
	}
	if response[0] != 0x05 || response[1] != 0x02 {
		return fmt.Errorf("不支持的 Socks5 握手响应: %v", response)
	}

	// 构建认证请求
	authRequest := []byte{
		0x01, // 认证协议版本
		byte(len(_self.UpUserName)),
	}
	authRequest = append(authRequest, []byte(_self.UpUserName)...)
	authRequest = append(authRequest, byte(len(_self.UpPassword)))
	authRequest = append(authRequest, []byte(_self.UpPassword)...)

	// 发送认证请求
	_, err = proxyConn.Write(authRequest)
	if err != nil {
		return err
	}

	// 读取认证响应
	response = make([]byte, 2)
	_, err = io.ReadFull(proxyConn, response)
	if err != nil {
		return err
	}
	// RFC 1929：用户名密码子协商响应的版本号为 0x01，STATUS=0x00 表示成功
	if response[0] != 0x01 || response[1] != 0x00 {
		return fmt.Errorf("Socks5 认证失败: %v", response)
	}
	return nil

}

func (_self *Socks5UpProxy) sendRequestToProxy(proxyConn net.Conn, targetAddr string) error {
	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	var addrType byte
	var addrData []byte
	if ip.To4() != nil {
		addrType = 0x01
		addrData = ip.To4()
	} else if ip.To16() != nil {
		addrType = 0x04
		addrData = ip.To16()
	} else {
		addrType = 0x03
		addrData = []byte(host)
	}
	request := []byte{0x05, 0x01, 0x00, addrType}
	request = append(request, addrData...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	request = append(request, portBytes...)
	_, err = proxyConn.Write(request)
	return err
}

// relayResponse 从目标Socks5代理服务器读取响应并发送给客户端
func (_self *Socks5UpProxy) relayResponse(proxyConn, clientConn net.Conn) error {
	respBuf := bufio.NewReader(proxyConn)
	resp, err := respBuf.Peek(10)
	if err != nil {
		return err
	}
	if resp[0] != 0x05 || resp[1] != 0x00 {
		return fmt.Errorf("目标Socks5代理服务器响应失败")
	}
	_, err = io.CopyN(clientConn, respBuf, 10)
	return err
}

func (_self *Socks5UpProxy) RunSocks5Proxy() {
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
