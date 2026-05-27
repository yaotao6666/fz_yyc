# Tasks

- [x] Task 1: 梳理并统一商品价格小数输入链路
  - [x] SubTask 1.1: 核对商家端商品编辑页中售价、原价、规格加价的输入控件、`v-model` 绑定与表单数据类型
  - [x] SubTask 1.2: 调整前端金额输入与校验逻辑，明确仅允许合法金额格式并支持两位小数
  - [x] SubTask 1.3: 核对商品创建、更新接口的请求结构与回显逻辑，确保小数价格保存后不丢失

- [x] Task 2: 优化用户端下单页商品预览布局
  - [x] SubTask 2.1: 调整 `pages/store/confirm` 商品列表项的左右布局，保证价格与数量区域稳定展示
  - [x] SubTask 2.2: 为商品名称、规格增加合理的换行、截断或间距策略，缓解长文本拥挤
  - [x] SubTask 2.3: 验证多商品、长名称、长规格等场景下的显示一致性

- [x] Task 3: 补齐类型与兼容性校验
  - [x] SubTask 3.1: 核对商品类型定义与接口返回结构，确保金额字段继续按小数语义处理
  - [x] SubTask 3.2: 确认整数价格、零原价、规格无加价等历史数据场景仍兼容

- [x] Task 4: 验证与回归检查
  - [x] SubTask 4.1: 执行小程序前端诊断与构建验证
  - [x] SubTask 4.2: 如涉及后端金额链路调整，执行后端测试或编译验证
  - [x] SubTask 4.3: 人工验证下单页预览展示和商家端小数价格保存回显

# Task Dependencies

- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 1]
- [Task 4] depends on [Task 2]
- [Task 4] depends on [Task 3]
