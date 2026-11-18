# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Glance：自建 RSS 阅读系统，由三个独立开发、构建的应用组成：

- `frontend/` — Taro 4 + React 18 + TypeScript，编译为微信小程序
- `frontend-admin/` — Vite 5 + React 18 + Ant Design 5，推荐分类与推荐源管理后台
- `backend/` — Go + Gin + GORM + SQLite，单体 API、抓取调度与后台静态托管服务

## 常用命令

### 后端（`backend/`）

```bash
go run .              # 启动服务，默认监听 :8080
go build -o glance-server .
go test ./...         # 单测（fetcher 清洗/摘要/链接、scheduler 到期判断）

# 真实抓取集成测试（需联网，默认跳过）
GLANCE_LIVE=1 go test ./fetcher/ -run TestLiveFetch -v
```

### 微信小程序端（`frontend/`）

```bash
npm install
npm run dev:weapp     # watch 编译到 dist/
npm run build:weapp   # 生产构建
npm test              # vitest，工具函数单测
```

微信开发者工具导入 `frontend/` 目录；本地 HTTP 联调需勾选「不校验合法域名」。请求基址在 `src/api/config.ts`：本地改为开发机地址，真机改为局域网 IP，生产使用 HTTPS 域名。

### 管理后台（`frontend-admin/`）

```bash
npm install
npm run dev           # Vite 开发服务，默认 :5174
npm run build         # 类型检查并构建到 backend/web/admin/
npm run preview       # 本地预览构建产物
```

后台通过 `/admin/` 访问，使用 HashRouter，无服务端 history 回退依赖。Vite `/api` 代理只用于开发模式，目标端口需与本地后端一致；生产环境由 Go 同域托管，不读取该代理配置。

后端配置文件为 `backend/config/config.yaml`（可用 `GLANCE_CONFIG` 指定路径）。AppID/Secret 为占位值时登录走开发兜底（固定 `dev-openid`），无需真实凭证即可联调。

## 架构

### 后端分层

依赖方向：`handler → service → model / fetcher`，`scheduler → fetcher`，不反向依赖。

- `main.go` — 加载配置、自动迁移与 seed、初始化管理员、注册路由、启动调度器并托管 `/admin/`
- `handler/` — 小程序 HTTP 层：收参、校验、回包，不写业务逻辑
- `handler/admin/` — 管理后台独立 HTTP 路由，写操作使用显式动词，资源 id 放 JSON body
- `service/` — 小程序业务、后台鉴权、推荐数据管理、操作日志与统一响应
- `model/` — GORM 模型、建表、推荐数据初始化与历史数据回填
- `middleware/auth.go` — 小程序用户 JWT；图片资源额外允许 `?token=`
- `middleware/admin_auth.go` — 后台 JWT，仅接受严格的 `Authorization: Bearer <token>`
- `fetcher/` — RSS 抓取与正文清洗，scheduler 与手动触发共用
- `scheduler/` — 每分钟检查全部订阅源，按全局 `server.refresh_interval` 判断是否到期
- `web/admin/` — 管理后台构建产物，被 `.gitignore` 排除，部署时必须重新构建或上传

关键数据流：文章入库时完成 GUID 去重、bluemonday 白名单清洗、相对链接转绝对、摘要提取、超出单源上限裁剪最旧。正文图片地址在**响应时**重写为 `/api/img-proxy`，数据库始终保留原始 URL。小程序详情页追加原文链接为 `referer`；代理依次尝试调用方 Referer、空 Referer、图片自身域名 Referer，并保留 SSRF、重定向、类型与大小校验。

### 管理后台与推荐数据

