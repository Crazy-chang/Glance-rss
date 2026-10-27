# Glance 后端

Go + Gin + GORM + SQLite 的 RSS 阅读小程序后端。

- RSS 解析：gofeed
- HTML 清洗：bluemonday（白名单）
- 定时抓取：robfig/cron
- SQLite 驱动：glebarez/sqlite（纯 Go，无需 CGO）
- 鉴权：微信 code2session 换 openid，签发 JWT（30 天）

## 启动

```bash
go run .
# 服务监听 :8080
curl http://localhost:8080/api/ping
# {"code":0,"msg":"ok","data":{"pong":true}}
```

构建：

```bash
go build -o glance-server .
```

配置文件为 `config/config.yaml`，也可用环境变量 `GLANCE_CONFIG` 指定其他路径。

```yaml
server:
  port: "8080"
  base_url: "http://localhost:8080"
database:
  path: "data/glance.db"
jwt:
  secret: "glance-dev-secret-change-me"   # 生产环境必须修改
  expire_days: 30
wechat:
  appid: "YOUR_APPID"      # 填入小程序 AppID
  secret: "YOUR_SECRET"    # 填入小程序 AppSecret
fetch:
  timeout_seconds: 10      # 单源抓取超时
  concurrency: 5           # 抓取并发数
```

## 目录职责

```
backend/
├── main.go            # 入口：加载配置、初始化 DB、注册路由、启动调度器
├── config/            # 配置结构与加载
├── model/             # GORM 模型与建表
│   ├── user.go                  # users
│   ├── source.go                # sources 订阅源
│   ├── article.go               # articles 文章（正文入库时即抓取清洗）
│   ├── user_article_state.go    # user_article_states 已读/收藏状态
│   ├── setting.go               # settings 用户设置
│   ├── recommend_source.go      # recommend_sources 发现页内置推荐源
│   └── recommend_seed.go        # 推荐源种子数据（启动时自动写入）
├── middleware/
│   └── auth.go        # JWT 校验，注入 user_id
├── handler/           # HTTP 层：收参、校验、回包
├── service/           # 业务逻辑（列表查询、源管理、设置读写、统一响应）
├── fetcher/           # RSS 抓取与正文清洗（scheduler 与手动触发共用）
└── scheduler/         # cron：每分钟检查到期订阅源并触发抓取
```

分层依赖方向：`handler → service → model / fetcher`，`scheduler → fetcher`，不反向依赖。

## 统一响应格式

```json
{ "code": 0, "msg": "ok", "data": {} }
```

- `code = 0` 成功；非 0 为错误码，直接复用 HTTP 状态码语义（400 参数/业务错误、401 未登录或登录失效、404 资源不存在、500 系统错误）
- 能明确定义的错误返回具体原因（如「RSS地址无效或无法访问」），其余统一返回「系统错误，请稍后再试」
- 分页统一返回：`{ list, total, page, size, hasMore }`

## 接口文档

通用约定：

- 除 `/api/ping` 与 `/api/auth/login` 外，均需鉴权：优先 Header `Authorization: Bearer <token>`；`/api/favicon/:id` 与 `/api/img-proxy` 因需被 `<Image src>` 引用（无法自定义请求头），额外支持 `?token=<JWT>` query 传参。失效返回 `code:401`，前端静默重新登录
- 时间字段一律为**秒级 Unix 时间戳**
- `code:0` 成功；非 0 时 `msg` 为可直接展示给用户的原因

### 鉴权

#### POST /api/auth/login

请求：`{"code": "<wx.login 返回的 code>"}`

响应：`{"code":0,"msg":"ok","data":{"token":"<JWT>"}}`

说明：后端调微信 code2session 换 openid，自动创建用户与默认设置。当 `config.yaml` 中微信 AppID/Secret 仍为占位值时走**开发兜底**：固定使用 `dev-openid`，模拟器联调无需真实凭证。

### 文章

#### GET /api/article/list

Query：

| 参数 | 必填 | 说明 |
|---|---|---|
| filter | 否 | `all`(默认) / `unread` / `read` / `today` / `favorited` |
| search | 否 | 标题+摘要模糊搜索，与 filter 叠加 |
| source_id | 否 | 限定单个订阅源（订阅文章页用） |
| page | 否 | 从 1 开始，默认 1 |
| size | 否 | 默认 50，最大 50 |

响应 data（排序按用户设置，desc/asc 作用于发布时间）：

```json
{
  "list": [{
    "id": 12, "source_id": 1, "source_name": "阮一峰的网络日志",
    "title": "...", "summary": "纯文本前100字", "author": "阮一峰",
    "link": "https://...", "published_at": 1760000000,
    "is_read": false, "is_favorited": false
  }],
  "total": 128, "page": 1, "size": 50, "hasMore": true
}
```

`today` 按 UTC+8 当天零点计算；列表项不含正文。

#### GET /api/article/detail/:id

响应 data：列表项全部字段 + `content`（清洗后的 HTML；`<img src>` 已在响应时改写为 `/api/img-proxy?url=<编码后的原图地址>`，数据库保留原始 URL）。文章不属于当前用户返回 404。

#### POST /api/article/read/:id · /unread/:id · /favorite/:id · /unfavorite/:id

