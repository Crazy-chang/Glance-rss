# Glance 前端（微信小程序）

Taro 4 + React 18 + TypeScript，编译目标为微信小程序。不用第三方 UI 库，样式全部手写（SCSS），紧凑布局。

## 环境要求

- Node.js ≥ 18（开发使用 Node 20）
- 微信开发者工具（稳定版）

## 启动

```bash
npm install
npm run dev:weapp     # watch 模式编译到 dist/
```

微信开发者工具中「导入项目」，项目目录选择本 `frontend/` 文件夹（工具读取 `project.config.json`，小程序根目录指向 `dist/`）。

- AppID 可先选「测试号」（project.config.json 中为 `touristappid`）
- 「详情 → 本地设置」勾选「不校验合法域名…」，否则无法请求 `http://localhost:8080`

打包生产版本：

```bash
npm run build:weapp
npm test           # vitest 跑工具函数单测
```

## 请求基址

开发环境默认请求 `http://localhost:8080`（见 `src/api/config.ts`）。真机调试改为开发机局域网 IP。

请求层（`src/api/request.ts`）统一处理：启动时 `Taro.login` 静默换 JWT 并缓存；每个请求自动带 `Authorization: Bearer <token>`；401 自动重新登录并重试一次；业务错误统一 `showToast`（后台动作用 `silent` 关闭）。`<Image>` 无法带请求头，favicon/img-proxy 用 `?token=` 传参（后端支持）。

## 目录职责

```
src/
├── app.tsx / app.config.ts / app.scss  # 入口（SettingProvider 包裹）、页面注册与 tabBar、全局样式与 4 套阅读主题 CSS 变量
├── styles/               # theme.scss（品牌色/尺寸变量）、variables.scss（mixin：省略号、按钮重置等）
├── pages/                # 6 个页面
│   ├── article/          # Tab1 文章列表：薄壳，渲染 ArticleListView
│   ├── source/           # Tab2 订阅源管理：列表、添加、重命名、删除、三点菜单
│   ├── setting/          # Tab3 设置：阅读主题/字号/每源条数/刷新间隔/排序/清理缓存
│   ├── discover/         # 发现（非 tab，入口在订阅页）：热门订阅横滑卡片 + 分类
│   ├── article-detail/   # 文章详情：正文渲染、已读/收藏、上一篇/下一篇、分享
│   └── source-article/   # 单源文章列表：薄壳，取路由 id/name 后渲染 ArticleListView
├── components/           # 公共组件（仅抽离多处复用的）
├── hooks/                # useArticleList（分页/筛选/搜索/下拉刷新）、useHeaderMetrics（胶囊位置）
├── api/                  # request 封装 + types + 各业务接口模块
├── store/                # setting.tsx：全局设置 Context（主题、字号等）
└── utils/                # time 相对时间、articleEvents 状态事件、navContext 详情本地序列、theme 主题静态色板
```

tabBar 共 3 项：文章、订阅、设置。发现页从订阅页顶部「推荐订阅源」按钮进入。

### 公共组件

| 组件 | 复用位置 |
|---|---|
| ArticleListView | 文章 Tab、订阅文章页的完整列表（搜索/筛选/分页/下拉刷新） |
| ArticleListItem | 列表项：未读小红点 + 标题两行 + 摘要两行 + 来源·相对时间 |
| FilterBar | 全部/未读/已读/今日/收藏 筛选（默认选中「未读」） |
| EmptyState | 列表空态（可带行动按钮） |
| LoadingMore | 列表底部「加载中」指示器 |
| Spinner | CSS 旋转指示器（首次加载等场景） |

只在单页面使用一次的结构不抽组件，直接写在页面内（如订阅页的添加弹窗、重命名弹窗、操作菜单）。

## 关键交互

- **订阅源列表**：整行可点击进入该源文章列表（标题后紧跟 chevron 提示）；右侧三点按钮弹出底部半屏菜单（分享/重命名/复制源/删除），菜单打开时隐藏原生 tabBar。
- **设置页**：阅读主题与字号通过 action sheet 选择；每源最大条数、刷新间隔、排序同理；清理缓存二次确认后调用 `/article/clear-read`。
- **文章详情**：进入自动标已读；正文图片懒加载（`loading="lazy"`）；支持上一篇/下一篇导航。

## 技术约定

- TypeScript 不开 strict；接口出入参和公共数据结构定义类型，局部逻辑不强标
- 样式：组件/页面强制 CSS Modules（`*.module.scss`，camelCase 类名 + `classnames`）；全局样式与主题类在 `app.scss`；主题色用 CSS 变量 `var(--card)` 等，品牌静态色用 `theme.scss` 的 SCSS 变量
- 样式单位按 750 设计稿写 `px`，Taro pxtransform 自动转换 rpx
- 路径别名 `@/` 指向 `src/`（webpack alias 与 tsconfig paths 均已配置）
- ESLint / Prettier 只保留基础规则
- 版本号取 `package.json` 的 `version`，在「设置」Tab 底部展示

## 提交规范

同根目录 README：`feat / fix / docs / refactor / test / chore`，见 [../README.md](../README.md)。
