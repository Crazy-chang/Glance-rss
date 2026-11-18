# Glance 后端

Go + Gin + GORM + SQLite 的 RSS 阅读服务，同时提供小程序 API、推荐数据管理后台 API、定时抓取和后台静态页面托管。

- RSS 解析：gofeed
- HTML 清洗：bluemonday（白名单）
- 定时抓取：robfig/cron
- SQLite 驱动：glebarez/sqlite（纯 Go，无需 CGO）
- 小程序鉴权：微信 code2session 换 openid，签发 JWT（默认 30 天）
- 后台鉴权：bcrypt 密码校验 + 独立管理员 JWT（12 小时）

## 启动

```bash
go run .
# 默认监听 :8080
curl http://localhost:8080/api/ping
# {"code":0,"msg":"ok","data":{"pong":true}}
```

构建：

```bash
go build -o glance-server .
```

启动时自动执行建表与历史数据迁移、推荐数据初始化、存量 `max_items` 归一化和默认管理员创建，无需手工 SQL。管理后台页面需先在 `../frontend-admin/` 执行 `npm run build`，然后访问 `http://localhost:8080/admin/`。

配置文件为 `config/config.yaml`，也可用环境变量 `GLANCE_CONFIG` 指定其他路径。

```yaml
server:
  port: "8080"
  base_url: "http://localhost:8080"
  refresh_interval: "1h"       # 全局定时抓取间隔：30m / 1h / 2h，非法或缺省为 1h
database:
  path: "data/glance.db"
jwt:
  secret: "glance-dev-secret-change-me"   # 生产环境必须修改
  expire_days: 30
wechat:
  appid: "YOUR_APPID"      # 填入小程序 AppID
  secret: "YOUR_SECRET"    # 填入小程序 AppSecret
fetch:
  timeout_seconds: 10      # 单源抓取超时（秒）
  concurrency: 5           # 抓取并发数
```

## 目录职责

