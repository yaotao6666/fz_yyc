# 双小程序最小配置切换方案

## Summary

* 目标：在 `miniprogram` 保持一套源码的前提下，支持频繁发布 2 个微信小程序，并且只通过极少量配置切换以下内容：

  * 小程序 `appid`

  * 应用名称

  * `manifest.json` 中的 `versionName` / `versionCode`

  * `VITE_API_BASE_URL` / `VITE_WS_BASE_URL`

* 偏好已确认：

  * 使用“一个品牌配置文件”方案

  * 版本号仍以 `miniprogram/src/manifest.json` 的版本字段为源头

  * 两个小程序之间除了 `appid/名称/版本` 外，还需要切换接口域名

* Git 侧当前初步结论：

  * 根目录 `.gitignore` 没有忽略 `miniprogram/src/manifest.json` 或 `miniprogram/package.json`

  * `.git/info/exclude` 也没有忽略这些文件

  * 因此“看不到变更”不是仓库显式忽略规则导致，后续执行阶段应进一步验证这两个文件是否已被 Git 跟踪，以及用户实际查看的是 IDE 面板还是命令行结果

## Current State Analysis

* `miniprogram/src/manifest.json`

  * 当前只有一套静态配置：

    * 顶层 `appid` 写死为 `wx71a54b66552dc9e8`

    * `mp-weixin.appid` 同样写死

    * `name` / `versionName` / `versionCode` 也写死

  * 这意味着发布第二个小程序时，必须手工改回改去，容易污染版本控制。

* `miniprogram/package.json`

  * 当前只有一套标准脚本：

    * `dev:mp-weixin`

    * `build:mp-weixin`

  * 没有按品牌/小程序区分的构建入口，也没有任何预构建脚本。

* `miniprogram/vite.config.ts`

  * 当前只使用 `loadEnv(mode, __dirname, '')` 读取环境变量。

  * Vite 已经具备切换 `.env` 文件的基础能力，但这套能力目前只作用于运行时 HTTP/WS 配置，不会自动写入 `manifest.json`。

* `miniprogram/src/config/env.ts`

  * 当前已经统一收口了 `VITE_API_BASE_URL` / `VITE_WS_BASE_URL`。

  * 说明“接口域名切换”本身已经具备条件，缺的是“按小程序品牌自动切不同 `.env` 文件”的入口约定。

* `miniprogram/.env.development` / `miniprogram/.env.production`

  * 当前 `.env.production` 里已经存在两套域名注释痕迹，说明项目确实已有多品牌/多小程序切换诉求，但仍停留在手工注释切换阶段。

* Git 忽略规则

  * `/Users/yaotao/Documents/trae_projects/fz_yyc/.gitignore` 未忽略 `manifest.json`、`package.json`、`package-lock.json`

  * `/Users/yaotao/Documents/trae_projects/fz_yyc/.git/info/exclude` 仅忽略 `.DS_Store`

  * 因此从仓库规则上看，这些文件应该可以被检测到修改

## Proposed Changes

### 1. 新增一份“品牌配置层”，只承载双小程序差异

* 新增目录建议：

  * `miniprogram/config/brands/`

* 新增两个极小配置文件，例如：

  * `miniprogram/config/brands/xunmeng.ts`

  * `miniprogram/config/brands/caixu.ts`

* 每个品牌配置只包含最小差异字段：

  * `brandKey`

  * `appName`

  * `appid`

  * `envFile`

  * `versionName`

  * `versionCode`

* 这样可以把“品牌差异”集中收口到 2 个很小的文件中，避免复制整份 `manifest.json`。

### 2. 将 `manifest.json` 改为“模板文件 + 预构建注入”

* 保留一个基准模板，建议方式：

  * 将现有 `miniprogram/src/manifest.json` 作为模板源保留在仓库中

  * 增加明确占位或由预构建脚本按 JSON 字段覆盖

* 新增预构建脚本，例如：

  * `miniprogram/scripts/apply-brand-config.mjs`

* 脚本职责：

  * 读取品牌配置文件

  * 读取 `miniprogram/src/manifest.json`

  * 仅覆盖这些字段：

    * `name`

    * `appid`

    * `versionName`

    * `versionCode`

    * `mp-weixin.appid`

  * 不碰其余业务配置，避免误改 `permission`、`setting`、`requiredPrivateInfos`

* 这样实现“最小变更、最低维护成本”，也满足“版本号以 manifest 为准”的偏好：

  * 模板中的默认版本可保留一份

  * 实际品牌版本则以品牌配置覆盖写入

### 3. 为每个小程序补一份独立环境变量文件

* 现有 `env.ts` 已可读 `VITE_API_BASE_URL` / `VITE_WS_BASE_URL`，因此无需改业务代码。

* 只需新增环境文件，例如：

  * `miniprogram/.env.production.xunmeng`

  * `miniprogram/.env.production.caixu`

