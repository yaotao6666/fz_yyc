# Tasks

- [x] Task 1: 梳理确认页金额来源与切换时机
  - [x] SubTask 1.1: 核对 `pages/store/confirm` 中商品金额、配送费、满减优惠、实付金额的本地计算逻辑
  - [x] SubTask 1.2: 核对创建订单接口返回结构，确认可直接使用的后端金额字段
  - [x] SubTask 1.3: 明确“提交前预估 / 提交后最终金额”的页面状态切换方案

- [x] Task 2: 调整确认页金额展示口径
  - [x] SubTask 2.1: 更新确认页金额区，避免将本地预估金额直接展示为最终支付金额
  - [x] SubTask 2.2: 在创建订单成功后使用后端返回的 `delivery_fee`、`discount_amount`、`pay_amount` 更新页面金额展示
  - [x] SubTask 2.3: 统一底部实付金额、支付前提示和支付埋点使用的金额口径

- [x] Task 3: 完成满减与配送费场景回归
  - [x] SubTask 3.1: 验证命中满减、未命中满减、配送免运费、普通配送费等场景的展示一致性
  - [x] SubTask 3.2: 确认创建订单后页面不再保留旧的本地预估金额

- [x] Task 4: 验证与收尾
  - [x] SubTask 4.1: 执行小程序诊断与构建验证
  - [x] SubTask 4.2: 如涉及后端结算口径调整，执行后端测试或编译验证
  - [x] SubTask 4.3: 人工验证确认页展示金额与最终实际支付金额一致

# Task Dependencies

- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 2]
- [Task 4] depends on [Task 3]
