# 修复 scroll-view 右侧贴边显示不全问题

## 问题根因

小程序中 `scroll-view` 设置 `scroll-y` 后，如果 CSS 同时具有水平 padding，scroll-view 的可滚动区域宽度 = 容器宽度 - 左右 padding。但 scroll-view 内部子元素的宽度计算可能未正确扣除 padding，导致右侧内容被裁切/贴边显示不全。

## 修复方案

对有水平 padding 的 scroll-view，添加 `width: calc(100% - 48rpx)` 属性（48rpx = 左 padding 24rpx + 右 padding 24rpx），确保内容区域宽度正确。

## 需要修改的文件（7个）

### 1. merchant/orders/list.vue（参考页面）
- 类名：`.order-list`
- 当前 padding：`padding: 24rpx 24rpx 0`
- 修改：添加 `width: calc(100% - 48rpx)`

### 2. store/my-orders.vue
- 类名：`.order-list`
- 当前 padding：`padding: 24rpx`
- 修改：添加 `width: calc(100% - 48rpx)`

### 3. store/home.vue
- 类名：`.product-list`
- 当前 padding：`padding: 24rpx`
- 修改：添加 `width: calc(100% - 48rpx)`

### 4. sp/home.vue
- 类名：`.sp-home-container`
- 当前 padding：`padding: 24rpx`
- 修改：添加 `width: calc(100% - 48rpx)`

### 5. sp/merchants/detail.vue
- 类名：`.detail-scroll`
- 当前 padding：`padding: 24rpx`
- 修改：添加 `width: calc(100% - 48rpx)`

### 6. merchant/products/list.vue
- 类名：`.product-list`
- 当前 padding：`padding: 24rpx`
- 修改：添加 `width: calc(100% - 48rpx)`

### 7. sp/announcements/index.vue
- 类名：`.announcement-list`
- 当前 padding：`padding: 24rpx`
- 修改：添加 `width: calc(100% - 48rpx)`

## 不需要修改的文件

| 文件 | 原因 |
|------|------|
| merchant/products/edit.vue | 仅有 padding-bottom，无水平 padding |
| sp/merchants/list.vue | scroll-view 本身无 padding |
| sp/merchants/edit.vue | scroll-view 本身无 padding |
| auth/login.vue | 弹窗内协议内容，非列表 |
| merchant/settlements/history.vue | scroll-view 本身无 padding |
| sp/settlements/history.vue | scroll-view 本身无 padding |
| sp/analytics/merchant-stats.vue | v-if="false"，不渲染 |
| sp/analytics/MerchantStatsPanel.vue | scroll-view 本身无 padding |
| sp/announcements/edit.vue | 仅有 padding-bottom |

## 实施步骤

1. 逐个修改上述 7 个文件的 CSS，在对应 scroll-view 类中添加 `width: calc(100% - 48rpx)`
2. 验证前端编译通过
