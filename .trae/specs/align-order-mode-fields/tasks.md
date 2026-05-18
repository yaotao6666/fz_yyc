# Tasks

- [x] Task 1: 补齐商家表三种下单方式字段与接口口径
  - [x] SubTask 1.1: 核对并补齐 `merchants` 模型中的 `takeout_enabled`、`dine_in_enabled`、`pickup_enabled`
  - [x] SubTask 1.2: 更新商家设置接口返回与保存逻辑，确保三个字段完整读写
  - [x] SubTask 1.3: 更新数据库初始化、迁移或补字段逻辑，确保 `pickup_enabled` 默认值正确

- [x] Task 2: 调整客户端下单方式与后端字段一一对应
  - [x] SubTask 2.1: 更新前端类型定义，补齐三种下单方式字段
  - [x] SubTask 2.2: 更新商家配送设置页，确保配送、堂食、自提三个开关正确回显与保存
  - [x] SubTask 2.3: 更新确认订单页，按三个字段动态展示下单方式

- [x] Task 3: 完成下单接口校验闭环
  - [x] SubTask 3.1: 创建订单接口按 `delivery_type` 一一校验三个字段
  - [x] SubTask 3.2: 返回明确错误文案，区分配送、堂食、自提未开启场景

- [x] Task 4: 验证与文档同步
  - [x] SubTask 4.1: 执行前端诊断与构建验证
  - [x] SubTask 4.2: 执行后端测试或编译验证
  - [x] SubTask 4.3: 同步 PRD 或相关说明文档中的字段与页面行为描述

# Task Dependencies

- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 1]
- [Task 4] depends on [Task 2]
- [Task 4] depends on [Task 3]
