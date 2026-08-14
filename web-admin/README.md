# 服务商 Web/PC 后台

## 项目说明

- 项目目录：`web-admin/`
- 技术栈：`Vite + Vue 3 + TypeScript + Pinia + Vue Router + Element Plus`
- 当前角色：仅承接服务商后台能力
- 后端接口：复用现有 `server/` 提供的 `/api/v1/sp/*` 接口，不新增中转层

当前后台用于承接原小程序服务商登录后的全部核心功能，包括：

- 服务商登录
- 工作台
- 商家列表 / 详情 / 新增 / 编辑 / 支付配置
- 公告管理
- 数据分析
- 服务商设置

## 安装依赖

```bash
cd web-admin
npm install
```

## 本地开发

```bash
cd web-admin
npm run dev
```

## 生产构建

```bash
cd web-admin
npm run build
```

- 默认构建产物目录：`web-admin/sp/`
- 默认访问前缀：`/sp`
- 生产环境部署时需要让静态资源和路由都挂载在 `/sp/` 下，例如：`https://your-domain.com/sp/login`

## 环境变量

- `.env.development`
  - `VITE_APP_TITLE`：后台标题
  - `VITE_API_BASE_URL`：接口基础地址
  - `VITE_API_PROXY_TARGET`：本地开发代理目标
- `.env.production`
  - `VITE_APP_TITLE`
  - `VITE_API_BASE_URL`

开发环境下若配置 `VITE_API_PROXY_TARGET`，Vite 会自动代理 `/api` 请求。

## 部署提示

- 当前项目已配置 Vite `base=/sp/` 和 Vue Router history base `/sp/`
- 如果使用 Nginx，需要对 `/sp/` 做 SPA 路由回退，例如：

```nginx
location /sp/ {
    alias /var/www/fz_yyc/web-admin/sp/;
    try_files $uri $uri/ /sp/index.html;
}
```

## 登录说明

- 登录页路由：`/sp/login`
- 默认首页：`/sp/dashboard`
- 登录态存储：
  - `sp_token`
  - `sp_info`
- 登录失效后，前端会清理本地登录态并跳回 `/sp/login`

## 目录结构

```text
web-admin/
├── src/
│   ├── api/          # 服务商接口封装
│   ├── config/       # 环境变量读取
│   ├── layouts/      # 后台布局
│   ├── router/       # 路由与守卫
│   ├── stores/       # Pinia 状态
│   ├── styles/       # 全局样式
│   ├── types/        # 服务商后台类型定义
│   ├── utils/        # 请求、格式化、七牛工具
│   └── views/        # 登录、工作台、商家、公告、分析、设置页面
└── README.md
```

## 角色边界

- 当前仓库的交付形态为：
  - `miniprogram/`：商家端 + C 端用户
  - `web-admin/`：服务商 Web/PC 后台
  - `server/`：统一后端接口
- 文档层面会按“用户 + 商家 + 服务商 PC 后台管理”的演进模式描述，但本轮真正落地到 PC 的只有服务商后台。
