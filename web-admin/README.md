# 商家 Web/PC 后台

## 项目说明

- 项目目录：`web-admin/`
- 技术栈：`Vite + Vue 3 + TypeScript + Pinia + Vue Router + Element Plus`
- 当前角色：商家后台（单商家模式）
- 后端接口：复用 `server/` 提供的 `/api/v1/merchant/*` 接口，不新增中转层

当前后台用于承接商家登录后的全部核心功能，包括：

- 商家登录
- 工作台
- 订单管理 / 订单详情
- 商品管理
- 分类管理
- 服务人员管理
- 商家资料（含支付配置）
- 数据分析

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

- 默认构建产物目录：`web-admin/sm/`
- 默认访问前缀：`/sm`
- 生产环境部署时需要让静态资源和路由都挂载在 `/sm/` 下，例如：`https://your-domain.com/sm/login`

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

- 当前项目已配置 Vite `base=/sm/` 和 Vue Router history base `/sm/`
- 如果使用 Nginx，需要对 `/sm/` 做 SPA 路由回退，例如：

```nginx
 location /sm/ {
     alias /var/www/fz_yyc/web-admin/sm/;
     try_files $uri $uri/ /sm/index.html;
}
```

## 登录说明

- 登录页路由：`/sm/login`
- 默认首页：`/sm/dashboard`
- 登录态存储：
  - `merchant_token`
  - `merchant_info`
- 登录失效后，前端会清理本地登录态并跳回 `/sm/login`

## 目录结构

```text
web-admin/
├── src/
│   ├── api/          # 商家接口封装
│   ├── config/       # 环境变量读取
│   ├── layouts/      # 后台布局
│   ├── router/       # 路由与守卫
│   ├── stores/       # Pinia 状态
│   ├── styles/       # 全局样式
│   ├── types/        # 商家后台类型定义
│   ├── utils/        # 请求、格式化、七牛工具
│   └── views/        # 登录、工作台、订单、商品、分类、服务人员、商家资料、分析页面
└── README.md
```

## 角色边界

- 当前仓库的交付形态为：
  - `miniprogram/`：C 端用户商城
  - `staff-miniprogram/`：服务人员接单小程序
  - `web-admin/`：商家 Web/PC 后台
  - `server/`：统一后端接口
- 系统采用单商家模式，`web-admin/` 为当前唯一商家的 PC 管理后台。