无请求体。响应：`{"id":12,"is_read":true}`（收藏类为 `is_favorited`）。操作不存在的文章返回 404。

#### POST /api/article/clear-read

清理当前用户**已读且未收藏**的文章（收藏的文章即使已读也保留），事务内删除文章记录与对应用户状态。

无请求体。响应：`{"deleted": 73}`，`deleted` 为实际删除的文章数量。

### 订阅源

#### GET /api/source/list

响应 data：`{"list":[{...}]}`，每项含：

```json
{
  "id": 1, "title": "...", "feed_url": "...", "site_url": "...",
  "favicon_url": "http://host/api/favicon/1", "category": "技术博客",
  "description": "...", "last_fetch_at": 1760000000, "last_error": "",
  "article_count": 30, "unread_count": 12
}
```

新订阅排在最前；`last_error` 非空表示最近一次抓取失败。

#### POST /api/source/add

请求：`{"feed_url":"https://.../atom.xml"}`

- 非 http(s) 地址 / RSS 无法解析 → 400 `RSS地址无效或无法访问`
- 同一 feed 重复订阅 → 400 `该订阅源已存在`
- 成功：创建订阅源并**异步立即首次抓取**，返回新建源 DTO（此时文章可能尚未入库）

#### POST /api/source/rename/:id

请求：`{"title":"新的名称"}`

空名称返回 400 `订阅源名称不能为空`；成功返回 `{"id": 1}`。

#### POST /api/source/delete/:id

事务内级联删除：该源全部文章 → 对应用户文章状态 → 源本身 → 本地 favicon 文件。不存在或不属当前用户返回 404。

### 发现

#### GET /api/discover/list

响应 data：

```json
{
  "hot": [{ "id":1, "title":"36氪", "feed_url":"...", "site_url":"...",
            "category":"科技资讯", "description":"...", "subscribed":false }],
  "categories": [{ "name":"技术博客", "sources":[ /* 同上结构 */ ] }]
}
```

`hot` 按 hot_rank 升序（最多 10 个）；`subscribed` 标记当前用户是否已订阅同 feed。

#### POST /api/discover/subscribe/:id

订阅推荐源（携带预置标题/分类/描述），同样异步首次抓取；重复订阅 400。

### 设置

#### GET /api/setting/info

响应 data：

```json
{ "theme":"white", "font_size":"medium", "max_items":100,
  "refresh_interval":"30m", "sort_order":"desc" }
```

#### POST /api/setting/update

请求体为**部分更新**，只传要改的字段。合法取值：

| 字段 | 合法值 |
|---|---|
| theme | `white` / `black` / `yellow` / `green` |
| font_size | `small` / `medium` / `large` |
| max_items | `100` / `200` / `300` |
| refresh_interval | `30m` / `1h` / `2h` / `manual` |
| sort_order | `desc`（最新在前） / `asc`（最早在前） |

非法取值返回 400 `设置项取值不合法`；响应 data 为更新后的完整设置。

### 刷新与资源

#### POST /api/refresh/start

立即返回 `{"refreshing":true}`，后台并发抓取当前用户全部订阅源（含「手动」间隔的源），不阻塞请求。

#### GET /api/favicon/:sourceId

返回本地缓存的源图标（带 1 天缓存头）；不存在或源不属于当前用户返回 404。

#### GET /api/img-proxy?url=<编码后的图片地址>

正文图片代理。**不做域名白名单**（RSS 图片普遍在第三方 CDN，如严格限定源域名会大面积裂图），但保留 SSRF 防护：

- 仅允许 http/https；DNS 解析后拦截回环/私有/链路本地/组播地址
- 重定向最多 3 跳，每一跳重新做上述校验
- 仅返回 `image/*` 类型，单图上限 5MB，带 1 天缓存头
- 校验失败统一返回 403 `图片无法访问`

## 抓取与调度说明

- 调度器**每分钟检查一次**数据库：对比每个源的 `last_fetch_at + 用户设置的刷新间隔`，到期才真正抓取（间隔为「手动」的源跳过）。最小间隔 30 分钟，检查误差不超过 1 分钟；服务启动后立即先跑一轮。
- 抓取并发 5、单源超时 10s；失败只记录 `sources.last_error`，不影响其他源。
- 文章入库流程：GUID 去重（为空时回退 link）→ bluemonday 白名单清洗 → 相对链接转绝对 → 正文超 500KB 截断 → published_at 缺失回退抓取时间 → 摘要取纯文本前 100 字 → 超出「单源最大条数」裁剪最旧。
- 正文图片地址在**响应时**重写为 `/api/img-proxy`，数据库保留原始 URL。

## 测试

少量核心逻辑单测：

- `fetcher/`：HTML 清洗、摘要提取、相对链接转绝对
- `scheduler/`：到期判断

```bash
go test ./...
```

另外有一个默认跳过的真实抓取集成测试（需联网，抓真实 RSS 验证解析/回填/favicon/去重/裁剪全链路）：

```bash
GLANCE_LIVE=1 go test ./fetcher/ -run TestLiveFetch -v
```

## 提交规范

`feat / fix / docs / refactor / test / chore`，见 [../README.md](../README.md)。
