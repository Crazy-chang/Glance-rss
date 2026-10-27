# Glance RSS 后端部署手册

> 部署目标：把 Go 后端跑在腾讯云轻量服务器上，对外通过 `https://glancelog.cn` 提供 HTTPS API 给微信小程序调用。
>
> 本文档同时是**操作手册** + **排错指南** + **概念笔记**，按顺序照做即可复现整套部署。

---

## 0. 总览：整条链路长什么样

```
微信小程序 ──HTTPS(443)──► Nginx ──HTTP(8080)──► glance-server ──► SQLite
                          ▲                        ▲
                          │                        │
                    装 SSL 证书                systemd 守护进程
                    反向代理                   开机自启+崩溃重启
```

**核心思想**：每个组件只做一件事，串起来形成完整链路。

| 组件 | 职责 | 为什么需要它 |
|---|---|---|
| **glance-server** | 业务逻辑、操作数据库 | 你的代码 |
| **systemd** | 进程守护 | SSH 断开不死、开机自启、崩溃重启 |
| **Nginx** | 反向代理 + SSL 终结 | Go 程序本身不处理 HTTPS，Nginx 解密后转发 |
| **SSL 证书** | 域名身份凭证 | 小程序强制 HTTPS；浏览器显示 🔒 |
| **腾讯云防火墙** | 网络层流量控制 | 只放通需要的端口 |

---

## 1. 准备工作（信息收集）

部署前必须确认的信息：

