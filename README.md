# Glance

一个自建 RSS 阅读微信小程序：订阅 RSS 源、自动抓取清洗文章、多用户静默登录。

- **前端**：Taro 4 + React 18 + TypeScript（编译为微信小程序）
- **后端**：Go + Gin + GORM + SQLite
- **登录**：微信 openid 静默登录（JWT，30 天有效）

## 目录结构

```
Glance-wx/
├── frontend/   # 微信小程序前端（Taro）
└── backend/    # Go 后端服务
```

前后端完全分离、独立启动；各自的目录职责、接口文档和详细启动步骤见：

- [frontend/README.md](frontend/README.md)
- [backend/README.md](backend/README.md)

## 快速开始

### 1. 启动后端（默认 :8080）

```bash
cd backend
go run .
# 验证：curl http://localhost:8080/api/ping
```

首次运行前在 `backend/config/config.yaml` 填入微信 AppID / Secret（本地模拟器联调可先不填真实值，登录功能需要真实凭证）。

SQLite 数据库文件自动创建在 `backend/data/glance.db`。

### 2. 启动前端

```bash
cd frontend
npm install
npm run dev:weapp
```

然后用微信开发者工具打开 `frontend/` 目录（编译产物在 `frontend/dist/`），详情里勾选「不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书」，即可访问 `http://localhost:8080` 的后端。

真机调试时把前端请求基址改为电脑的局域网 IP（如 `http://192.168.x.x:8080`）。

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
