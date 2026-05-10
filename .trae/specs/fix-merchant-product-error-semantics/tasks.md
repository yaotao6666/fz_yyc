# Tasks
- [x] Task 1: 梳理商家商品详情相关错误来源并收敛返回语义
  - [x] SubTask 1.1: 检查 `product.go` 中商品详情、创建后回读、更新后回读的错误分支
  - [x] SubTask 1.2: 明确哪些错误属于“记录不存在”，哪些错误属于“内部错误”
  - [x] SubTask 1.3: 统一返回 `404` / `500` 及对应业务码

- [x] Task 2: 修正商品详情查询的关系加载稳定性
  - [x] SubTask 2.1: 校验 `models.Product` 与 `Specs` 的关联定义是否完整
  - [x] SubTask 2.2: 确保详情查询、创建后回读、更新后回读复用一致的加载逻辑
  - [x] SubTask 2.3: 避免关系加载错误再次被包装成“商品不存在”

- [x] Task 3: 验证错误语义与前端感知结果
  - [x] SubTask 3.1: 验证商品存在时详情接口返回正常结构
  - [x] SubTask 3.2: 验证商品不存在时返回 `404`
  - [x] SubTask 3.3: 验证内部异常路径返回 `500`
  - [x] SubTask 3.4: 验证商品编辑页在接口语义修正后不会再被误导性错误码干扰

- [x] Task 4: 修正前端对商品详情错误语义的透传与展示
  - [x] SubTask 4.1: 调整 `miniprogram/src/utils/request.ts`，在非 `200` 响应时优先透传后端返回的错误码与错误信息
  - [x] SubTask 4.2: 调整 `miniprogram/src/pages/merchant/products/edit.vue`，区分商品不存在与服务端异常的提示文案
  - [x] SubTask 4.3: 验证商品编辑页在 `404` 与 `500` 场景下均可展示可排查的错误信息

# Task Dependencies
- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 1]
- [Task 4] depends on [Task 3]