- 后台管理员与小程序用户完全隔离：密码使用 bcrypt；后台令牌密钥为 `jwt.secret + ":admin"`，claims 带 `scope="glance-admin"`，有效期 12 小时，两类令牌不能混用
- 首次启动创建 `admin / 666666`；生产环境不应沿用默认密码
- `RecommendCategory` 是独立分类表；`RecommendSource` 通过 `category_id` 关联，并保留 `category` 名称快照。分类与分类内推荐源均按 `sort_order` 升序展示
- 当前内置 6 个分类、37 个推荐源，游戏类源归在「其他」
- 分类表或推荐源表为空时才写入对应 seed，避免后台删除的数据在重启后复活；旧推荐源启动时自动回填 `category_id`
- 分类下仍有推荐源时禁止删除；JSON 导入和“恢复初始设置”都会整体重建分类与推荐源，原 id 会变化
- 后台写操作与登录/退出记录操作日志，最多保留最近 500 条

### 微信小程序端结构

- `src/api/request.ts` — 启动时 `Taro.login` 静默换 JWT；自动带 `Authorization`；401 自动重登录并重试一次；业务错误统一 `showToast`（后台动作用 `silent` 关闭）
- 当前 5 个页面、2 个 tab（文章、订阅）；`index/` 与 `source-article/` 是薄壳，共用 `components/ArticleListView`，分页/筛选/搜索/下拉刷新逻辑在 `hooks/useArticleList`
- 阅读设置位于文章详情页底部面板，包含主题、字号、行距与对齐；全局状态走 `store/setting.tsx` Context
- 详情页支持收藏、分享、下一篇、图片预览和分享落地后的同源订阅引导；订阅源也支持分享与一键添加
- 详情页与列表间的已读/收藏同步走 `utils/articleEvents` 事件

### 前后端约定

- 统一响应：`{ "code": 0, "msg": "ok", "data": {} }`；非 0 时 `msg` 可直接展示给用户；分页返回 `{ list, total, page, size, hasMore }`
- 小程序业务 DTO 的时间字段使用**秒级 Unix 时间戳**；后台直接返回的 GORM 模型时间字段按 Go JSON 格式序列化
- 小程序 JWT 30 天有效；后台 JWT 12 小时有效且独立验签
- `/api/favicon/:id` 与 `/api/img-proxy` 因被 `<Image src>` 引用无法带请求头，额外支持 `?token=`；后台接口不使用 query token

## 代码风格约定

### 后端

- 错误信息能明确定义的返回具体原因，否则统一「系统错误，请稍后再试」
- 非纯 Go SQLite 驱动（glebarez/sqlite），无需 CGO

### 微信小程序端

- 组件/页面强制 CSS Modules（`*.module.scss`，camelCase 类名 + `classnames`）；全局样式与 4 套阅读主题 CSS 变量在 `app.scss`；品牌色用 `theme.scss` 的 SCSS 变量
- 样式按 750 设计稿写 `px`，由 Taro pxtransform 转 rpx
- TypeScript 不开 strict；接口出入参和公共数据结构定义类型，局部逻辑不强标
- 路径别名 `@/` 指向 `src/`
- 小程序端不用第三方 UI 库，样式全部手写（SCSS）；只在多处复用时才抽公共组件，单页面结构直接写在页面内

### 管理后台

- 使用 Ant Design 组件与 HashRouter；构建 `base` 固定为 `/admin/`
- 后台令牌保存在 `glance_admin_token`，401 时清除令牌并跳转登录页
- 编辑弹窗通过 `Form initialValues + key` 回填，避免表单未挂载时调用 `setFieldsValue`

## 提交规范

Conventional Commits，前缀只用：`feat / fix / docs / refactor / test / chore`，无工具强校验。提交信息用中文，一句话说清改动。

## 文档

- `backend/README.md` — 完整小程序与后台接口文档、推荐数据规则、抓取与调度细节
- `frontend/README.md` — 小程序页面/组件职责、关键交互说明
- `docs/deploy-backend.md` — 管理后台与 Go 服务构建、systemd、Nginx 和验证流程
- 接口或行为变更时需同步更新对应 README；部署链路变更时需同步更新部署文档