```text
backend/
├── main.go            # 入口：配置、迁移、seed、路由、调度器与后台静态托管
├── config/            # 配置结构与加载
├── model/             # GORM 模型与建表
│   ├── user.go                  # users 小程序用户
│   ├── source.go                # sources 用户订阅源
│   ├── article.go               # articles 文章（入库时抓取清洗）
│   ├── user_article_state.go    # user_article_states 已读/收藏状态
│   ├── setting.go               # settings 用户设置
│   ├── recommend_category.go    # recommend_categories 推荐分类
│   ├── recommend_source.go      # recommend_sources 推荐源及分类关联、排序
│   ├── recommend_seed.go        # 内置 6 分类、37 个推荐源
│   ├── admin_user.go            # admin_users 管理员、admin_logs 操作日志
│   └── db.go                    # AutoMigrate、seed、旧数据回填与重置
├── middleware/
│   ├── auth.go                  # 小程序 JWT，注入 user_id
│   └── admin_auth.go            # 后台独立 JWT，注入管理员信息
├── handler/           # 小程序 HTTP 层：收参、校验、回包
│   └── admin/                   # 后台登录、分类、推荐源、系统接口
├── service/           # 业务逻辑、后台管理、操作日志与统一响应
├── fetcher/           # RSS 抓取与正文清洗（scheduler 与手动触发共用）
├── scheduler/         # cron：每分钟检查到期订阅源并触发抓取
└── web/admin/         # frontend-admin 构建产物（gitignore，部署时生成）
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

- 小程序接口除 `/api/ping` 与 `/api/auth/login` 外均需用户 JWT：优先 Header `Authorization: Bearer <token>`；`/api/favicon/:id` 与 `/api/img-proxy` 因需被 `<Image src>` 引用，额外支持 `?token=<JWT>`。失效返回 `code:401`，前端静默重新登录
- `/api/admin/*` 使用独立管理员 JWT，不能与小程序用户令牌混用，详见「管理后台接口」
- 小程序业务 DTO 的时间字段为**秒级 Unix 时间戳**；后台直接返回的 GORM 模型时间字段按 Go JSON 格式序列化
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

响应 data：

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

`today` 按 UTC+8 当天零点计算；列表项不含正文。普通筛选按发布时间倒序，`read` 按标记已读时间倒序，`favorited` 按收藏时间倒序。

#### GET /api/article/detail/:id

响应 data 为列表项全部字段，另含：

- `content`：清洗后的 HTML；`<img src>` 在响应时改写为 `/api/img-proxy?url=<编码后的原图地址>`，数据库保留原始 URL
- `share_sig`：文章分享签名
- `source_feed_url`：文章所属源的 feed 地址，用于分享后的订阅引导
- `subscribed`：当前用户是否已订阅同源 feed

普通详情只允许读取当前用户订阅下的文章，否则返回 404。分享接收方使用 `GET /api/article/detail/:id?share=1&sig=<share_sig>`，仍需登录，但验签通过后可绕过“必须已订阅”的隔离限制；此时已读/收藏固定为 `false`，`subscribed` 按接收方自己的订阅列表计算。

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

- `hot` 包含 `hot_rank > 0` 的推荐源，按 `hot_rank` 升序
- 分类按 `recommend_categories.sort_order` 升序，分类内推荐源按 `recommend_sources.sort_order` 升序；空分类不返回
- `subscribed` 按归一化 feed 地址判断：忽略协议、host 大小写、`www.` 前缀和尾斜杠，保留路径与查询参数

推荐分类、推荐源和展示顺序均通过 `/api/admin/*` 后台维护，不依赖修改代码或重启服务。

#### POST /api/discover/subscribe/:id

订阅推荐源（携带预置标题/分类/描述），异步执行首次抓取；重复订阅返回 400。

#### POST /api/discover/unsubscribe/:id

按推荐源的归一化 feed 地址找到当前用户对应订阅并删除，级联删除该源文章与状态。未订阅或推荐源不存在时返回 404。

### 设置

#### GET /api/setting/info

响应 data：

```json
{
  "theme": "white",
  "font_size": "medium",
  "line_height": "normal",
  "text_align": "left",
  "max_items": 300,
  "refresh_interval": "1h",
  "sort_order": "desc"
}
```

#### POST /api/setting/update

请求体为**部分更新**，只传要改的字段。合法取值：

| 字段 | 合法值 |
|---|---|
| theme | `white` / `black` / `yellow` / `green` |
| font_size | `small` / `medium` / `large` |
| line_height | `compact` / `normal` / `loose` |
| text_align | `left` / `justify` |
| max_items | `100` / `200` / `300` |
| refresh_interval | `30m` / `1h` / `2h` / `manual` |
| sort_order | `desc` / `asc` |

非法取值返回 400 `设置项取值不合法`；响应 data 为更新后的完整设置。

当前小程序只在文章详情页开放主题、字号、行距和对齐设置。`max_items`、`refresh_interval`、`sort_order` 仍保留在模型和接口中以兼容既有数据，但：启动时会把存量 `max_items` 统一为 300；定时抓取只读取全局 `server.refresh_interval`；文章列表排序规则固定，见文章列表说明。

### 刷新与资源

#### POST /api/refresh/start

立即返回 `{"refreshing":true}`，后台并发抓取当前用户全部订阅源，不阻塞请求。手动刷新不检查全局抓取间隔。

#### GET /api/favicon/:sourceId

返回本地缓存的源图标（带 1 天缓存头）；不存在或源不属于当前用户返回 404。

#### GET /api/img-proxy?url=<编码后的图片地址>&referer=<可选的原文地址>

正文图片代理。`url` 必填；`referer` 可选，小程序详情页传文章原文链接，用于兼容防盗链。代理抓取失败时按以下顺序重试：

1. 使用调用方传入的 `referer`
2. 使用空 Referer
3. 使用图片自身域名作为 Referer

代理**不做图片域名白名单**（RSS 图片普遍在第三方 CDN，严格限定源域名会大面积裂图），但保留 SSRF 防护：

- 仅允许 http/https；DNS 解析后拦截回环/私有/链路本地/组播地址
- 重定向最多 3 跳，每一跳重新做上述校验
- 仅返回 `image/*` 类型，单图上限 5MB，带 1 天缓存头
- 校验失败统一返回 403 `图片无法访问`

## 抓取与调度说明

- 调度器**每分钟检查一次**全部订阅源：对比 `last_fetch_at + server.refresh_interval`，到期才抓取。全局间隔合法值为 `30m` / `1h` / `2h`，缺省或非法值回退为 `1h`；不再读取用户设置中的 `refresh_interval`
- 服务启动后立即先跑一轮，之后按 cron 每分钟检查；检查误差不超过 1 分钟
- 默认抓取并发 5、单源超时 10s，均可通过 `fetch` 配置调整；失败只记录 `sources.last_error`，不影响其他源
- 文章入库流程：GUID 去重（为空时回退 link）→ bluemonday 白名单清洗 → 相对链接转绝对 → 正文超 500KB 截断 → published_at 缺失回退抓取时间 → 摘要取纯文本前 100 字 → 超出单源最大条数裁剪最旧
- 单源最大条数默认为 300，服务启动时统一归一化存量用户设置
- 正文图片地址在**响应时**重写为 `/api/img-proxy`，数据库保留原始 URL

## 管理后台

### 页面、构建与鉴权

管理后台源码位于 `../frontend-admin/`，技术栈为 Vite + React + Ant Design，使用 HashRouter。构建命令：

```bash
cd ../frontend-admin
npm ci
npm run build
```

产物输出到 `backend/web/admin/`，Go 服务通过 `/admin/` 托管；`/admin` 会 302 跳转到 `/admin/`。该目录被 `.gitignore` 排除，服务器部署时需重新构建或上传产物。后端工作目录必须是 `backend/`，否则相对静态目录 `web/admin` 无法正确定位。

首次启动自动创建管理员 `admin / 666666`，密码使用 bcrypt 存储。后台 JWT 使用 `jwt.secret + ":admin"` 作为独立密钥，claims 带 `scope="glance-admin"`，有效期 12 小时；小程序用户令牌不能访问后台，后台令牌也不能访问小程序接口。除登录外，后台接口只接受严格的 `Authorization: Bearer <token>` 请求头，不支持 query token。退出登录只记录日志，前端清除本地令牌即可。

### 推荐数据初始化与维护规则

- 当前内置 6 个分类、37 个推荐源：科技资讯、技术博客、财经商业、国内新闻、AI 人工智能、其他；游戏类源归在「其他」
- 分类表为空时才写入内置分类，推荐源表为空时才写入内置推荐源，避免后台删除的数据在重启后复活
- 历史推荐源缺少 `category_id` 时，启动自动按分类名与 seed 回填关联和排序
- 分类名称不可重复；分类下仍有推荐源时禁止删除
- 推荐源 feed 地址必须合法且全局唯一；查重忽略协议、host 大小写、`www.` 前缀与尾斜杠差异
- 新增分类或推荐源自动排到对应列表末尾；修改推荐源分类时移动到新分类末尾
- JSON 导入和恢复初始设置都会整体重建分类与推荐源，原 id 会变化；导入不接受空分类列表、非法地址或重复数据
- 登录、退出及后台写操作记录操作日志，最多保留最近 500 条

### 管理后台接口

以下接口统一使用 `{code,msg,data}` 响应。除 `/api/admin/login` 外均需管理员 JWT。创建、更新、删除、排序等写接口使用显式动词，资源 id 放在 JSON body 中。

#### 登录与退出

| 方法与路径 | 请求体 | 响应 data / 说明 |
|---|---|---|
| `POST /api/admin/login` | `{"username":"admin","password":"666666"}` | `{"token":"<JWT>","username":"admin"}`；账号或密码错误返回 400 |
| `POST /api/admin/logout` | 无 | `{"ok":true}`；记录退出日志，令牌本身无状态 |

#### 分类管理

`GET /api/admin/category/list` 返回按展示顺序排列的分类数组，每项包含 `id`、`name`、`description`、`sort_order`、`source_count`。

| 方法与路径 | 请求体 | 说明 |
|---|---|---|
| `POST /api/admin/category/create` | `{"name":"分类名","description":"可选描述"}` | 返回新分类；自动排到末尾 |
| `POST /api/admin/category/update` | `{"id":1,"name":"新名称","description":"新描述"}` | 同步更新推荐源中的分类名快照；返回 `{"ok":true}` |
| `POST /api/admin/category/delete` | `{"id":1}` | 仅允许删除空分类；返回 `{"ok":true}` |
| `POST /api/admin/category/reorder` | `{"ids":[3,1,2]}` | 按 id 数组顺序重排；返回 `{"ok":true}` |

#### 推荐源管理

`GET /api/admin/recommend/list?category_id=1` 返回推荐源数组；`category_id` 可省略或传 0 表示全部。列表项包含 `id`、`title`、`feed_url`、`site_url`、`category`、`category_id`、`description`、`hot_rank`、`sort_order` 及模型时间字段。

新增和修改共用以下字段：

| 字段 | 必填 | 说明 |
|---|---|---|
| title | 是 | 推荐源名称 |
| feed_url | 是 | http/https RSS 地址，全局查重 |
| site_url | 否 | 源站地址 |
| description | 否 | 推荐描述 |
| category_id | 是 | 所属分类 id，分类必须存在 |
| hot_rank | 否 | 大于 0 进入热门数据，数字越小越靠前；负数按 0 处理 |

| 方法与路径 | 请求体 | 说明 |
|---|---|---|
| `POST /api/admin/recommend/create` | 上述字段 | 返回新推荐源；自动排到所属分类末尾 |
| `POST /api/admin/recommend/update` | `{"id":1, ...上述字段}` | 返回 `{"ok":true}`；变更分类时排到新分类末尾 |
| `POST /api/admin/recommend/delete` | `{"id":1}` | 返回 `{"ok":true}` |
| `POST /api/admin/recommend/reorder` | `{"ids":[3,1,2]}` | 按同一分类内的 id 顺序重排；返回 `{"ok":true}` |

#### 导入与导出

`GET /api/admin/recommend/export` 的响应 data 为完整快照，分类与源顺序即当前展示顺序：

```json
{
  "categories": [
    {
      "name": "科技资讯",
      "description": "科技行业新闻与数码资讯",
      "sources": [
        {
          "title": "36氪",
          "feed_url": "https://www.36kr.com/feed",
          "site_url": "https://36kr.com",
          "description": "提供科技与商业领域的资讯、观点和深度报道",
          "hot_rank": 1
        }
      ]
    }
  ]
}
```

`POST /api/admin/recommend/import` 的请求体直接使用上述快照结构（不是 `{data: ...}` 包装）。导入在事务内整体替换现有分类和推荐源，成功返回 `{"ok":true}`；校验或写入失败不会留下部分导入结果。

#### 系统

| 方法与路径 | 参数 | 说明 |
|---|---|---|
| `POST /api/admin/system/reset` | 无请求体 | 事务内清空并恢复内置推荐分类与推荐源，返回 `{"ok":true}` |
| `GET /api/admin/system/logs?limit=100` | `limit` 可选，合法范围 1–500；缺省或超出范围按 100 | 按 id 倒序返回最新管理员登录与操作日志 |

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
