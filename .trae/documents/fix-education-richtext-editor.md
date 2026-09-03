# 健康宣教正文富文本编辑器 + 七牛私有签名图片 改造方案

## Summary（概述）

将 web-admin「健康宣教」文章（`sm/health/education`，即 `web-admin/src/views/health/EducationArticlesView.vue`）的正文输入从纯文本框改为 **wangEditor v5** 富文本编辑器，支持粘贴/上传图片。图片上传至七牛后按「**存原始地址 + 读取时签名**」约定处理：

- 入库/落库保存**未签名的原始七牛地址**（正文 `<img>` 不带 token/e）。

- 后端在**所有对外返回正文的接口**（C端文章详情/列表、服务端列表、后台文章列表/编辑）中，对正文内七牛图片 `<img src>` 统一调用 `BuildPrivateURL` 签名后返回，保证前端可访问。

- 编辑器内实时预览新粘贴图片时，前端调用新增的签名接口拿签名地址展示。

## 为什么必须后端签名（现状分析）

- [qiniu.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/utils/qiniu.ts) 的 `uploadSpImage` 只返回拼接后的**裸地址**（`{domain}/{key}`），不含签名；七牛为私有空间，**裸地址前端无法直接访问**。

- 签名密钥只存在于后端：[qiniu.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/pkg/qiniu/qiniu.go#L200-L224) 的 `BuildPrivateURL`，前端无法自行签名。

- `BuildPrivateURL` 的 `normalizeResourceURL` 会**剥离已有 query 后重新签名**，因此即使正文里带旧 token 也能正确重签。

- C端 [education-detail.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages/store/education-detail.vue) 已支持 `rich-text` 渲染含 HTML 的正文（`isHtmlContent` 判断），无需改动即可展示编辑器输出的 HTML。

## Current State Analysis（现状）

- **编辑器**：当前项目未安装任何富文本编辑器依赖（`web-admin/package.json` 仅含 element-plus/vue/vue-router/pinia/axios）。

- **正文字段**：`HealthEducationArticle.Content`（`content` text 字段），当前用 `el-input type="textarea"` 编辑，存纯文本。

- **读写接口**（[education.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/health/education.go)）：

  - 读：`UserListEducationArticles`、`UserGetEducationArticle`（C端）、`StaffListEducationArticles`（服务端）、`MerchantListEducationArticles`（后台）。后台编辑弹窗用列表行 `row.content`。

  - 写：`MerchantCreateEducationArticle` / `MerchantUpdateEducationArticle`（`content` 为 text，直接入库，无图片签名处理）。

- **签名能力**：`pkg/qiniu.GetService().BuildPrivateURL(resource)` 可用；`pkg/qiniu` 已有 `GetService()` 取全局实例。

## Proposed Changes（改动清单）

### 后端

**1.** **`server/pkg/qiniu/qiniu.go`** **— 新增正文 HTML 图片签名/去签名助手**

- 新增正则 `imgSrcRe = regexp.MustCompile(`(?i)(src\s\*=\s\*\["'])(\[^"']+)(\["'])`)`。

- 新增方法 `func (q *QiniuService) SignContentImages(content string) string`：

  - 空串直接返回；`q == nil` 时原样返回。

  - 对正文中每个 `src` 值调用 `q.BuildPrivateURL(src)` 并回填，替换签名后的地址。`BuildPrivateURL` 对非七牛域名/空资源会原样透传，天然安全。

- 新增方法 `func (q *QiniuService) StripContentSignatures(content string) string`：

  - 对每个 `src` 值用 `url.Parse` 解析，删除 query 中 `e`、`token` 参数，输出**去除签名的原始地址**，用于入库。

**2.** **`server/internal/handlers/upload/upload.go`** **— 新增单资源签名接口**

- 新增 `SignRequest struct { URL string \`json:"url" binding:"required"\` }\`。

- 新增 `func (h *UploadHandler) Sign(c *gin.Context)`：解析请求体 → `qiniu.GetService().BuildPrivateURL(req.URL)` → `response.Success(c, gin.H{"url": signed})`（复用响应风格，与 `GetToken` 一致）。

**3.** **`server/cmd/server/main.go`** **— 注册签名路由**

- 在既有 `uploadGroup`（已挂 `JWTaut`）内新增 `uploadGroup.POST("/sign", uploadHandler.Sign)`。

**4.** **`server/internal/handlers/health/education.go`** **— 读写签名/去签名**

- 新增包内助手：`func signContent(content string) string { q := qiniu.GetService(); if q == nil { return content }; return q.SignContentImages(content) }`；`func stripSignatures(content string) string`（同理调 `StripContentSignatures`）。

- 读取时签名：

  - `MerchantListEducationArticles`：对每行 `article.Content = signContent(article.Content)`（表内正文用于后台编辑弹窗预览）。

  - `UserListEducationArticles`、`UserGetEducationArticle`、`StaffListEducationArticles`：同样对返回的 `Content` 签名。

- 写入时去签名（保证 DB 存裸地址）：

  - `MerchantCreateEducationArticle` / `MerchantUpdateEducationArticle`：对提交的 `req.Content` 先 `stripSignatures` 再入库。

### 前端（web-admin）

**5.** **`web-admin`** **安装 wangEditor v5**

- 新增依赖：`@wangeditor/editor`、`@wangeditor/editor-for-vue`（`npm ls` 默认最新 v5.x，与 Vue 3.5/Vite 8 ESM 兼容）。运行 `npm install`。

**6.** **`web-admin/src/api/sp.ts`** **— 新增签名 API**

- 新增 `export async function signMediaUrl(url: string): Promise<string>`：`POST /api/v1/upload/sign`，body `{ url }`，返回 `data.url`。参照现有 `unwrapApiResponse` 风格。

**7.** **`web-admin/src/components/WangEditor.vue`（新增，小封装）**

- `props: { modelValue: string }`，`emits: ['update:modelValue']`。

- 引入 wangEditor v5 `Editor`/`Toolbar`/`IDomEditor`。

- `editorConfig`：`MENU_CONF['uploadImage'].customUpload = async (file, insertFn) => { const { url } = await uploadSpImage(file); const signed = await signMediaUrl(url); insertFn(signed, '', '') }`。

  - 说明：wangEditor v5 图片菜单的 `customUpload` 同时覆盖「工具栏插入」与「直接粘贴图片」，保证二者都走「上传→签名→插入」。

- `onMounted` 用 `setHtml(modelValue)` 初始化；`watch(modelValue)` 外部重置时同步；`onBeforeUnmount` 调 `editor.destroy()`。

- 变更时 `emit('update:modelValue', editor.getHtml())`。

- 引入 wangEditor CSS（`@wangeditor/editor/dist/css/style.css`）。

**8.** **`web-admin/src/views/health/EducationArticlesView.vue`** **— 正文改用组件**

- 引入并注册 `WangEditor`。

- 将 `<el-form-item label="正文">` 内的 `<el-input type="textarea">` 替换为 `<WangEditor v-model="editForm.content" style="..."/>`。

- `handleSave` 的正文校验优化为：`content` 去除 HTML 标签后非空（避免只含 `<p><br>` 空内容通过）；其余（标题/分类/发布校验）不变。

- 由于后台列表接口已返回签名后的 `content`（见后端改动 4），打开编辑时正文图片能正常预览；保存时后端再去签名入库，双向闭环。

## Assumptions & Decisions（假设与决策）

- **编辑器**：采用 **wangEditor v5**（`@wangeditor/editor` + `@wangeditor/editor-for-vue`），Vue3 兼容、图片粘贴/上传通过 `customUpload` 统一处理。

- **签名策略**：采用「**存原始地址 + 读取时签名**」——DB 存未签名地址，所有读取接口对正文图片签名，DB 干净且符合项目「仅上传后 url 入库」的既有约定。

- **C端展示**：`education-detail.vue` 已用 `rich-text`，预期可渲染编辑器 HTML；若真机 `rich-text` 对 `<img>` 尺寸/显示有问题，属于独立跟进项，不在本次改动范围（见验证步骤）。

- **单资源签名接口鉴权**：挂在已有 JWT 鉴权的 `/upload` 组，web-admin 登录态可调用。

## Verification（验证步骤）

1. 后端：`cd server && go build ./...` 通过；`docker compose up -d --build api` 重建生效。
2. 前端：`cd web-admin && npm install && npm run build` 通过。
3. 功能验证（web-admin）：

   - 新增文章，工具栏/粘贴插入图片 → 编辑器内图片正常显示（已签名）。

   - 保存后查库：`health_education_articles.content` 中 `<img src>` 为**无 token/e 的原始地址**。

   - 再次打开编辑该文章 → 图片仍正常显示（读取接口已签名）。

   - 列表/详情接口返回的 `content` 中 `<img>` 带 `e=` 与 `token=` 签名参数。
4. C端验证：发布文章后，C端「健康宣教」文章详情能显示正文与粘贴的图片（`rich-text` 渲染）；若图片显示异常，记录为独立跟进，不改本次方案。

