# Glance

一个自建 RSS 阅读系统：微信小程序负责订阅与阅读，Web 管理后台负责推荐分类和推荐源维护，Go 服务统一提供 API、定时抓取与静态页面托管。

- **小程序端**：Taro 4 + React 18 + TypeScript
- **管理后台**：Vite 5 + React 18 + Ant Design 5
- **后端**：Go + Gin + GORM + SQLite（纯 Go 驱动，无需 CGO）
- **鉴权**：小程序使用微信 openid 静默登录（JWT 30 天）；管理后台使用独立管理员 JWT（12 小时）

## 目录结构

```text
Glance-wx/
├── frontend/        # 微信小程序端（Taro）
├── frontend-admin/  # 推荐分类与推荐源管理后台（Vite）
├── backend/         # Go API、抓取调度、SQLite 与后台静态托管
└── docs/            # 部署等专项文档
```

三个部分独立开发、独立构建。管理后台构建到 `backend/web/admin/`，由 Go 服务在 `/admin/` 路径下同进程托管，不需要单独的 Web 服务器或 Node 常驻进程。

详细说明见：

- [frontend/README.md](frontend/README.md)：小程序页面、组件与交互
- [backend/README.md](backend/README.md)：后端架构、配置与完整 API
- [docs/deploy-backend.md](docs/deploy-backend.md)：服务器构建、Nginx、SSL 与 systemd 部署

## 快速开始

### 1. 启动后端（默认 :8080）

```bash
cd backend
go run .
# 验证：curl http://localhost:8080/api/ping
```

配置文件为 `backend/config/config.yaml`，也可用 `GLANCE_CONFIG` 指定其他路径。本地模拟器联调时，微信 AppID/Secret 保持占位值会使用固定的 `dev-openid` 开发兜底；真实微信登录需要填入有效凭证。

SQLite 数据库默认自动创建在 `backend/data/glance.db`。启动时自动建表、迁移历史推荐数据并初始化默认管理员，无需手工执行 SQL。

### 2. 启动微信小程序端

```bash
cd frontend
npm install
npm run dev:weapp
```

本地联调前，将 `frontend/src/api/config.ts` 的 `BASE_URL` 改为 `http://localhost:8080`；真机调试改为电脑局域网 IP（如 `http://192.168.x.x:8080`），生产环境使用 `https://glancelog.cn`。

用微信开发者工具导入 `frontend/` 目录，编译产物在 `frontend/dist/`。本地 HTTP 联调需在「详情 → 本地设置」勾选「不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书」。

### 3. 使用管理后台

构建后台静态资源：

```bash
cd frontend-admin
npm install
npm run build
```

构建产物输出到 `backend/web/admin/`。后端启动后访问 `http://localhost:8080/admin/`，首次初始化的账号为 `admin`，密码为 `666666`。公网环境请勿沿用默认密码。

需要单独调试后台页面时，可在 `frontend-admin/` 运行 `npm run dev`，访问 `http://localhost:5174`。Vite 的 `/api` 代理目标在 `vite.config.ts` 中配置，需与实际后端端口一致；该代理只影响开发模式，不影响生产构建和线上访问。

## 提交规范

不做工具强校验，按 Conventional Commits 约定书写提交信息，前缀只用这 6 个：

| 前缀 | 用途 | 示例 |
|---|---|---|
| `feat` | 新功能 | `feat: 文章列表无限滚动` |
| `fix` | 修复缺陷 | `fix: 未读筛选下已读文章未从列表移除` |
| `docs` | 文档 | `docs: 补充后端接口文档` |
| `refactor` | 重构（不改外部行为） | `refactor: 抽取文章列表项公共组件` |
| `test` | 测试相关 | `test: 补充 HTML 清洗单测` |
| `chore` | 构建/依赖/杂项 | `chore: 初始化 Taro 项目` |

建议一次提交只做一件事，信息用一句话说清改动。
