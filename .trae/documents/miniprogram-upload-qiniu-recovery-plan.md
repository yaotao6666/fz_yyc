# 小程序图片上传七牛直传修复计划

## Summary

* 目标：恢复小程序图片上传链路，确保以下两个场景都可正常上传并回显：

  * 商家端商品编辑页上传商品图片：`miniprogram/src/pages/merchant/products/edit.vue`

  * 商家详情页上传商家 Logo / 背景图：`miniprogram/src/pages/sp/merchants/detail.vue`

* 修复策略：继续沿用现有 `/api/v1/upload/token + 七牛直传` 方案，不新增本地上传兜底。

* 当前已确认：两个页面共享同一个前端上传函数 `uploadImage()`，问题不是单页逻辑，而是上传基础链路回退。

## Current State Analysis

### 1. 前端上传入口已共用同一条链路

* `miniprogram/src/pages/merchant/products/edit.vue`

  * 商品图片上传通过 `uploadImage(tempFilePath)` 执行。

  * 上传成功后把返回的 `result.url` 追加到 `formData.images`。

* `miniprogram/src/pages/sp/merchants/detail.vue`

  * 商家 Logo / 背景图上传通过 `uploadImage(filePath)` 执行。

  * 上传成功后调用 `updateSpMerchantAssets()` 保存图片地址。

* `miniprogram/src/api/index.ts`

  * `uploadImage()` 先调用 `getUploadToken()` 获取 `/api/v1/upload/token`，再 `uni.uploadFile()` 直传七牛。

  * `getUploadToken()` 会缓存后端返回的 `domain` 到 `qiniu_domain`。

* `miniprogram/src/api/qiniu_upload.ts`

  * 存在一份几乎重复的旧上传实现，与 `api/index.ts` 中的 `uploadImage()` 重叠。

  * 这是后续维护风险点，会导致“一个地方修了、另一个地方仍旧坏”。

### 2. 后端上传接口可用，但当前返回的是占位七牛配置

* `server/cmd/server/main.go`

  * `/api/v1/upload/token` 已注册在 `JWTAuth()` 下，登录态本身不是根因。

* `server/internal/handlers/upload/upload.go`

  * 会根据 `user_type` 返回不同的 `prefix`：

    * `merchant` -> `uploads/merchant/{userID}`

    * `sp` -> `uploads/sp`

  * <br />

* 实测接口结果

  * `sp` 登录后请求 `/api/v1/upload/token` 返回：

    * `domain: https://your-domain.com`

    * `token: your-access-key:...`

    * `upload_url: https://up-z0.qiniup.com`

  * `merchant` 登录后请求 `/api/v1/upload/token` 同样返回占位配置。

  * 说明上传问题是共享的配置回退，不是单个身份、单个页面或单个路径问题。

### 3. 配置来源当前明显回退到占位值

* `server/.env`

  * 当前仓库中的七牛配置是占位值：

    * `QINIU_ACCESS_KEY=your-access-key`

    * `QINIU_SECRET_KEY=your-secret-key`

    * `QINIU_BUCKET=your-bucket`

    * `QINIU_DOMAIN=https://your-domain.com`

    * `QINIU_UPLOAD_URL=https://up-z0.qiniup.com`

* `server/docker-compose.yml`

  * API 容器通过 `env_file: .env` 与环境变量注入七牛配置。

  * 如果未覆盖这些值，容器就会直接吃到当前占位配置。

* `server/pkg/qiniu/qiniu.go`

  * 仅判断“配置是否为空”，不会识别“占位值也是无效配置”。

  * 因此占位值会被当作有效配置继续生成 token，最终前端得到“看似成功、实际不可用”的上传凭证。

## Proposed Changes

### 1. 后端：把占位七牛配置识别为无效配置

#### `server/pkg/qiniu/qiniu.go`

* 增加七牛配置有效性判断，识别以下占位值为“未配置”：

  * `your-access-key`

  * `your-secret-key`

  * `your-bucket`

  * `https://your-domain.com`

* 调整 `InitQiniu()` / `GetUploadToken()` 的配置校验逻辑：

  * 不是只检查空字符串；

  * 遇到占位配置时，按“未配置”处理并返回明确错误。

* 目标：

  * 避免接口继续返回“成功但不可用”的伪 token；

  * 让问题暴露为可诊断的后端错误，而不是前端上传时才失败。

#### `server/internal/handlers/upload/upload.go`

* 保持现有 `prefix` 逻辑不变。

* 调整错误返回信息，使其能明确区分：

  * 七牛未配置

  * 获取上传凭证失败

* 不改接口路径与返回字段结构，避免影响现有前端调用。

### 2. 配置：恢复实际七牛配置来源

#### `server/.env`

* 将仓库中的占位七牛配置恢复为可用配置，或至少替换为项目当前真实联调用配置。

