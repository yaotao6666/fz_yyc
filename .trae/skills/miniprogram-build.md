# 小程序编译技能

## 概述
财旭商贸小程序基于 uni-app + Vue 3 + TypeScript 开发，支持微信小程序和 H5 平台。

## 前置条件

### 必需环境
- Node.js >= 16.x（推荐使用 Node.js 18 LTS）
- npm >= 8.x 或 yarn >= 1.22
- 微信开发者工具（用于预览和调试小程序）

### 环境验证
```bash
# 检查 Node.js 版本
node -v

# 检查 npm 版本
npm -v
```

## 编译步骤

### 1. 安装依赖

#### 进入小程序目录
```bash
cd miniprogram
```

#### 安装依赖
```bash
npm install
```

**注意**：如果遇到 ERESOLVE 警告，这是正常的 peer dependency 警告，通常不影响安装。如果遇到 E404 错误（如 `@pinia/uni`），请检查 `package.json` 中的依赖名称是否正确。uni-app 项目只需要标准的 `pinia` 包，不需要 `@pinia/uni`。

### 2. 开发环境编译

#### 微信小程序开发模式
```bash
npm run dev:mp-weixin
```

编译成功后，输出目录为 `dist/dev/mp-weixin`，使用微信开发者工具导入此目录即可预览。

**微信开发者工具配置**：
1. 打开微信开发者工具
2. 导入项目，选择 `miniprogram/dist/dev/mp-weixin` 目录
3. AppID 使用测试号或正式 AppID
4. 勾选"不校验合法域名"（开发环境）

#### H5 开发模式
```bash
npm run dev:h5
```

编译成功后，可通过 http://localhost:3000 访问。

### 3. 生产环境编译

#### 微信小程序生产构建
```bash
npm run build:mp-weixin
```

编译成功后，输出目录为 `dist/build/mp-weixin`，使用微信开发者工具导入此目录即可上传审核。

#### H5 生产构建
```bash
npm run build:h5
```

编译成功后，可通过 `dist/build/h5` 目录部署到服务器。

### 4. 目录结构说明

```
miniprogram/
├── dist/                      # 编译输出目录
│   ├── dev/
│   │   └── mp-weixin/        # 微信小程序开发环境
│   └── build/
│       └── mp-weixin/        # 微信小程序生产环境
├── src/                       # 源代码目录
│   ├── api/                   # API 接口封装
│   ├── pages/                 # 页面文件
│   │   ├── auth/             # 认证页面（登录、注册）
│   │   ├── merchant/          # 商户管理中心
│   │   └── store/             # C端店铺页面
│   ├── stores/                # Pinia 状态管理
│   ├── types/                 # TypeScript 类型定义
│   ├── utils/                 # 工具函数
│   ├── App.vue
│   ├── main.ts
│   ├── pages.json             # 页面路由配置
│   └── manifest.json          # 应用配置
├── package.json
├── vite.config.ts             # Vite 配置
└── tsconfig.json              # TypeScript 配置
```

## 常见问题

### 1. npm install 报错 E404

**问题**：`@pinia/uni` 包找不到

**解决方案**：
```bash
# 检查 package.json，移除不存在的依赖
# 确保只有标准的 pinia 包
"pinia": "^2.1.7"
```

### 2. 编译后页面空白

**可能原因**：
- Vite 配置中的路径别名未正确设置
- TypeScript 路径解析问题

**解决方案**：
```bash
# 删除 node_modules 和 dist
rm -rf node_modules dist

# 重新安装
npm install

# 重新编译
npm run dev:mp-weixin
```

### 3. 微信开发者工具中接口请求失败

**原因**：开发环境下未配置合法域名

**解决方案**：
1. 在微信开发者工具中，勾选"详情" → "本地设置" → "不校验合法域名"
2. 或者在 `manifest.json` 中配置 h5 的 devServer 代理

### 4. 样式不生效

**可能原因**：使用了微信小程序不支持的 CSS 特性

**解决方案**：
- 避免使用 `calc()` 中的复杂表达式
- 使用 rpx 作为单位时注意计算

## 开发规范

### 页面文件结构
```
pages/
└── module/
    ├── index.vue           # 列表/首页
    ├── detail.vue          # 详情页
    └── components/         # 模块私有组件
```

### API 接口调用
使用统一封装的请求工具 `src/utils/request.ts`：
```typescript
import { request } from '@/utils/request'

// 示例
const res = await request.get('/api/v1/merchant/profile', {
  headers: {
    Authorization: `Bearer ${token}`
  }
})
```

### 状态管理
使用 Pinia 进行状态管理：
```typescript
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
```

## 相关资源

- [uni-app 官方文档](https://uniapp.dcloud.net.cn/)
- [微信小程序官方文档](https://developers.weixin.qq.com/miniprogram/dev/framework/)
- [Pinia 文档](https://pinia.vuejs.org/)
