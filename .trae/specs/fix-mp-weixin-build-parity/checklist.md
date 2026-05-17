- [x] 明确 `build:mp-weixin` 为小程序交付与验收基准
- [x] `pages/store/home` 在 `build` 模式下能正常请求店铺首页接口
- [x] `pages/store/product` 在 `build` 模式下能正常请求商品详情接口
- [x] `pages/store/confirm` 在 `build` 模式下能正常请求配送规则接口
- [x] `pages/store/home`、`pages/store/product`、`pages/store/confirm` 在 `dev` 与 `build` 下入口参数解析行为一致
- [x] 构建态不再出现缺失静态资源导致的本地图片加载失败
- [x] `npm run build:mp-weixin` 构建通过且回归结果已记录

## 验证依据
- 2026-05-17 在 `miniprogram` 目录执行 `npm run build:mp-weixin`，命令退出码为 `0`，产物输出至 `dist/build/mp-weixin`。
- `pages/store/home.js` 构建产物确认通过 `../../api/store.js` 调用 `getStoreHome()` 与 `getStoreProducts()`，未退回到 `api/index.js` 或出现接口导出缺失。
- `pages/store/product.js` 构建产物确认通过 `../../api/store.js` 调用 `getStoreProduct()`，并通过 `parseStoreProductEntryOptions()` 统一解析 `merchant_id`、`product_id` 与 `scene`。
- `pages/store/confirm.js` 构建产物确认通过 `../../api/store.js` 调用 `getStoreDeliveryRules()` 与 `createOrder()`，且页面先加载公开配送规则，再异步触发 `ensureAuth()`，不会被登录流程阻塞。
- 源码与构建产物均使用 `@utils/storeEntry` 统一解析 `merchant_id` / `product_id` / `scene`，覆盖 `pages/store/home`、`pages/store/product`、`pages/store/confirm` 三个页面。
- 默认资源 `BrandAsset.DEFAULT_PRODUCT_IMAGE=/static/logo.png`、`BrandAsset.DEFAULT_MERCHANT_LOGO=/static/brand-icons/xunmeng-private-butler/icon-circle-store-data-butler.svg` 已实际存在于 `dist/build/mp-weixin/static`，未发现占位资源缺失。
