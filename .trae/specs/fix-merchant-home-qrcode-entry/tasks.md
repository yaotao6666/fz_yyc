# Tasks
- [x] Task 1: 核对商家二维码当前生成链路与店铺首页解析规则
  - [x] SubTask 1.1: 确认 `pages/merchant/home` 当前二维码加载与展示方式未偏离现有商家主页交互
  - [x] SubTask 1.2: 核对 `/api/v1/merchant/qrcode` 当前返回的 `scene`、`page` 与二维码内容生成逻辑
  - [x] SubTask 1.3: 对照 `miniprogram/src/utils/storeEntry.ts` 确认当前被支持的 `scene` / `merchant_id` 参数格式

- [x] Task 2: 修复二维码生成接口的页面路径与商家 ID 参数格式
  - [x] SubTask 2.1: 将二维码生成逻辑切换到当前实际的店铺首页路径
  - [x] SubTask 2.2: 将二维码参数格式调整为可被 `pages/store/home` 现有解析逻辑识别的商家 ID
  - [x] SubTask 2.3: 保留接口调试返回字段，确保返回中的 `scene` 和 `page` 与生成配置一致

- [x] Task 3: 验证商家主页二维码在已发布小程序链路中的可用性
  - [x] SubTask 3.1: 校验商家主页仍能正常加载二维码图片
  - [x] SubTask 3.2: 校验二维码接口返回的 `scene` / `page` 已对齐 `pages/store/home` 的入口要求
  - [x] SubTask 3.3: 以 `npm run build:mp-weixin` 为验收基准，确认相关改动不影响小程序构建

# Task Dependencies
- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 2]
