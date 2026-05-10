# 商家商品详情错误语义修正规范

## Why
当前商家商品详情接口在 `loadProductWithRelations()` 返回内部错误时，统一对外响应 404“商品不存在”，这会掩盖真实故障并误导前端排查。对于类似 `Specs: unsupported relations for schema Product` 这类明确的服务端实现错误，接口应返回 500 而不是 404。

## What Changes
- 修正商家商品详情查询接口的错误分类逻辑，区分“商品不存在”和“服务端内部错误”
- 明确商品详情查询链路中关系预加载、DTO 转换、数据库查询失败时的响应语义
- 为商品创建/更新后回读详情的接口补充一致的错误处理策略

## Impact
- Affected specs: 商家商品管理、商家商品详情查询、统一错误响应
- Affected code: `server/internal/handlers/merchant/product.go`, `server/pkg/response/response.go`, `server/internal/models/models.go`

## ADDED Requirements
### Requirement: 商家商品详情接口必须返回准确的错误类别
系统 SHALL 在商家商品详情查询失败时，根据错误来源返回准确的业务码与 HTTP 状态，而不是统一返回“商品不存在”。

#### Scenario: 商品确实不存在
- **WHEN** 商家查询一个不属于自己、已被删除或数据库中不存在的商品
- **THEN** 接口返回资源不存在语义
- **AND** 返回的业务码为商品不存在或通用不存在类错误码
- **AND** HTTP 状态为 `404`

#### Scenario: 关系预加载或服务端实现异常
- **WHEN** 商品详情查询过程中发生关系配置错误、预加载异常、结构转换异常或其他服务端内部错误
- **THEN** 接口返回服务端内部错误语义
- **AND** 返回的业务码为服务器内部错误类错误码
- **AND** HTTP 状态为 `500`
- **AND** 不得返回“商品不存在”

#### Scenario: 创建或更新后回读详情失败
- **WHEN** 商品创建成功或更新成功，但在回读详情阶段发生服务端内部错误
- **THEN** 接口返回 `500`
- **AND** 错误信息表达为读取详情失败或等价的内部错误提示

## MODIFIED Requirements
### Requirement: 商家商品详情查询
系统 SHALL 通过 `merchant_id + product_id` 查询商品详情，并在预加载分类、规格等附加数据后返回前端可直接使用的商品结构。

#### Scenario: 详情查询成功
- **WHEN** 商品存在且当前商家有权访问
- **THEN** 接口返回商品基础信息、分类信息和规格信息
- **AND** 返回的数据结构满足前端编辑页回填要求

#### Scenario: 详情查询出现内部异常
- **WHEN** 查询逻辑命中服务端实现错误或数据库内部异常
- **THEN** 接口返回 `500`
- **AND** 不得将该错误降级为 `404`

### Requirement: 商家商品写操作后的详情回读
系统 SHALL 在商品创建或更新后执行详情回读，以返回标准化后的商品数据；若回读失败，必须暴露为内部错误。

#### Scenario: 创建后回读失败
- **WHEN** 商品创建已落库，但详情回读失败
- **THEN** 接口返回 `500`

#### Scenario: 更新后回读失败
- **WHEN** 商品更新已完成，但详情回读失败
- **THEN** 接口返回 `500`

## REMOVED Requirements
### Requirement: 详情查询失败统一按商品不存在处理
**Reason**: 该行为会把真实的服务端实现错误伪装成资源不存在，影响排障和前端错误展示。
**Migration**: 将详情查询逻辑调整为先识别“记录不存在”，其余异常统一归类为服务器内部错误；创建/更新后的回读逻辑同步采用相同策略。
