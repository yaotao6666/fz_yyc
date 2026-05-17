# 商家统计页面改造 + 编辑页输入框 + 列表用户数修复

## 任务一：商家统计页面改为顶部选项卡切换

### 问题
`pages/sp/analytics/merchant-stats` 页面内容过长（经营总览 + 商家转化分析 + 订单量分析 + 商家排行榜），需要滚动很长才能看完。

### 方案
将页面改为顶部选项卡切换，每个选项卡独立展示，列表类内容支持上拉加载。

**选项卡设计**：
- **总览**：经营总览卡片（商家数、访问用户数、下单用户数、下单金额）+ 订单量分析图表
- **商家分析**：商家转化分析列表（带分页上拉加载，每页6条）+ 指标切换
- **排行榜**：商家排行榜列表（带分页上拉加载）+ 指标切换

### 修改文件
- `miniprogram/src/pages/sp/analytics/MerchantStatsPanel.vue`

### 实现步骤
1. 添加顶部选项卡（总览 / 商家分析 / 排行榜），样式与现有 metric-tabs 一致
2. 将现有4个 section 拆分到3个选项卡内容中：
   - 总览 tab：overview-grid + 订单量分析
   - 商家分析 tab：商家转化分析列表（保留现有分页逻辑，改为 scrolltolower 上拉加载）
   - 排行榜 tab：商家排行榜列表（添加分页上拉加载）
3. 商家分析 tab 的 scroll-view 使用 `@scrolltolower` 触发加载更多（替代现有的点击"加载更多"按钮）
4. 排行榜 tab 添加分页支持（后端 `getTopMerchants` 已有 limit 参数，前端添加分页状态和上拉加载）
5. 选项卡切换时懒加载数据（首次切换到某 tab 时才加载）
6. 容器改为 flex 布局，scroll-view 占满剩余空间

---

## 任务二：修复编辑商家页面输入框过矮

### 问题
`pages/sp/merchants/edit` 页面输入框高度不足，输入的文字基本被挡住。

### 原因分析
`.form-input` 样式设置了 `padding: 22rpx 24rpx`，但没有设置 `height` 或 `line-height`。微信小程序的 `<input>` 组件默认高度较小，padding 撑开后整体高度不够，导致文字被裁切。

### 修改文件
- `miniprogram/src/pages/sp/merchants/edit.vue`

### 实现步骤
1. 给 `.form-input` 添加 `height: 84rpx; line-height: 40rpx;` 确保输入框有足够高度
2. 给 `.form-textarea` 添加 `line-height: 40rpx;` 确保多行文本输入也有足够行高

---

## 任务三：修复商家列表用户数显示为0

### 问题
`pages/sp/merchants/list` 页面每个商家的用户数显示为0。

### 原因分析
1. 后端 `GetMerchantList` 直接返回 `models.Merchant` 结构体，该结构体没有 `total_users` 字段
2. 后端 `GetMerchantDetail` 中 `total_users` 被硬编码为 `0`
3. `Merchant` 模型没有 `total_users`/`total_orders`/`total_amount` 字段，这些是统计值需要查询计算
4. 前端 `getMerchantList` 中做了 `Number(item?.total_users || 0)` 的兜底处理，由于后端不返回该字段，所以始终为0

### 修改文件
- `server/internal/handlers/sp/handler.go` — `GetMerchantList` 和 `GetMerchantDetail`

### 实现步骤
1. **`GetMerchantList`**：返回商家列表时，为每个商家计算 `total_users`（从 `user_visits` 表 DISTINCT user_id 计数）、`total_orders`（从 `orders` 表计数）、`total_amount`（从 `orders` 表 SUM pay_amount）
2. **`GetMerchantDetail`**：将硬编码的 `"total_users": 0` 改为实际查询 `user_visits` 表的 DISTINCT user_id 计数
3. 重启 Docker 服务使后端改动生效
