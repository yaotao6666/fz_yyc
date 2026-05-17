# Tasks
- [x] Task 1: 确认 build 与 dev 差异范围
  - [x] SubTask 1.1: 复现 `pages/store/home` 在 `build:mp-weixin` 下的异常
  - [x] SubTask 1.2: 对比 `dev` 与 `build` 产物中的模块引用、接口导出与静态资源路径
  - [x] SubTask 1.3: 归类差异属于模块加载、资源缺失还是页面初始化顺序问题

- [x] Task 2: 修复 C 端店铺页在构建态的模块与请求链路
  - [x] SubTask 2.1: 修复 `pages/store/home` 在构建态下的首页接口加载异常
  - [x] SubTask 2.2: 统一 `pages/store/product` 与 `pages/store/confirm` 的入口参数解析与公开接口调用方式
  - [x] SubTask 2.3: 确保公开接口请求不被登录或埋点链路阻塞

- [x] Task 3: 修复构建态静态资源回退路径
  - [x] SubTask 3.1: 排查店铺页使用的默认封面与默认商品图是否真实存在
  - [x] SubTask 3.2: 将缺失资源替换为可用占位资源或样式
  - [x] SubTask 3.3: 确认 `build` 产物不再出现本地图片资源加载失败

- [x] Task 4: 建立小程序构建态回归基线
  - [x] SubTask 4.1: 执行 `npm run build:mp-weixin`
  - [x] SubTask 4.2: 验证 `pages/store/home`、`pages/store/product`、`pages/store/confirm` 的构建态行为
  - [x] SubTask 4.3: 输出“以 build 产物为验收基准”的最终结论与回归结果

# Task Dependencies
- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 1]
- [Task 4] depends on [Task 2]
- [Task 4] depends on [Task 3]

## 回归结论
- 本次以 `npm run build:mp-weixin` 生成的 `dist/build/mp-weixin` 为唯一验收基准，`dev:mp-weixin` 仅作为开发预览辅助，不再替代发布验收。
- `pages/store/home` 构建产物直接依赖 `../../api/store.js`，并调用 `getStoreHome()` / `getStoreProducts()`，可证明构建态首页公开数据链路已落在独立店铺 API 模块上。
- `pages/store/product` 构建产物直接依赖 `../../api/store.js` 并调用 `getStoreProduct()`，入口参数通过 `parseStoreProductEntryOptions()` 统一解析 `merchant_id`、`product_id` 与 `scene`。
- `pages/store/confirm` 构建产物直接依赖 `../../api/store.js` 并调用 `getStoreDeliveryRules()` / `createOrder()`；源码与产物均体现先加载公开配送规则、再触发 `ensureAuth()`，满足“公开接口不被登录阻塞”的要求。
- `pages/store/home`、`pages/store/product`、`pages/store/confirm` 源码统一复用 `@utils/storeEntry`，构建产物同步保留该解析逻辑，说明 `dev` / `build` 入口参数解析路径一致。
- 默认资源 `/static/logo.png`、`/static/brand-icons/xunmeng-private-butler/icon-circle-store-data-butler.svg`、`/static/brand-icons/xunmeng-private-butler/icon-circle-aggregated-portal.svg` 已落入构建产物，构建态未再出现占位静态资源缺失。
- `npm run build:mp-weixin` 于 2026-05-17 执行成功，退出码 `0`；当前可基于该构建产物继续在微信开发者工具中做真机前回归。
