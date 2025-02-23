package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type HttpProxyServer struct {
	ListenAddr string
	CertFile   string
	KeyFile    string
	NeedAuth   bool
	UserName   string
	Password   string
}

// 验证用户名和密码
func (_self *HttpProxyServer) authenticate(r *http.Request, username, password string) bool {
	if !_self.NeedAuth {
		//如果不需要验证，直接返回true
		return true
	}

	auth := r.Header.Get("Proxy-Authorization")
	if auth == "" {
		return false
	}
	//fmt.Println("Auth：", auth)
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return false
	}
	decoded, err := url.QueryUnescape(parts[1])
	if err != nil {
		return false
	}

	user_pwd, err := base64.StdEncoding.DecodeString(decoded)
	if err != nil {
		fmt.Printf("解码出错: %v\n", err)
		return false
	}

	creds := strings.SplitN(string(user_pwd), ":", 2)
	if len(creds) != 2 {
		fmt.Println("UserAndPwd：", user_pwd)
		return false
	}

	authResult := username == creds[0] && password == creds[1]
	if !authResult {
		fmt.Println("auth failed.")
	}

	return authResult
}

// 处理HTTP请求
func (_self *HttpProxyServer) handleHTTP(w http.ResponseWriter, r *http.Request, username, password string) {
	if !_self.authenticate(r, username, password) {
		w.Header().Set("Proxy-Authenticate", "Basic realm=\"Proxy Authentication\"")
		http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
		return
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: r.URL.Scheme,
		Host:   r.URL.Host,
	})
	proxy.Transport = transport
	proxy.ServeHTTP(w, r)
}

// 处理HTTPS连接
func (_self *HttpProxyServer) handleHTTPS(w http.ResponseWriter, r *http.Request, username, password string) {
	if !_self.authenticate(r, username, password) {
		w.Header().Set("Proxy-Authenticate", "Basic realm=\"Proxy Authentication\"")
		http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
		return
	}

	// 连接到目标服务器
	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, "Failed to connect to destination", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}
	go _self.transfer(destConn, clientConn)
	go _self.transfer(clientConn, destConn)
}

// 数据传输
func (_self *HttpProxyServer) transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func (_self *HttpProxyServer) RunHttpProxy() {

	// 创建HTTP服务器
	server := &http.Server{
		Addr: _self.ListenAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(GetCurrentTime(), " Http Proxy to：", r.RequestURI)
			if r.Method == http.MethodConnect {
				_self.handleHTTPS(w, r, _self.UserName, _self.Password)
			} else {
				_self.handleHTTP(w, r, _self.UserName, _self.Password)
			}
		}),
	}

	log.Println("Starting Http Proxy server on ", _self.ListenAddr)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}

}

func (_self *HttpProxyServer) RunHttpsProxy() {
	_CertFile := fmt.Sprint("./cert/", _self.CertFile)
	_KeyFile := fmt.Sprint("./cert/", _self.KeyFile)

	// 加载证书
	cert, err := tls.LoadX509KeyPair(_CertFile, _KeyFile)
	if err != nil {
		log.Fatalf("Failed to load TLS certificate: %v", err)
	}

	// 创建HTTP服务器
	server := &http.Server{
		Addr: _self.ListenAddr,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(GetCurrentTime(), " Https Proxy to：", r.RequestURI)
			if r.Method == http.MethodConnect {
				_self.handleHTTPS(w, r, _self.UserName, _self.Password)
			} else {
				_self.handleHTTP(w, r, _self.UserName, _self.Password)
			}
		}),
	}

	log.Println("Starting Https Proxy server on ", _self.ListenAddr)

	err = server.ListenAndServeTLS(_CertFile, _KeyFile)
	if err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}

}