* 每个文件仅维护该小程序对应的：

  * `VITE_API_BASE_URL`

  * `VITE_WS_BASE_URL`

  * 如有必要再加 `VITE_DEV_PROXY_TARGET`

* 同时新增一个极薄的切换步骤：

  * 预构建脚本在执行前，根据品牌配置选择对应 env 文件

  * 将其复制/同步为构建实际读取的 `.env.production`

  * 或通过脚本参数/Node 逻辑在构建前写入临时 env 文件

* 由于当前 Vite 并未按“品牌”分层加载 env，所以执行方案应明确采用“构建前生成标准 `.env.production`”而不是改动所有业务代码。

### 4. 在 `package.json` 增加按品牌构建脚本

* 修改文件：

  * `miniprogram/package.json`

* 新增脚本建议：

  * `prepare:brand:xunmeng`

  * `prepare:brand:caixu`

  * `build:mp-weixin:xunmeng`

  * `build:mp-weixin:caixu`

* 脚本调用顺序建议：

  * `prepare:brand:*`：执行品牌配置注入脚本，写入 `manifest.json` 与 `.env.production`

  * `build:mp-weixin:*`：在 `prepare` 后执行 `uni build -p mp-weixin`

* 这样发布时只需：

  * `npm run build:mp-weixin:xunmeng`

  * `npm run build:mp-weixin:caixu`

* 不需要手工改 `manifest.json`、注释/反注释 `.env.production`

### 5. 保持 `package.json` 本身尽量不参与品牌差异

* 当前你问到 `package` 文件是否纳入控制。

* 从发布双小程序的目标看，`miniprogram/package.json` 更适合只承载：

  * 通用依赖

  * 通用构建脚本

* 不建议把品牌差异直接塞进 `package.json` 内容本身。

* 这样一来：

  * `package.json` 应长期稳定、纳入版本控制

  * 真正频繁变动的差异放到 `config/brands/*.ts` 与 `env.production.*`

### 6. Git 跟踪问题的执行期核查方案

* 当前只读排查已确认：

  * 不是 `.gitignore`

  * 不是 `.git/info/exclude`

* 执行阶段应补充以下只读验证：

  * `git ls-files --error-unmatch miniprogram/src/manifest.json`

  * `git ls-files --error-unmatch miniprogram/package.json`

  * `git status -- miniprogram/src/manifest.json miniprogram/package.json`

* 若结果显示文件已跟踪但 IDE 不显示：

  * 进一步排查 IDE 源码管理过滤设置

  * 排查是否打开的是错误工作区根目录

* 若结果显示文件未跟踪：

  * 则需要显式 `git add` 纳入版本控制

* 但从当前仓库规则看，更大概率是“文件应可跟踪，只是尚未实际验证”

### 7. 文档同步

* 根据项目规则，改动产品功能/工程配置需同步文档。

* 建议同步更新：

  * `miniprogram/README.md`

  * `docs/prd/PRD-功能说明.md`

  * `PRD.md`

* 文档需要明确：

  * 双小程序发布入口

  * 品牌配置文件位置

  * 版本号维护规则

  * 构建命令示例

## Assumptions & Decisions

* 决策：采用“一套源码 + 两个极小品牌配置文件 + 一个预构建脚本”的方案。

* 决策：小程序间差异仅覆盖：

  * `appid`

  * 应用名称

  * `versionName`

  * `versionCode`

  * `VITE_API_BASE_URL`

  * `VITE_WS_BASE_URL`

* 决策：继续以 `manifest.json` 的版本字段语义为发布版本源，但实际发布时由品牌配置覆盖写入。

* 决策：不维护两整份 `manifest.json`，避免重复和手工同步。

* 决策：不把品牌差异直接写入业务代码或散落在多个页面/常量文件中。

* 判断：`manifest.json` / `package.json` 当前从忽略规则上看应纳入 Git 控制；“检测不到变动”暂不归因为忽略规则。

## Verification Steps

* 配置切换验证

  * 执行品牌准备脚本后，检查 `miniprogram/src/manifest.json` 中：

    * `name`

    * `appid`

    * `mp-weixin.appid`

    * `versionName`

    * `versionCode`
      是否切换为目标小程序配置

  * 检查构建前实际生效的 `.env.production` 是否切到目标域名

* 构建验证

  * 在 `miniprogram` 目录分别执行：

    * `npm run build:mp-weixin:xunmeng`

    * `npm run build:mp-weixin:caixu`

  * 确认两次构建均成功，且 `dist/build/mp-weixin` 中产物对应目标品牌配置

* Git 验证

  * 执行：

    * `git ls-files --error-unmatch miniprogram/src/manifest.json`

    * `git ls-files --error-unmatch miniprogram/package.json`

    * `git status -- miniprogram/src/manifest.json miniprogram/package.json`

  * 确认两文件是否已纳入版本控制，以及修改后能否正常显示差异

* 文档验证

  * 检查 README / PRD 中的双小程序发布说明与实际脚本命名一致

  * 确保没有继续保留“手工注释切域名、手工改 manifest”的旧说明