| 项 | 我的值 | 在哪拿 |
|---|---|---|
| 服务器公网 IP | `43.138.218.39` | 腾讯云控制台 → 轻量服务器 → 实例详情 |
| 服务器操作系统 | Ubuntu 24.04 | 购买时选择 |
| SSH 登录方式 | ubuntu 用户 + 密码 | 控制台「重置密码」 |
| 域名 | `glancelog.cn` | 腾讯云域名服务 |
| 域名是否备案 | 必须已备案 | 工信部 ICP 备案（未备案 80/443 不通） |
| 小程序 AppID/Secret | wx*****/***** | mp.weixin.qq.com → 开发管理 |
| Git 仓库 | `github.com/Crazy-chang/Glance-rss` | GitHub |

**关键硬门槛**：域名必须**已备案**，否则腾讯云会拦 80/443 端口，怎么配都不通。

---

## 2. 服务器环境准备

### 2.1 安装 Go

apt 源里的 Go 版本太旧（1.21），项目要求 1.26+，所以从官网下二进制包：

```bash
cd /tmp
wget https://go.dev/dl/go1.26.6.linux-amd64.tar.gz

# 解压到 /usr/local（系统目录，需要 sudo）
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.6.linux-amd64.tar.gz

# 把 /usr/local/go/bin 加入 PATH，永久生效
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile

# 当前终端立即生效
export PATH=$PATH:/usr/local/go/bin

go version  # 验证：go version go1.26.6 linux/amd64
```

**踩过的坑**：
- ❌ 不用 `sudo` 解压到 `/usr/local` → 静默失败
- ❌ `export PATH= "/usr/..."`（等号后有空格）→ 语法错误，PATH 被清空，所有命令找不到
- ❌ 改了 `/etc/profile` 当前终端没生效 → 用 `export` 临时设一次或重开终端

**知识点**：
- `/usr/local` 是用户级软件的标准安装位置（区别于 `/usr` 系统软件）
- `/etc/profile` 是全局 shell 启动脚本，所有用户登录都会读
- 修改 PATH 后必须 `source /etc/profile` 或重新登录才生效

### 2.2 安装 Git

```bash
git --version || sudo apt update && sudo apt install -y git
```

### 2.3 配置 Go 模块代理（国内必做）

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

不设的话 `go mod download` 会拉不下来依赖（`proxy.golang.org` 在国内不稳定）。

---

## 3. Git 部署（手动版 CI/CD）

### 3.1 部署模式说明

这步用的模式叫「**Git 部署**」：

```
本地 push ──► GitHub ──► 服务器 git pull ──► 服务器本地 go build ──► 重启服务
```

**对比其他模式**：

| 模式 | 构建在哪 | 服务器需要 | 适用 |
|---|---|---|---|
| scp 源码 + 服务器 build | 服务器 | Go 环境 | 学习/小项目（**当前用**） |
| 本地交叉编译 + scp 二进制 | 本地 | 啥都不要 | 快速验证 |
| CI/CD | CI 服务器 | 啥都不要 | 团队/正式项目 |
| Docker | 任意 | Docker | 微服务/复杂环境 |

### 3.2 clone 代码

```bash
cd ~
git clone https://github.com/Crazy-chang/Glance-rss.git
cd Glance-rss
```

**注意**：服务器上这份代码跟你本地、GitHub 上是**同一个仓库的三个副本**。更新代码 = `git pull` + 重新 build + 重启服务，这就是最小化的"部署流水线"。

### 3.3 编译

```bash
cd ~/Glance-rss/backend
go mod download           # 下载所有依赖到本地缓存
go build -o glance-server .
ls -lh glance-server      # 应该看到约 39MB 的二进制
```

**为什么在服务器上直接 build 而不是本地交叉编译**：
- 服务器就是 Linux amd64，不需要交叉编译参数
- 学习目的：看清楚「代码 → 可执行文件」这个构建过程
- CI/CD 阶段的"自动构建"就是在 CI 服务器上做同样的事

### 3.4 配置文件分离（重要安全原则）

**密钥、生产配置绝对不能进 Git**——一旦提交，永远留在 git 历史里。

项目已支持 `GLANCE_CONFIG` 环境变量指定配置文件，且 `.gitignore` 已排除 `config.local.yaml`。

```bash
cd ~/Glance-rss/backend

# 复制模板为本地配置
cp config/config.yaml config/config.local.yaml

# 生成随机 JWT 密钥
openssl rand -base64 32
# 输出一串随机字符，复制下来下一步用
```

编辑 `config/config.local.yaml`（用 `nano` 或 `vim`），改 4 处：

```yaml
server:
  port: "8080"
  base_url: "https://glancelog.cn"          # 改成你的域名，必须 https

jwt:
  secret: "刚才openssl生成的随机字符串"      # 改这里

wechat:
  appid: "你的小程序AppID"
  secret: "你的小程序AppSecret"
```

**nano 操作**：方向键移动光标 → 改完按 `Ctrl+O`（保存）→ 回车 → `Ctrl+X`（退出）。

### 3.5 手动验证一次

```bash
GLANCE_CONFIG=config/config.local.yaml ./glance-server
```

预期输出：
```
... INFO 数据库就绪 path=data/glance.db
... INFO Glance 后端启动 port=8080
```

**另开一个终端**测试：

```bash
curl http://localhost:8080/api/source/list
# 预期：{"code":401,"data":{},"msg":"登录已失效，请重新登录"}
```

看到 401 说明服务正确启动并拒绝了未登录请求。`Ctrl+C` 停掉。

**知识点**：
- `GLANCE_CONFIG=... ./glance-server` 这种写法是**临时**给这一条命令设置环境变量
- `./` 表示当前目录。Linux 不会自动在当前目录找可执行文件（安全设计）

---

## 4. systemd 进程守护

### 4.1 为什么需要 systemd

直接 `./glance-server` 启动的问题：
- SSH 一断服务就死
- 服务器重启后要手动起
- 服务崩了没人拉

**systemd** 是 Linux 的"服务管家"，解决以上所有问题。

### 4.2 创建 service 文件

```bash
sudo nano /etc/systemd/system/glance-backend.service
```

完整内容：

```ini
[Unit]
Description=Glance RSS Backend
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/Glance-rss/backend
Environment="GLANCE_CONFIG=config/config.local.yaml"
ExecStart=/home/ubuntu/Glance-rss/backend/glance-server
Restart=always
RestartSec=5

StandardOutput=journal
StandardError=journal
SyslogIdentifier=glance-backend

[Install]
WantedBy=multi-user.target
```

### 4.3 启动 + 开机自启

```bash
sudo systemctl daemon-reload              # 每次改 service 文件都要做
sudo systemctl start glance-backend       # 启动
sudo systemctl enable glance-backend      # 开机自启
sudo systemctl status glance-backend      # 查看状态
```

**预期关键输出**：

```
Loaded: loaded (... glance-backend.service; enabled; ...)   ← enabled = 开机自启
Active: active (running) since ...                          ← 正在运行
Main PID: 1696503 (glance-server)
```

### 4.4 常用管理命令

```bash
sudo systemctl start glance-backend       # 启动
sudo systemctl stop glance-backend        # 停止
sudo systemctl restart glance-backend     # 重启（改代码后用）
sudo systemctl status glance-backend      # 看状态
sudo journalctl -u glance-backend -f      # 实时看日志（Ctrl+C 退出）
sudo journalctl -u glance-backend -n 100  # 看最近 100 行日志
```

### 4.5 service 文件关键配置解读

| 配置 | 含义 |
|---|---|
| `After=network.target` | 等网络就绪后再启动（Web 服务必需） |
| `User=ubuntu` | 用 ubuntu 用户运行，**不要用 root**（最小权限原则） |
| `WorkingDirectory` | 工作目录。配置文件和 SQLite 用相对路径，所以必须设 |
| `Restart=always` | 无论什么原因退出都自动重启 |
| `RestartSec=5` | 重启前等 5 秒，避免崩溃循环刷屏 |
| `StandardOutput=journal` | 日志进 systemd 日志系统，用 `journalctl` 看 |

---

## 5. Nginx 反向代理 + SSL

### 5.1 为什么需要 Nginx

`glance-server` 只会 HTTP，**不会处理 HTTPS**。需要 Nginx 在最前面：

1. 接收 443 端口的 HTTPS 请求
2. 用 SSL 证书解密成 HTTP
3. 转发给 8080 端口的 glance-server
4. 把后端响应加密回 HTTPS 返回给客户端

这种架构叫**反向代理**（Reverse Proxy），是 Web 服务标准模式。

### 5.2 安装 + 验证

```bash
sudo apt install -y nginx
sudo systemctl status nginx   # 应该 active (running)
curl http://localhost          # 应该看到 "Welcome to nginx!"
```

### 5.3 申请 SSL 证书（腾讯云控制台操作）

1. 腾讯云控制台 → 搜索「**SSL 证书**」→ 我的证书 → 申请免费证书
2. 填表：
   - 证书品牌：`TrustAsia`
   - 证书类型：`单域名`
   - 绑定域名：`glancelog.cn`
   - 验证方式：**自动添加 DNS**（域名在腾讯云时最快）
3. 等审核（DV 证书一般几分钟到 1 小时）
4. 审核通过 → 下载 → 选 **Nginx** 格式
5. 解压得到 4 个文件，只需要其中 2 个：
   - `glancelog.cn_bundle.crt`（证书，公钥）
   - `glancelog.cn.key`（私钥，**绝不能泄露**）

### 5.4 上传证书到服务器

在**你电脑的终端**：

```bash
cd ~/Downloads/glancelog.cn_nginx
scp glancelog.cn_bundle.crt ubuntu@43.138.218.39:/tmp/
scp glancelog.cn.key ubuntu@43.138.218.39:/tmp/
```

切到 **TAT 终端**：

```bash
sudo mkdir -p /etc/nginx/ssl
sudo mv /tmp/glancelog.cn_bundle.crt /etc/nginx/ssl/
sudo mv /tmp/glancelog.cn.key /etc/nginx/ssl/

# 私钥 600（只有 owner 能读写），证书 644（其他人可读）
sudo chmod 600 /etc/nginx/ssl/glancelog.cn.key
sudo chmod 644 /etc/nginx/ssl/glancelog.cn_bundle.crt

ls -l /etc/nginx/ssl/  # 验证
```

### 5.5 配置 Nginx 站点

```bash
sudo nano /etc/nginx/sites-available/glance
```

完整内容：

```nginx
# HTTP 自动跳转 HTTPS
server {
    listen 80;
    server_name glancelog.cn;
    return 301 https://$host$request_uri;
}

# HTTPS 反向代理
server {
    listen 443 ssl;
    server_name glancelog.cn;

    ssl_certificate     /etc/nginx/ssl/glancelog.cn_bundle.crt;
    ssl_certificate_key /etc/nginx/ssl/glancelog.cn.key;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_read_timeout 60s;
        client_max_body_size 10m;
    }
}
```

启用 + 重载：

```bash
# 建立软链接启用站点
sudo ln -s /etc/nginx/sites-available/glance /etc/nginx/sites-enabled/glance

# ⚠️ 删掉默认站点，否则会抢占请求
sudo rm /etc/nginx/sites-enabled/default

# 测试配置语法（必须每次改完都做）
sudo nginx -t

# 平滑重载
sudo systemctl reload nginx
```

### 5.6 Nginx 目录约定

| 目录 | 用途 |
|---|---|
| `sites-available/` | **可用**站点的配置仓库（不直接生效） |
| `sites-enabled/` | **启用**站点的软链接（Nginx 加载这里） |
| `nginx.conf` | 主配置，会 include `sites-enabled/*` |

**启用/停用站点 = 增删软链接**，配置文件本身不动。这是 Debian/Ubuntu 系 Nginx 的经典约定。

---

## 6. 防火墙（腾讯云控制台）

腾讯云轻量服务器的防火墙**默认只开少数端口**，需要手动加 443。

控制台 → 轻量服务器 → 实例 → 「防火墙」→ 添加规则：

| 应用类型 | 来源 | 协议 | 端口 | 策略 |
|---|---|---|---|---|
| HTTP (80) | 0.0.0.0/0 | TCP | 80 | 允许 |
| **HTTPS (443)** | **0.0.0.0/0** | **TCP** | **443** | **允许** |
| Linux 登录 (22) | 0.0.0.0/0 | TCP | 22 | 允许 |

**踩过的坑**：默认有条 `Ping - ICMP - ALL` 规则，**"ALL" 指的是 ICMP 协议全开，不是"所有端口"**。TCP 443 仍然要单独加。

---

## 7. 验证

### 7.1 命令行验证

```bash
# 服务器本地（绕过 HTTPS，验证后端）
curl http://localhost:8080/api/source/list

# 走完整链路（HTTPS + 域名 + Nginx + 后端）
curl https://glancelog.cn/api/source/list
```

预期返回：
```json
{"code":401,"data":{},"msg":"登录已失效，请重新登录"}
```

### 7.2 浏览器验证

打开 `https://glancelog.cn/api/source/list`，应该看到：
- 地址栏 🔒 锁图标
- 页面显示 JSON 响应

### 7.3 关于 curl -I 返回 404 的说明

```bash
curl -I https://glancelog.cn/api/source/list
# 返回 404，但 curl https://... 返回 401，为什么？
```

`-I` 让 curl 发 **HEAD 请求**（不取 body）。Gin 后端只注册了 `GET` 路由，**没注册 HEAD**，所以 HEAD 请求返回 404 是正常的。**测试 API 慎用 `-I`**。

---

## 8. 排错决策树

遇到问题按这个顺序排查：

### 症状：`curl: (7) Failed to connect to ... port 443`

```
Nginx 在监听 443 吗？
└─ sudo ss -tlnp | grep 443
   ├─ 没有 → Nginx 配置没生效，看 nginx -t 输出
   └─ 有 → 腾讯云防火墙 443 没开
      └─ 控制台 → 实例 → 防火墙 → 添加 TCP:443
```

### 症状：`curl: (6) Could not resolve host`

```
域名解析对吗？
└─ nslookup glancelog.cn
   ├─ 解析到错的 IP → 改 DNS 解析
   └─ 解析不到 → 域名未实名/未配置 DNS
```

### 症状：浏览器显示证书错误

```
证书装对了吗？
└─ echo | openssl s_client -connect glancelog.cn:443 -servername glancelog.cn 2>/dev/null | openssl x509 -noout -dates -subject
   ├─ 域名不匹配 → 证书签发的域名 ≠ 访问的域名
   ├─ 已过期 → 续期（免费证书 90 天）
   └─ 不受信任 → 用了自签名证书，要换 CA 签发的
```

### 症状：返回 Nginx 默认 404 页面（text/plain）

```
是不是默认站点抢占了？
└─ ls /etc/nginx/sites-enabled/
   └─ 有 default → sudo rm /etc/nginx/sites-enabled/default && sudo systemctl reload nginx
```

### 症状：返回 502 Bad Gateway

```
后端挂了吗？
└─ sudo systemctl status glance-backend
   ├─ inactive/failed → sudo journalctl -u glance-backend -n 50 看崩溃原因
   └─ active → 看端口监听：sudo ss -tlnp | grep 8080
```

### 症状：服务起来了但 API 返回 500

```
看后端日志
└─ sudo journalctl -u glance-backend -f
   常见：数据库权限、配置文件路径错、依赖服务挂了
```

---

## 9. 后续升级路线

当前是「Git 部署」，学习路径下一步：

### 9.1 升级到 CI/CD（GitHub Actions）

把「git pull → build → restart」写成 `.github/workflows/deploy.yml`，push 后自动执行：

```yaml
# 大致结构（具体后续做）
on: push
jobs:
  deploy:
    steps:
      - 构建二进制
      - scp 到服务器
      - ssh 执行 systemctl restart glance-backend
```

**收益**：改完代码 `git push` 就完事，不用 SSH 上服务器。

### 9.2 升级到 Docker

把整个环境（Go + 依赖 + 配置 + 二进制）打包成镜像：

```
push 代码 → CI 构建镜像 → 推到镜像仓库 → 服务器拉镜像 → docker run
```

**收益**：环境一致性（开发=测试=生产）、秒级回滚（切回旧镜像 tag）。

---

## 10. 核心概念速查（能吹两句的那种）

| 概念 | 一句话解释 |
|---|---|
| **反向代理** | Nginx 站在最前面收请求，按规则转发给后面的服务，客户端不知道后面是谁 |
| **SSL/TLS 终结** | Nginx 负责加解密 HTTPS，后端只处理明文 HTTP，职责分离 |
| **DV 证书** | 只验证域名归属的 SSL 证书，免费、签发快，适合个人/小项目 |
| **systemd unit** | Linux 服务管理单元，描述"怎么启动、谁来运行、何时重启" |
| **journal** | systemd 的日志系统，`journalctl -u <服务名>` 查看 |
| **PATH** | shell 找命令的搜索路径列表，命令找不到 99% 是 PATH 问题 |
| **环境变量** | 进程级的"配置参数"，GLANCE_CONFIG=xxx ./app 临时传 |
| **软链接** | 文件系统的"快捷方式"，sites-enabled 里都是软链接 |
| **CGO** | Go 调用 C 库的机制。CGO_ENABLED=0 生成纯静态二进制 |
| **交叉编译** | 在 A 平台编译出 B 平台的可执行文件，Go 通过 GOOS/GOARCH 一行搞定 |
| **JWT** | JSON Web Token，登录态凭证，用密钥签名防伪 |
| **NAT 回环** | 服务器通过自己的公网 IP 访问自己可能失败，不影响外部访问 |

---

## 11. 关键路径速查

| 文件/目录 | 用途 |
|---|---|
| `~/Glance-rss/` | 服务器上的代码仓库 |
| `~/Glance-rss/backend/glance-server` | 编译产物 |
| `~/Glance-rss/backend/config/config.local.yaml` | 生产配置（**不进 Git**） |
| `~/Glance-rss/backend/data/glance.db` | SQLite 数据库 |
| `/etc/systemd/system/glance-backend.service` | systemd 服务定义 |
| `/etc/nginx/sites-available/glance` | Nginx 站点配置 |
| `/etc/nginx/sites-enabled/glance` | 站点启用软链接 |
| `/etc/nginx/ssl/glancelog.cn_bundle.crt` | SSL 证书 |
| `/etc/nginx/ssl/glancelog.cn.key` | SSL 私钥（**保护它**） |

## 12. 改代码后怎么部署（当前阶段）

```bash
# 1. 本地 push 代码
git push

# 2. SSH 上服务器
ssh ubuntu@43.138.218.39

# 3. 拉最新代码 + 重新编译 + 重启服务
cd ~/Glance-rss
git pull
cd backend
go build -o glance-server .
sudo systemctl restart glance-backend

# 4. 验证
curl https://glancelog.cn/api/source/list
```

这 5 步是**手动版 CI/CD**。等做了 GitHub Actions，就把这 5 步写进 workflow 文件，让机器替你跑。

---

**部署完成时间**：2026-10-27
**版本**：v1（Git 部署）
