# Tasks
- [x] Task 1: 确认商家首页 5 个图标位与资源路径。
  - [x] SubTask 1.1: 核对 `pages/merchant/home.vue` 中顶部操作区与快捷功能区的图标引用。
  - [x] SubTask 1.2: 确认实际对应的 5 个图标文件为 `store.png`、`qrcode.png`、`category.png`、`product.png`、`order.png`。

- [x] Task 2: 生成并替换正式图标资源。
  - [x] SubTask 2.1: 为“暂停营业/开始营业”生成商店营业状态语义图标。
  - [x] SubTask 2.2: 为“店铺二维码”生成二维码语义图标。
  - [x] SubTask 2.3: 为“分类管理”“商品管理”“快速核销”生成 3 个快捷功能图标。
  - [x] SubTask 2.4: 保持图标资源文件名和现有引用路径不变，直接替换对应资源。

- [x] Task 3: 验证商家首页图标显示效果。
  - [x] SubTask 3.1: 检查 `pages/merchant/home` 顶部操作区 2 个图标显示正常。
  - [x] SubTask 3.2: 检查快捷功能区 3 个图标显示正常。
  - [x] SubTask 3.3: 确认图标无破图、无拉伸、风格统一。

# Task Dependencies
- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 2]
