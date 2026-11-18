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

请求基址集中在 `src/api/config.ts` 的 `BASE_URL`，仓库默认值为生产地址 `https://glancelog.cn`。本地联调改为 `http://localhost:8080`；真机调试改为开发机局域网 IP；提交生产版本前确认使用 HTTPS 域名。

请求层（`src/api/request.ts`）统一处理：启动时 `Taro.login` 静默换 JWT 并缓存；每个请求自动带 `Authorization: Bearer <token>`；401 自动重新登录并重试一次；业务错误统一 `showToast`（后台动作用 `silent` 关闭）。`<Image>` 无法带请求头，favicon/img-proxy 用 `?token=` 传参；文章详情还会给 img-proxy 追加原文链接作为 `referer`，配合后端防盗链重试。

## 目录职责

```text
src/
├── app.tsx / app.config.ts / app.scss  # 入口（SettingProvider）、页面与 tabBar 注册、全局样式和 4 套主题变量
├── styles/               # theme.scss（品牌色/尺寸变量）、variables.scss（mixin）
├── pages/                # 5 个页面
│   ├── index/            # Tab1 全部文章：薄壳，折叠头部 + ArticleListView
│   ├── source/           # Tab2 订阅源：列表、添加、首抓轮询、刷新、分享、重命名、复制、删除
│   ├── discover/         # 发现（非 tab）：分类折叠、订阅/取消订阅，数据由管理后台维护
│   ├── article-detail/   # 文章详情：正文、阅读设置、收藏、下一篇、分享、图片预览与分享落地
│   └── source-article/   # 单源文章列表：薄壳，取路由 id/name 后渲染 ArticleListView
├── components/           # 公共组件（仅抽离多处复用的）
├── hooks/                # useArticleList（分页/筛选/搜索/刷新）、useHeaderMetrics（胶囊位置）
├── api/                  # request 封装 + types + 各业务接口模块
├── store/                # setting.tsx：全局阅读设置 Context
└── utils/                # time、articleEvents、navContext、theme 等工具
```

tabBar 共 2 项：文章、订阅。发现页从订阅页顶部「推荐订阅源」按钮进入；阅读设置位于文章详情页底部操作栏，不再单独提供设置 Tab。

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

- **文章列表**：默认筛选「未读」，支持全部/已读/今日/收藏、标题与摘要搜索、分页加载和下拉刷新；首页头部随滚动折叠。文章 Tab 与单源文章页复用 `ArticleListView` 和 `useArticleList`
- **订阅源列表**：整行进入该源文章列表；三点菜单提供分享、重命名、复制源和删除，弹窗打开时隐藏原生 tabBar。新增后轮询首次抓取状态，下拉会触发当前用户全部源手动刷新
- **源分享**：分享路径携带 `feed_url` 与名称；接收方打开后按 feed 查重，未订阅时弹窗引导一键添加
- **发现页**：按后台顺序展示非空分类，点击分类展开推荐源；可直接订阅，已订阅项再次点击需二次确认并会级联删除对应文章。分类、推荐源与排序均在 `/admin/` 管理后台维护
- **文章详情**：进入自动标已读并同步列表；支持收藏、分享和下一篇导航，列表不足时会继续加载后续分页。正文图片限制宽度、懒加载、异步解码并支持点击预览，img-proxy 请求携带 token 与原文 referer
- **阅读设置**：详情页底部「设置」面板提供 4 套主题、3 档字号、3 档行距和左对齐/两端对齐，实时预览并持久化到后端。每源条数、刷新间隔和列表排序不在当前 UI 中开放
- **分享阅读**：通过带 `share=1&sig` 的链接进入时只读，隐藏底部操作栏；接收方未订阅同源 feed 时显示订阅引导，成功后隐藏

## 技术约定

- TypeScript 不开 strict；接口出入参和公共数据结构定义类型，局部逻辑不强标
- 样式：组件/页面强制 CSS Modules（`*.module.scss`，camelCase 类名 + `classnames`）；全局样式与主题类在 `app.scss`；主题色用 CSS 变量 `var(--card)` 等，品牌静态色用 `theme.scss` 的 SCSS 变量
- 样式单位按 750 设计稿写 `px`，Taro pxtransform 自动转换 rpx
- 路径别名 `@/` 指向 `src/`（webpack alias 与 tsconfig paths 均已配置）
- ESLint / Prettier 只保留基础规则

## 提交规范

同根目录 README：`feat / fix / docs / refactor / test / chore`，见 [../README.md](../README.md)。
