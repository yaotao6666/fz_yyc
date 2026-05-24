# 小程序启动和停止命令

## 环境要求

- Node.js >= 16.x
- npm 或 yarn 包管理器
- 微信开发者工具（用于运行和调试微信小程序）

## 当前角色范围

- 当前小程序仅承载商家端与 C 端用户链路。
- 原服务商登录、商家管理、分账历史、数据分析、服务商设置等能力已迁移至仓库根目录的 `web-admin/` 服务商 Web/PC 后台。
- 小程序交付验收仍以 `npm run build:mp-weixin` 为准。

## 安装依赖

```bash
# 进入小程序目录
cd miniprogram

# 安装依赖
npm install
```

## 启动命令

### 开发环境启动（微信小程序）

```bash
# 进入小程序目录
cd miniprogram

# 启动微信小程序开发模式
npm run dev:mp-weixin
```

启动后，会在 `dist/dev/mp-weixin` 目录生成微信小程序代码，使用微信开发者工具打开该目录即可预览。

### H5 开发模式

```bash
# 进入小程序目录
cd miniprogram

# 启动 H5 开发模式
npm run dev:h5
```

## 停止命令

### 停止开发服务器

- 在运行开发服务器的终端窗口，按 `Ctrl + C`（Windows/Linux）或 `Cmd + C`（macOS）停止服务器

### 清理构建文件

```bash
# 进入小程序目录
cd miniprogram

# 清理开发构建文件
rm -rf dist/dev

# 清理所有构建文件（包括生产环境）
rm -rf dist
```

## 构建命令

### 构建微信小程序（生产环境）

```bash
# 进入小程序目录
cd miniprogram

# 构建微信小程序
npm run build:mp-weixin
```

构建产物在 `dist/build/mp-weixin` 目录。

### 双小程序发布

- 当前项目支持通过一套源码切换发布 2 个微信小程序。
- 品牌差异配置位于：
  - `config/brands/xunmeng.json`
  - `config/brands/caixu.json`
- 生产环境域名配置位于：
  - `.env.production.xunmeng`
  - `.env.production.caixu`
- 预构建脚本：
  - `scripts/apply-brand-config.mjs`
- 发布命令：

```bash
# 寻梦私域管家
npm run build:mp-weixin:xunmeng

# 财旭商贸
npm run build:mp-weixin:caixu
```

- 执行上述命令时，脚本会自动：
  - 按品牌配置写入 `src/manifest.json` 中的 `name`、`appid`、`mp-weixin.appid`
  - 按品牌配置同步写入 `src/pages.json` 的全局导航标题
  - 将对应品牌的环境变量复制为 `.env.production`
- 页面运行时的品牌名称统一通过 `src/config/env.ts` 导出的 `APP_NAME` 读取，登录页、协议页等文案不再手工硬编码品牌名。
- 版本号仍以 `src/manifest.json` 中的 `versionName` / `versionCode` 为准，发布前只需维护这一处。
- `caixu` 当前默认使用占位 `appid`，正式发布前需替换为真实小程序 AppID。

### 构建 H5（生产环境）

```bash
# 进入小程序目录
cd miniprogram

# 构建 H5
npm run build:h5
```

## 常用开发流程

1. **启动开发**：在 `miniprogram` 目录执行 `npm run dev:mp-weixin`
2. **打开微信开发者工具**：导入 `dist/dev/mp-weixin` 目录
3. **修改代码**：保存后自动热更新
4. **停止开发**：在终端按 `Ctrl + C`
5. **构建生产版本**：执行 `npm run build:mp-weixin`

## 注意事项

- 确保微信开发者工具已登录且开启小程序调试模式
- 修改代码后，微信开发者工具会自动刷新
- 如果遇到构建问题，可以先清理 `node_modules` 和 `dist` 目录，然后重新安装依赖
