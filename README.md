# HiLocalProxy

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-lightgrey)
![License](https://img.shields.io/badge/License-MIT-green)
![Version](https://img.shields.io/badge/Version-v1.0.3-blue)

把需要用户名密码验证的 SOCKS5 代理服务映射到本地，变成无需验证的本地代理服务。

Map the proxy service that needs to be verified to a local direct connection.

---

## 📖 工作原理

HiLocalProxy 在本地启动一个 SOCKS5 代理服务，收到客户端连接后，向上游 SOCKS5 服务器发起连接并自动完成用户名/密码认证（RFC 1929），之后将两侧数据透明转发：

```
本地应用（浏览器 / curl / 任意 SOCKS5 客户端）
    │  SOCKS5 握手（本地无需认证，监听端口如 :1081）
    ▼
HiLocalProxy
    │  SOCKS5 握手 + 用户名/密码认证（凭据来自 config.json）
    ▼
上游 SOCKS5 服务器（如 127.0.0.1:7080）──► 互联网
```

## ✨ 特性

- **本地零配置接入**：客户端无需任何认证即可使用；同时也兼容被配置为「必须填写用户名密码」的客户端（本地凭据不做校验，填写任意值即可）
- **自动上游认证**：按 RFC 1929 完成用户名/密码子协商，凭据只需在 config.json 中配置一次
- **目标地址全支持**：CONNECT 目标支持域名 / IPv4 / IPv6
- **协议解析健壮**：严格按字段长度解析 SOCKS5 握手报文，容忍半包、分包到达
- **跨平台、零依赖**：纯 Go 标准库实现，支持 Windows / Linux

## 🚀 快速开始

### 1. 配置

程序从**工作目录**读取 `config.json`，请与可执行文件放在同一目录下运行：

```json
{
  "AppName": "HiLocalProxy",
  "Socks5ListenAddr": ":1081",
  "UpSocks5Server": "127.0.0.1:7080",
  "UpUserName": "user",
  "UpPassword": "123456"
}
```

| 字段 | 说明 | 示例 |
|------|------|------|
| `AppName` | 应用名称，仅作标识 | `HiLocalProxy` |
| `Socks5ListenAddr` | 本地监听地址与端口 | `:1081` |
| `UpSocks5Server` | 上游 SOCKS5 服务器地址 | `127.0.0.1:7080` |
| `UpUserName` | 上游认证用户名 | `user` |
| `UpPassword` | 上游认证密码 | `123456` |

### 2. 构建

```bash
# 安装 Go 1.23+ 后，在源码目录执行
cd HiLocalProxy
go build -trimpath -ldflags "-w -s"
```

一键交叉编译 Windows / Linux 发布包（输出到 `build/` 目录）：

```bash
cd build
sh release_HiLocalProxy.sh v1.0.3
```

### 3. 运行

> ⚠️ 必须在 `config.json` 所在目录启动，否则程序会报错退出。

```bash
# Windows
./HiLocalProxy.exe

# Linux
./HiLocalProxy
```

启动成功后会看到：

```
--------------------------
HiLocalProxy v1.0.3
--------------------------
Forward to Upstream Socks5 proxy server： 127.0.0.1:7080
Starting Socks5 Proxy server on  :1081
```

### 4. 自动发布（GitHub Actions）

推送 `v` 开头的版本号 tag 即可自动构建发布，无需手动编译：

```bash
git tag v1.0.3
git push origin v1.0.3
```

CI 会自动完成以下流程（见 [.github/workflows/release.yml](.github/workflows/release.yml)）：

1. **版本号一致性校验**：tag 必须为 `vX.Y.Z` 格式，且与 `HiLocalProxy/main.go` 中 `AppVersion` 的值完全一致，否则构建失败并给出明确报错
2. **四平台构建**（`-trimpath -ldflags "-w -s"`，产物校验通过后再打包）：

   | 发布包 | 平台 | 构建方式 |
   |--------|------|----------|
   | `HiLocalProxy-linux-amd64-vX.Y.Z.zip` | Linux x86_64 | Ubuntu runner 上交叉编译 |
   | `HiLocalProxy-windows-amd64-vX.Y.Z.zip` | Windows x86_64 | Ubuntu runner 上交叉编译 |
   | `HiLocalProxy-macos-arm64-vX.Y.Z.zip` | macOS Apple Silicon (M 系列) | macOS ARM runner 上原生编译 |
   | `HiLocalProxy-macos-amd64-vX.Y.Z.zip` | macOS Intel (x86_64) | **macOS ARM 平台上交叉编译** |

3. 汇总 4 个发布包（每个内含二进制 + config.json）
4. 创建 GitHub Release 并上传全部产物

> ⚠️ 发版前请先同步修改 `HiLocalProxy/main.go` 中的 `AppVersion`，再打相同版本号的 tag。

## 🔌 客户端接入

**浏览器 / 系统代理**：代理类型选择 `SOCKS5`，地址 `127.0.0.1`，端口 `1081`（与 `Socks5ListenAddr` 一致）。若客户端强制要求填写用户名密码，填任意值即可。

**curl**：

```bash
# 通过本地代理访问任意网站（由上游出口访问）
curl -x socks5h://127.0.0.1:1081 https://www.example.com
```

> 💡 排查问题时可先直连上游验证其可用性，再走本地代理对比：
>
> ```bash
> curl -x socks5h://UpUserName:UpPassword@UpSocks5Server https://www.example.com
> ```

## 🔍 常见问题（日志排查）

| 日志信息 | 原因与处理 |
|----------|-----------|
| `Read config File err: ...` | 未在 `config.json` 所在目录启动，或文件不存在 |
| `Parse config File err: ...` | `config.json` 格式错误，请检查 JSON 语法 |
| `Listen fail: ...` | 本地监听端口被占用，修改 `Socks5ListenAddr` 或释放端口 |
| `无法连接到目标Socks5代理服务器: ...` | 上游未启动或地址端口配错，确认 `UpSocks5Server` 可达 |
| `不支持的 Socks5 握手响应: [5 255]` | 上游不支持用户名/密码认证方式（0x02），需调整上游配置 |
| `Socks5 认证失败: [1 1]` | 上游用户名或密码错误（STATUS=1），核对 `UpUserName` / `UpPassword` |
| `只支持Socks5代理` | 客户端流量不是 SOCKS5（例如把 HTTP 代理指向了本端口） |

## ⚠️ 已知限制

- 仅支持 SOCKS5 **CONNECT** 命令（TCP 流量）；不支持 UDP ASSOCIATE 与 BIND，因此不代理 UDP/QUIC 流量（浏览器会自动回退 TCP，一般无感知）
- 上游仅支持**用户名/密码认证**方式的 SOCKS5 服务器，不支持无认证上游或 GSSAPI 等
- 本地链路不做加密，请仅在可信网络（本机 / 内网）中使用

---

## 使用申明

本系列工具软件仅供工程师安全研究和技术交流使用，禁止用于商业用途。

👨‍💻我的昵称：犀利的远哥

✉️我的微信：hicode0101

💞️微信公众号：远哥说安全

如果你在我分享的工具中，遇到了问题，请联系我，我会及时回复。

如果你在我分享的工具上，有其它功能需求，也请告知我，我非常乐意尝试满足你的需求。

### 我的微信：

<img src="screenshot/weixin.png" width="200" />

---

### 微信公众号：

<img src="screenshot/gzh.png" width="300" />

---

## 📄 许可证

[MIT License](LICENSE) © 2025 HiCode0101
