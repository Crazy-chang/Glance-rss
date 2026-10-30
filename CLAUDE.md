# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Glance：自建 RSS 阅读微信小程序。前后端分离的两个独立应用：

- `frontend/` — Taro 4 + React 18 + TypeScript，编译为微信小程序
- `backend/` — Go + Gin + GORM + SQLite，单体服务

## 常用命令

### 后端（`backend/`）

```bash
go run .              # 启动服务，监听 :8080
go build -o glance-server .
go test ./...         # 单测（fetcher 清洗/摘要/链接、scheduler 到期判断）

# 真实抓取集成测试（需联网，默认跳过）
GLANCE_LIVE=1 go test ./fetcher/ -run TestLiveFetch -v
```

### 前端（`frontend/`）

```bash
npm install
npm run dev:weapp     # watch 编译到 dist/
npm run build:weapp   # 生产构建
npm test              # vitest，工具函数单测
```

微信开发者工具导入 `frontend/` 目录；需勾选「不校验合法域名」。请求基址在 `src/api/config.ts`，真机调试改为开发机局域网 IP。

配置文件 `backend/config/config.yaml`（可用环境变量 `GLANCE_CONFIG` 指定路径）。AppID/Secret 为占位值时登录走开发兜底（固定 `dev-openid`），无需真实凭证即可联调。

## 架构

### 后端分层

依赖方向：`handler → service → model / fetcher`，`scheduler → fetcher`，不反向依赖。

- `main.go` — 加载配置、初始化 DB、注册路由、启动调度器
- `handler/` — HTTP 层：收参、校验、回包，不写业务逻辑
- `service/` — 业务逻辑
- `model/` — GORM 模型与建表（SQLite 文件自动创建于 `backend/data/glance.db`）
- `fetcher/` — RSS 抓取与正文清洗，scheduler 与手动触发共用
- `scheduler/` — cron 每分钟检查各订阅源是否到期并触发抓取（按用户设置的刷新间隔，最小 30 分钟）

关键数据流：文章入库时即完成 GUID 去重、bluemonday 白名单清洗、相对链接转绝对、摘要提取、超出单源上限裁剪最旧。正文图片地址在**响应时**才重写为 `/api/img-proxy`，数据库始终保留原始 URL。

### 前端结构

- `src/api/request.ts` — 请求层核心：启动时 `Taro.login` 静默换 JWT；自动带 `Authorization` 头；401 自动重登录并重试一次；业务错误统一 `showToast`（后台动作用 `silent` 关闭）
- 6 个页面，其中 `article/` 与 `source-article/` 是薄壳，共用 `components/ArticleListView`（分页/筛选/搜索/下拉刷新逻辑在 `hooks/useArticleList`）
- 全局设置（主题、字号）走 `store/setting.tsx` 的 Context；详情页与列表间的已读/收藏同步走 `utils/articleEvents` 事件

### 前后端约定

- 统一响应：`{ "code": 0, "msg": "ok", "data": {} }`；非 0 时 `msg` 可直接展示给用户；分页返回 `{ list, total, page, size, hasMore }`
- 时间字段一律为**秒级 Unix 时间戳**
- 鉴权：JWT 走 `Authorization: Bearer`；`/api/favicon/:id` 与 `/api/img-proxy` 因被 `<Image src>` 引用无法带请求头，额外支持 `?token=` query
- 微信登录：`code2session` 换 openid，JWT 30 天有效

## 代码风格约定

### 后端

- 错误信息能明确定义的返回具体原因，否则统一「系统错误，请稍后再试」
- 非纯 Go SQLite 驱动（glebarez/sqlite），无需 CGO

### 前端

- 组件/页面强制 CSS Modules（`*.module.scss`，camelCase 类名 + `classnames`）；全局样式与 4 套阅读主题 CSS 变量在 `app.scss`；品牌色用 `theme.scss` 的 SCSS 变量
- 样式按 750 设计稿写 `px`，由 Taro pxtransform 转 rpx
- TypeScript 不开 strict；接口出入参和公共数据结构定义类型，局部逻辑不强标
- 路径别名 `@/` 指向 `src/`
- 不用第三方 UI 库，样式全部手写（SCSS）；只在多处复用时才抽公共组件，单页面结构直接写在页面内

## 提交规范

Conventional Commits，前缀只用：`feat / fix / docs / refactor / test / chore`，无工具强校验。提交信息用中文，一句话说清改动。

## 文档

- `backend/README.md` — 完整接口文档（含所有端点的请求/响应格式）、抓取与调度细节
- `frontend/README.md` — 页面/组件职责、关键交互说明
- 接口或行为变更时需同步更新对应 README