* 若真实密钥不能进入仓库，则执行时需要改为：

  * 保留 `.env` 模板占位；

  * 但把当前运行环境（容器/本地）实际读取的配置源改成有效值。

#### `server/docker-compose.yml`

* 复核 API 容器的七牛环境变量注入路径是否正确。

* 若执行时发现 `.env` 无法代表真实运行环境，则以容器环境覆盖为准，但不改现有上传接口协议。

### 3. 前端：统一上传实现，避免重复逻辑分叉

#### `miniprogram/src/api/index.ts`

* 保留 `uploadImage()` 作为唯一上传入口。

* 补强以下行为：

  * 获取上传凭证失败时，优先透传后端返回信息；

  * 直传七牛失败时，保留状态码和原始返回内容，便于定位是 token、域名、bucket 还是跨域问题；

  * 在生成 `key` 与拼接 `url` 时，统一使用已有 `joinQiniuFileUrl()`。

#### `miniprogram/src/api/qiniu_upload.ts`

* 停止作为独立实现继续维护。

* 方案：

  * 若当前仓库已无调用，删除该重复文件；

  * 若仍有调用，改为简单转发到 `api/index.ts` 的 `uploadImage()`。

* 目标：避免未来再次出现“同名上传函数两套逻辑不一致”的回归。

### 4. 前端页面：统一修复受影响上传入口

#### `miniprogram/src/pages/merchant/products/edit.vue`

* 保持现有图片选择和多图追加逻辑。

* 仅切换到统一上传入口后的稳定错误提示：

  * 获取上传凭证失败

  * 上传失败

* 确认上传成功后仍写入 `formData.images`，不改商品保存协议。

#### `miniprogram/src/pages/sp/merchants/detail.vue`

* 保持 `logo` / `cover_image` 两个字段共用上传逻辑。

* 使用统一上传入口后，确保：

  * 上传成功立即回显；

  * `updateSpMerchantAssets()` 保存成功后页面状态更新；

  * 上传失败时不错误清空已有图片。

### 5. 如有必要，同步前端图片 URL 归一化逻辑

#### `miniprogram/src/api/index.ts`

* 复核以下函数是否仍与修复后的上传返回一致：

  * `joinQiniuFileUrl()`

  * `normalizeImageUrl()`

* 目标：

  * 上传返回 `key` 时能正确转完整 URL；

  * 上传返回完整 URL 时不重复拼接域名；

  * 商品图、商家 Logo、商家背景图回显口径一致。

## Assumptions & Decisions

* 决策 1：继续使用现有七牛直传方案，不新增本地上传兜底接口。

* 决策 2：本次修复重点是“恢复上传基础链路 + 统一上传实现”，不改商品接口和商家资产保存接口。

* 决策 3：`sp` 与 `merchant` 上传都继续共用 `/api/v1/upload/token`，只通过后端返回的 `prefix` 隔离目录。

* 决策 4：若真实七牛密钥不能提交到仓库，执行时仍需保证运行环境中存在有效配置；否则只能把错误暴露得更清晰，无法真正恢复上传。

* 决策 5：重复上传实现 `miniprogram/src/api/qiniu_upload.ts` 不再保留为独立逻辑。

## Verification Steps

### 1. 后端接口验证

* 使用商家账号登录后请求 `/api/v1/upload/token`：

  * 返回 `code=0`

  * `prefix=uploads/sp`

  * `domain` 不再是 `https://your-domain.com`

  * `token` 不再包含 `your-access-key`

* 使用商家账号登录后请求 `/api/v1/upload/token`：

  * 返回 `code=0`

  * `prefix=uploads/merchant/{merchant_user_id}`

### 2. 七牛直传验证

* 使用真实图片文件直传七牛，确认返回 `key`。

* 前端拼接后的 `url` 可直接访问或至少能被页面 `image` 正常回显。

### 3. 页面级验证

* `pages/merchant/products/edit`

  * 选择商品图片后上传成功；

  * 页面立即展示新图；

  * 保存商品后再次进入仍能看到图片。

* `pages/sp/merchants/detail`

  * 上传商家 Logo 成功并立即回显；

  * 上传商家背景图成功并立即回显；

  * 刷新详情页后图片仍正确展示。

### 4. 代码与诊断验证

* 检查以下文件无新增诊断错误：

  * `miniprogram/src/api/index.ts`

  * `miniprogram/src/api/qiniu_upload.ts`（若保留）

  * `miniprogram/src/pages/merchant/products/edit.vue`

  * `miniprogram/src/pages/sp/merchants/detail.vue`

  * `server/internal/handlers/upload/upload.go`

  * `server/pkg/qiniu/qiniu.go`

### 5. 回归重点

* 不影响商品图片、商家 Logo、商家背景图的显示 URL 归一化。

* 不改变 `GET /api/v1/upload/token` 的前端调用方式与字段名，避免额外页面联动修改。

