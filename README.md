# HiLocalProxy
把需要验证的代理服务映射到本地直接连接

Map the proxy service that needs to be verified to a local direct connection

# HiLocalProxy（本地代理）

```
1、将需要用户名和密码验证的socks5代理服务映射到本地，变成无需验证的代理服务。

2、需要将上游socks5代理服务器的地址端口、用户名密码，正确配置到config.json文件中。

```

# HiProxyServer（代理服务器）

```
1、部署在内网服务器，或远端服务器上，作为代理服务器，提供代理服务。

2、支持 Http 用户名密码验证的代理服务，监听端口默认8080，可通过config.json配置文件修改端口号。

3、支持 Socks5 用户名密码验证的代理服务，监听端口默认1080，可通过config.json配置文件修改端口号。

```

