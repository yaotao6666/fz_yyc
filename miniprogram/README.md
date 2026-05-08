# 小程序启动和停止命令

## 环境要求

- Node.js >= 16.x
- npm 或 yarn 包管理器
- 微信开发者工具（用于运行和调试微信小程序）

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
