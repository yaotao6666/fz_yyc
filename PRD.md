# 寻梦私域管家 API 产品需求文档 (PRD)

## 1. 项目概述

### 1.1 项目背景
开发一套基于微信支付服务商模式的商家管理系统后端 API，支持服务商统一管理多个商家，为每个商家提供商品管理、订单管理、数据分析等完整的经营工具。

### 1.2 小程序产品定位
**产品名称**：寻梦私域管家

**产品定位**：专为中小商家打造的统一私域经营平台

**阶段性策略**：
- **第一阶段（当前）**：专注商家服务
  - 商家入驻与管理
  - 商品与订单管理
  - 数据分析工具
  - 商家营销能力

- **后续阶段**：扩展用户运营能力

**小程序结构**：
```
┌─────────────────────────────────────────────────────────────┐
│                      寻梦私域管家                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌─────────────┐              ┌─────────────┐             │
│   │   首页/活动   │              │  商户管理中心  │             │
│   │   (活动板块)  │              │  (商家后台)   │             │
│   │              │              │              │             │
│   │  • 商家活动   │              │  • 店铺管理   │             │
│   │  • 邀请入驻   │              │  • 商品管理   │             │
│   │  • Banner   │              │  • 订单管理   │             │
│   │  • 公告通知   │              │  • 数据分析   │             │
│   └─────────────┘              └─────────────┘             │
│                                                             │
│   ┌─────────────────────────────────────────┐               │
│   │              商家店铺入口                   │               │
│   │    (扫码进入商家专属店铺，查看商品下单)      │               │
│   └─────────────────────────────────────────┘               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**首页活动板块**：
- 平台统一活动Banner轮播
- 限时优惠、节日活动等
- 商家排行榜入口
- 系统公告

**商户管理中心**：
- 商家登录后进入
- 独立管理后台
- 商品、订单、数据分析

### 1.3 商家邀请入驻机制
为加速商家生态发展，引入邀请入驻机制：

```
┌─────────────────────────────────────────────────────────────────────┐
│                        邀请入驻奖励体系                               │
└─────────────────────────────────────────────────────────────────────┘
                                    │
          ┌─────────────────────────┼─────────────────────────┐
          │                         │                         │
          ▼                         ▼                         ▼
   ┌─────────────┐           ┌─────────────┐           ┌─────────────┐
   │   免年费    │           │  0.2%费率   │           │  推荐奖励   │
   │  (限时活动) │           │ (最低费率)   │           │  (待定)    │
   └─────────────┘           └─────────────┘           └─────────────┘

邀请流程：
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  商家A    │───▶│ 生成邀请码 │───▶│ 分享邀请  │───▶│ 商家B入驻 │───▶│ 奖励发放 │
│ (邀请人)  │    │          │    │ (海报/链接) │    │ (被邀请人) │    │ (年费/费率) │
└──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘
```

**邀请规则**：
| 奖励类型 | 条件 | 说明 |
|----------|------|------|
| 免年费 | 被邀请商家完成入驻并首次交易 | 邀请人可获得一定期限的免年费 |
| 最低费率 | 被邀请商家完成入驻 | 可申请最低0.2%交易费率 |
| 累计邀请 | 多商家入驻 | 可叠加享受更多权益 |

### 1.2 业务场景
本项目采用**统一小程序 + 商家独立入口**的模式：

```
┌─────────────────────────────────────────────────────────────┐
│                      统一小程序                              │
│                   (一个小程序承载所有商家)                    │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
  ┌───────────┐         ┌───────────┐         ┌───────────┐
  │ 商家A二维码 │         │ 商家B二维码 │         │ 商家C二维码 │
  │  扫码进入   │         │  扫码进入   │         │  扫码进入   │
  └─────┬─────┘         └─────┬─────┘         └─────┬─────┘
        │                     │                     │
        ▼                     ▼                     ▼
  ┌───────────┐         ┌───────────┐         ┌───────────┐
  │ 商家A店铺  │         │ 商家B店铺  │         │ 商家C店铺  │
  │ 商品/下单  │         │ 商品/下单  │         │ 商品/下单  │
  └───────────┘         └───────────┘         └───────────┘
```

**用户使用流程：**
1. 用户扫描商家二维码 → 进入统一小程序（携带商家ID参数）
2. 小程序展示该商家的店铺信息、商品列表
3. 用户浏览商品、加入购物车、下单支付
4. 支付资金直接结算到该商家的子商户账户

**开发策略：**
- 当前阶段以**单商家视角**为切入点，优先满足单个商家的完整服务能力
- 后续扩展多商家管理、服务商数据看板等功能

### 1.3 目标用户
- **服务商管理员**：管理所有商家，处理商家进件，查看整体运营数据
- **商家管理员**：管理自己的店铺、商品、订单、数据分析
- **C端消费者**：通过扫码进入商家店铺，浏览商品、下单购买

### 1.4 核心价值
- **服务商层面**：统一管理多商家，支持商家进件，资金清算透明
- **商家层面**：独立管理店铺，灵活配置商品，完整的经营工具
- **用户层面**：扫码即用，无需关注多个小程序，体验统一
- **微信支付无缝对接**：服务商模式下的子商户进件与资金分账

### 1.5 图片存储方案

本项目使用**七牛云存储**作为图片存储解决方案：

```
┌─────────────────────────────────────────────────────────────┐
│                      图片上传流程                            │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
  ┌───────────┐         ┌───────────┐         ┌───────────┐
  │ 商家Logo   │         │ 商品图片   │         │ 营业执照   │
  │ 门头照片   │         │ 商品详情   │         │ 身份证照片 │
  └───────────┘         └───────────┘         └───────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   七牛云存储     │
                    │  (CDN加速访问)   │
                    └─────────────────┘
```

**上传方式：**
1. 客户端直传：前端获取七牛云上传凭证后直接上传，减少服务器压力
2. 服务端代理：敏感图片（如营业执照、身份证）通过服务端上传，便于审核

**存储空间规划：**
| 空间名称 | 用途 | 访问权限 |
|----------|------|----------|
| merchant-public | 商家Logo、店铺图片 | 公开 |
| product-public | 商品图片 | 公开 |
| merchant-private | 营业执照、身份证照片 | 私有（带签名访问） |

**相关接口：**
```
GET /api/v1/upload/token      # 获取上传凭证
POST /api/v1/upload/callback  # 上传回调（可选）
```

### 1.6 微信支付服务商模式说明

#### 1.6.1 服务商模式架构
```
┌─────────────────────────────────────────────────────┐
│                    服务商商户号                       │
│              (微信支付服务商平台账号)                  │
│                                                      │
│  职责：                                              │
│  - 管理子商户（商家）                                 │
│  - 处理进件申请                                       │
│  - 统一对账                                          │
│  - 收取技术服务费                                     │
└─────────────────────┬───────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
        ▼             ▼             ▼
┌───────────┐  ┌───────────┐  ┌───────────┐
│ 子商户A   │  │ 子商户B   │  │ 子商户C   │
│ (商家1)   │  │ (商家2)   │  │ (商家3)   │
│           │  │           │  │           │
│ 资金独立  │  │ 资金独立  │  │ 资金独立  │
│ 结算到账  │  │ 结算到账  │  │ 结算到账  │
└───────────┘  └───────────┘  └───────────┘
```

#### 1.6.2 商家进件流程
商家进件是指服务商向微信支付申请为商家创建子商户号的过程：

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   商家提交    │────▶│  服务商审核   │────▶│  微信支付审核  │────▶│   进件成功    │
│   进件资料    │     │   (平台审核)  │     │  (1-3工作日)  │     │  获得子商户号  │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
```

**进件所需资料：**
| 资料类型 | 必填 | 说明 |
|----------|------|------|
| 商户名称 | 是 | 营业执照上的名称 |
| 统一社会信用代码 | 是 | 营业执照编号 |
| 法人姓名 | 是 | 法定代表人姓名 |
| 法人身份证号 | 是 | 法人身份证号码 |
| 法人身份证照片 | 是 | 正反面照片 |
| 营业执照照片 | 是 | 营业执照图片 |
| 结算银行卡号 | 是 | 用于收款结算 |
| 开户银行 | 是 | 银行名称及支行 |
| 联系人信息 | 是 | 手机号、邮箱 |
| 经营场景 | 是 | 线上/线下/两者都有 |
| 店铺照片 | 是 | 门头照、内景照 |

#### 1.6.3 支付与结算流程
```
用户支付流程：
┌────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐
│  用户   │───▶│  统一小程序  │───▶│  服务商API  │───▶│  微信支付   │
│  下单   │    │  选择商品   │    │  创建订单   │    │  发起支付   │
└────────┘    └────────────┘    └────────────┘    └────────────┘
                                                          │
                                                          ▼
┌────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐
│  用户   │◀───│  支付成功   │◀───│  支付回调   │◀───│  扣款成功   │
│  收到   │    │  页面提示   │    │  更新订单   │    │  通知结果   │
└────────┘    └────────────┘    └────────────┘    └────────────┘

资金结算流程：
┌────────────────────────────────────────────────────────────┐
│                      用户支付金额                           │
└─────────────────────────┬──────────────────────────────────┘
                          │
                          ▼
┌────────────────────────────────────────────────────────────┐
│                    微信支付清算                             │
│  ┌────────────────────┐    ┌────────────────────┐          │
│  │  子商户到账金额     │    │  服务商技术服务费   │          │
│  │  (商家收入)        │    │  (可选，费率可配置)  │          │
│  │  金额 × (1-费率)   │    │  金额 × 费率        │          │
│  └────────────────────┘    └────────────────────┘          │
└────────────────────────────────────────────────────────────┘
```

---

## 2. 功能模块

### 2.1 服务商管理模块
| 功能 | 描述 | 优先级 |
|------|------|--------|
| 服务商入驻 | 服务商注册与资质审核 | P0 |
| 服务商配置 | 微信支付服务商配置、API密钥、证书管理 | P0 |
| 商家进件管理 | 处理商家进件申请，提交微信支付审核 | P0 |
| 进件状态查询 | 查询商家进件审核进度和结果 | P0 |
| 商家审核 | 审核商家入驻申请（平台层面） | P0 |
| 数据看板 | 服务商级别的数据统计 | P1 |
| 邀请入驻管理 | 管理邀请码、查看邀请记录、设置邀请奖励 | P1 |
| 首页活动管理 | 管理首页Banner、活动板块 | P1 |

### 2.2 商家管理模块
| 功能 | 描述 | 优先级 |
|------|------|--------|
| 商家入驻 | 商家注册、提交资质资料 | P0 |
| 商家进件申请 | 提交微信支付进件所需资料 | P0 |
| 进件状态跟踪 | 查看进件审核进度 | P0 |
| 商家信息 | 商家基本信息管理 | P0 |
| 商家设置 | 营业执照、门店公告、营业时间等自定义设置 | P0 |
| 商家二维码 | 生成商家专属小程序码 | P0 |
| 商家状态 | 开启/关闭店铺 | P0 |
| 邀请入驻 | 生成邀请码、邀请商家入驻、查看邀请奖励 | P1 |
| 商家员工 | 员工账号管理 | P2 |

### 2.3 商品分类模块
| 功能 | 描述 | 优先级 |
|------|------|--------|
| 分类创建 | 商家创建自己的商品分类 | P0 |
| 分类编辑 | 修改分类名称、排序 | P0 |
| 分类删除 | 删除分类（需检查关联商品） | P0 |
| 分类排序 | 调整分类显示顺序 | P1 |

### 2.4 商品管理模块
| 功能 | 描述 | 优先级 |
|------|------|--------|
| 商品创建 | 创建商品信息 | P0 |
| 商品编辑 | 修改商品信息 | P0 |
| 商品上下架 | 控制商品是否可售 | P0 |
| 商品删除 | 删除商品（软删除） | P0 |
| 库存管理 | 商品库存数量管理 | P1 |
| 规格管理 | 商品多规格支持 | P2 |

### 2.5 订单管理模块
| 功能 | 描述 | 优先级 |
|------|------|--------|
| 创建订单 | 用户下单 | P0 |
| 支付订单 | 微信支付集成（服务商模式） | P0 |
| 订单列表 | 商家查看订单列表 | P0 |
| 订单详情 | 查看订单详细信息 | P0 |
| 订单状态 | 待支付/已支付/已完成/已取消/已退款 | P0 |
| 订单核销 | 线下核销订单 | P1 |
| 订单退款 | 处理退款申请 | P1 |

### 2.6 数据分析模块
| 功能 | 描述 | 优先级 |
|------|------|--------|
| 销售统计 | 销售额、订单量统计 | P0 |
| 商品分析 | 商品销量排行、库存预警 | P1 |
| 时段分析 | 分时段销售趋势 | P1 |
| 用户分析 | 用户消费行为分析 | P2 |

### 2.7 C端用户模块（扫码进入商家店铺）
| 功能 | 描述 | 优先级 |
|------|------|--------|
| 微信登录 | 用户授权登录 | P0 |
| 扫码进店 | 扫描商家二维码进入店铺首页 | P0 |
| 店铺首页 | 展示商家信息、公告、商品分类 | P0 |
| 商品浏览 | 浏览当前商家的商品列表 | P0 |
| 商品详情 | 查看商品详细信息、规格选择 | P0 |
| 购物车 | 加入购物车、修改数量 | P1 |
| 下单支付 | 确认订单、微信支付 | P0 |
| 我的订单 | 查看个人订单（按商家分组） | P0 |
| 订单详情 | 查看订单详情、申请退款 | P0 |

### 2.8 服务号通知模块
| 功能 | 描述 | 优先级 |
|------|------|--------|
| 模板消息配置 | 服务商配置微信服务号模板消息 | P1 |
| 商家订阅配置 | 商家配置需要接收的通知类型 | P1 |
| 订单下单通知 | 用户下单后推送给商家 | P0 |
| 订单支付通知 | 订单支付成功推送给商家 | P0 |
| 订单退款通知 | 退款申请/退款成功通知商家 | P1 |
| 新订单声音提醒 | 支持商家端开启/关闭声音提醒 | P2 |

**通知流程**：
```
┌────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐
│ 用户下单 │ ──▶│  后端处理   │ ──▶│  微信模板  │ ──▶│  商家服务号 │
│        │    │  创建订单   │    │  消息推送   │    │   收到通知  │
└────────┘    └────────────┘    └────────────┘    └────────────┘
```

### 2.9 云打印模块
| 功能 | 描述 | 优先级 |
|------|------|--------|
| 打印机配置 | 添加/编辑/删除云打印机 | P1 |
| 打印机管理 | 查看打印机列表、状态 | P1 |
| 自动打印开关 | 商家开启/关闭自动打印 | P1 |
| 打印模板设置 | 设置小票打印格式 | P2 |
| 打印记录 | 查看历史打印记录 | P2 |
| 打印测试 | 测试打印机连接 | P1 |

**云打印流程**：
```
┌────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐
│ 订单创建 │ ──▶│  触发打印  │ ──▶│ 云打印API  │ ──▶│  打印机   │
│        │    │  条件判断   │    │  (易联云等) │    │  打印小票  │
└────────┘    └────────────┘    └────────────┘    └────────────┘
```

---

## 3. API 接口设计

### 3.1 认证相关接口

#### 3.1.1 服务商管理员登录
```
POST /api/v1/admin/auth/login
```

**请求参数：**
```json
{
  "username": "管理员账号",
  "password": "密码"
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "token": "JWT Token",
    "admin": {
      "id": 1,
      "username": "admin",
      "name": "管理员名称",
      "role": "service_provider"
    }
  }
}
```

#### 3.1.2 商家管理员登录
```
POST /api/v1/merchant/auth/login
```

**请求参数：**
```json
{
  "username": "商家账号",
  "password": "密码"
}
```

#### 3.1.3 C端用户微信登录
```
POST /api/v1/user/auth/wechat-login
```

**请求参数：**
```json
{
  "code": "微信登录凭证",
  "nickname": "用户昵称",
  "avatar": "头像URL"
}
```

### 3.2 服务商管理接口

#### 3.2.1 服务商信息
```
GET /api/v1/admin/service-provider
Authorization: Bearer {token}
```

#### 3.2.2 更新服务商配置
```
PUT /api/v1/admin/service-provider
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "name": "服务商名称",
  "contact_name": "联系人",
  "contact_phone": "联系电话",
  "wechat_pay_config": {
    "mch_id": "服务商商户号",
    "api_v3_key": "APIv3密钥",
    "cert_serial_no": "证书序列号",
    "private_key": "商户私钥内容",
    "public_key": "平台公钥内容"
  }
}
```

#### 3.2.3 商家进件申请列表
```
GET /api/v1/admin/merchant-applications
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| status | string | 否 | 状态：draft/submitted/auditing/approved/rejected |

**响应：**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "merchant_name": "美味餐厅",
        "contact_name": "张三",
        "contact_phone": "13800138000",
        "status": "auditing",
        "applyment_id": "微信支付申请单号",
        "sub_mch_id": null,
        "submit_time": "2024-01-01T10:00:00Z",
        "audit_detail": null
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 10
  }
}
```

#### 3.2.4 商家进件详情
```
GET /api/v1/admin/merchant-applications/{application_id}
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "merchant_name": "美味餐厅",
    "business_license": {
      "license_no": "91110000...",
      "license_name": "美味餐厅",
      "license_image": "图片URL",
      "legal_person": "张三",
      "legal_person_id": "身份证号",
      "legal_person_id_front": "身份证正面图片",
      "legal_person_id_back": "身份证反面图片"
    },
    "bank_account": {
      "bank_name": "中国工商银行",
      "bank_branch": "北京分行",
      "account_no": "银行卡号",
      "account_name": "账户名称"
    },
    "store_info": {
      "store_name": "美味餐厅",
      "store_address": "店铺地址",
      "store_images": ["门头照", "内景照"]
    },
    "contact_info": {
      "contact_name": "张三",
      "contact_phone": "13800138000",
      "contact_email": "test@example.com"
    },
    "status": "auditing",
    "applyment_id": "微信支付申请单号",
    "sub_mch_id": null,
    "audit_detail": {
      "audit_time": "2024-01-02T10:00:00Z",
      "audit_result": "审核中"
    }
  }
}
```

#### 3.2.5 提交商家进件（向微信支付申请）
```
POST /api/v1/admin/merchant-applications/{application_id}/submit
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "applyment_id": "微信支付申请单号",
    "status": "auditing"
  }
}
```

#### 3.2.6 查询进件状态
```
GET /api/v1/admin/merchant-applications/{application_id}/status
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "applyment_id": "微信支付申请单号",
    "status": "approved",
    "sub_mch_id": "子商户号",
    "sign_url": "签约链接（如需法人签约）"
  }
}
```

#### 3.2.7 商家审核列表（平台审核）
```
GET /api/v1/admin/merchants/pending
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| status | string | 否 | 审核状态：pending/approved/rejected |

#### 3.2.8 审核商家（平台审核）
```
POST /api/v1/admin/merchants/{merchant_id}/audit
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "status": "approved",
  "remark": "审核备注"
}
```

#### 3.2.9 服务商数据看板
```
GET /api/v1/admin/dashboard
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "total_merchants": 100,
    "active_merchants": 80,
    "pending_applications": 5,
    "total_orders": 10000,
    "total_sales": 500000.00,
    "today_orders": 150,
    "today_sales": 7500.00,
    "order_trend": [
      {"date": "2024-01-01", "orders": 120, "sales": 6000.00}
    ]
  }
}
```

#### 3.2.10 首页活动管理 - 获取活动列表
```
GET /api/v1/admin/activities
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "banners": [
      {
        "id": 1,
        "title": "新商家入驻优惠",
        "image": "图片URL",
        "link_type": "invite",
        "link_value": ""
      }
    ],
    "announcements": [
      {
        "id": 1,
        "title": "平台升级通知",
        "content": "升级内容...",
        "publish_time": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
```

#### 3.2.11 首页活动管理 - 创建/更新活动
```
POST /api/v1/admin/activities
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "type": "banner",
  "title": "新商家入驻优惠",
  "image": "图片URL",
  "link_type": "invite",
  "sort": 1,
  "status": 1
}
```

| link_type | 描述 |
|-----------|------|
| invite | 跳转邀请页面 |
| merchant | 跳转商家详情 |
| webview | 网页链接 |
| none | 无跳转 |

#### 3.2.12 邀请入驻统计
```
GET /api/v1/admin/invites/stats
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "total_invites": 50,
    "total_rewards_given": 25,
    "total_reward_amount": 0,
    "pending_invites": 5,
    "invite_trend": [
      {"month": "2024-01", "invites": 10},
      {"month": "2024-02", "invites": 15}
    ]
  }
}
```

#### 3.2.13 邀请记录列表
```
GET /api/v1/admin/invites
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| status | string | 否 | 状态：pending/completed/cancelled |

**响应：**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "inviter": {
          "id": 1,
          "name": "美味餐厅",
          "phone": "138****8000"
        },
        "invitee": {
          "id": 2,
          "name": "隔壁小馆",
          "phone": "139****8000"
        },
        "invite_code": "ABC123",
        "reward_type": "free_year",
        "reward_status": "completed",
        "created_at": "2024-01-01T10:00:00Z",
        "completed_at": "2024-01-15T10:00:00Z"
      }
    ],
    "pagination": {
      "total": 50,
      "page": 1,
      "page_size": 10
    }
  }
}
```

#### 3.2.14 设置邀请奖励规则
```
PUT /api/v1/admin/invite-rewards
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "enabled": true,
  "rewards": [
    {
      "type": "free_year",
      "condition": "first_transaction",
      "description": "被邀请商家完成首次交易后，邀请人免年费"
    },
    {
      "type": "lowest_rate",
      "condition": "merchant_joined",
      "description": "被邀请商家入驻后，邀请人可申请最低0.2%费率"
    }
  ]
}
```

### 3.3 商家管理接口

#### 3.3.1 商家入驻申请
```
POST /api/v1/merchant/register
```

**请求参数：**
```json
{
  "name": "商家名称",
  "contact_name": "联系人",
  "contact_phone": "联系电话",
  "contact_email": "联系邮箱",
  "address": "店铺地址",
  "business_category": "餐饮",
  "license": {
    "license_no": "营业执照号",
    "license_name": "营业执照名称",
    "license_image": "营业执照图片URL",
    "legal_person": "法人姓名",
    "legal_person_id": "法人身份证号",
    "legal_person_id_front": "法人身份证正面图片URL",
    "legal_person_id_back": "法人身份证反面图片URL",
    "valid_from": "有效期开始",
    "valid_to": "有效期结束"
  },
  "bank_account": {
    "bank_name": "开户银行",
    "bank_branch": "开户支行",
    "account_no": "银行账号",
    "account_name": "账户名称",
    "account_type": "账户类型：1对公 2对私"
  },
  "store_info": {
    "store_name": "门店名称",
    "store_images": ["门头照URL", "内景照URL"]
  }
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "merchant_id": 1,
    "application_id": 1,
    "status": "draft"
  }
}
```

#### 3.3.2 获取商家信息
```
GET /api/v1/merchant/profile
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "美味餐厅",
    "logo": "店铺Logo",
    "contact_name": "张三",
    "contact_phone": "13800138000",
    "address": "XX市XX区XX路XX号",
    "business_category": "餐饮",
    "status": "active",
    "license": {
      "license_no": "91110000...",
      "license_name": "美味餐厅",
      "license_image": "图片URL",
      "legal_person": "张三",
      "valid_from": "2020-01-01",
      "valid_to": "2030-01-01"
    },
    "settings": {
      "announcement": "今日特惠：全场8折",
      "business_hours": "09:00-22:00",
      "min_order_amount": 20.00,
      "delivery_fee": 5.00
    },
    "wechat_pay": {
      "sub_mch_id": "子商户号",
      "status": "bound",
      "applyment_status": "approved"
    },
    "qrcode_url": "商家小程序码URL",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 3.3.3 更新商家基本信息
```
PUT /api/v1/merchant/profile
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "name": "商家名称",
  "logo": "店铺Logo URL",
  "contact_name": "联系人",
  "contact_phone": "联系电话",
  "address": "店铺地址"
}
```

#### 3.3.4 更新商家自定义设置
```
PUT /api/v1/merchant/settings
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "announcement": "门店公告内容",
  "business_hours": "09:00-22:00",
  "min_order_amount": 20.00,
  "takeout_enabled": true,
  "dine_in_enabled": true,
  "delivery_settings": {
    "enabled": true,
    "base_fee": 5.00,
    "free_delivery_amount": 50.00,
    "distance_rules": [
      {"min_distance": 0, "max_distance": 2, "fee": 0},
      {"min_distance": 2, "max_distance": 5, "fee": 3.00},
      {"min_distance": 5, "max_distance": 10, "fee": 6.00}
    ],
    "max_distance": 10
  }
}
```

**配送距离规则说明：**
| 字段 | 类型 | 描述 |
|------|------|------|
| enabled | boolean | 是否开启配送 |
| base_fee | decimal | 基础配送费 |
| free_delivery_amount | decimal | 满额免配送费金额 |
| distance_rules | array | 按距离收费规则 |
| max_distance | int | 最大配送距离（公里） |

**distance_rules 结构：**
| 字段 | 类型 | 描述 |
|------|------|------|
| min_distance | int | 最小距离（公里），包含 |
| max_distance | int | 最大距离（公里），不包含 |
| fee | decimal | 该距离范围内的配送费 |

**响应：**
```json
{
  "code": 0,
  "message": "设置更新成功"
}
```

#### 3.3.5 获取配送设置
```
GET /api/v1/merchant/delivery-settings
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "enabled": true,
    "base_fee": 5.00,
    "free_delivery_amount": 50.00,
    "distance_rules": [
      {"min_distance": 0, "max_distance": 2, "fee": 0},
      {"min_distance": 2, "max_distance": 5, "fee": 3.00},
      {"min_distance": 5, "max_distance": 10, "fee": 6.00}
    ],
    "max_distance": 10
  }
}
```

#### 3.3.6 更新营业执照
```
PUT /api/v1/merchant/license
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "license_no": "营业执照号",
  "license_name": "营业执照名称",
  "license_image": "营业执照图片URL",
  "legal_person": "法人姓名",
  "legal_person_id": "法人身份证号",
  "legal_person_id_front": "法人身份证正面图片URL",
  "legal_person_id_back": "法人身份证反面图片URL",
  "valid_from": "有效期开始",
  "valid_to": "有效期结束"
}
```

#### 3.3.7 更新结算银行卡
```
PUT /api/v1/merchant/bank-account
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "bank_name": "开户银行",
  "bank_branch": "开户支行",
  "account_no": "银行账号",
  "account_name": "账户名称",
  "account_type": 1
}
```

#### 3.3.8 开启/关闭店铺
```
POST /api/v1/merchant/status
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "status": "active"
}
```

#### 3.3.9 获取商家进件状态
```
GET /api/v1/merchant/application/status
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "application_id": 1,
    "status": "approved",
    "applyment_id": "微信支付申请单号",
    "sub_mch_id": "子商户号",
    "submit_time": "2024-01-01T10:00:00Z",
    "audit_time": "2024-01-03T15:00:00Z",
    "audit_detail": {
      "legal_person_validation": "passed",
      "license_validation": "passed",
      "bank_account_validation": "passed"
    }
  }
}
```

#### 3.3.10 获取商家小程序码
```
GET /api/v1/merchant/qrcode
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| page | string | 否 | 小程序页面路径，默认首页 |
| width | int | 否 | 二维码宽度，默认430 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "qrcode_url": "小程序码图片URL",
    "expire_time": null
  }
}
```

#### 3.3.11 生成邀请码
```
POST /api/v1/merchant/invite/generate
Authorization: Bearer {token}
```

> 商家生成专属邀请码，用于邀请其他商家入驻

**响应：**
```json
{
  "code": 0,
  "data": {
    "invite_code": "XM20240101ABCD",
    "qrcode_url": "邀请二维码图片URL",
    "share_link": "https://example.com/invite/XM20240101ABCD",
    "poster_url": "邀请海报图片URL",
    "expire_time": null
  }
}
```

#### 3.3.12 获取我的邀请信息
```
GET /api/v1/merchant/invite/info
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "invite_code": "XM20240101ABCD",
    "total_invites": 5,
    "completed_invites": 3,
    "pending_invites": 2,
    "rewards": {
      "free_year_count": 2,
      "lowest_rate_qualified": true
    }
  }
}
```

#### 3.3.13 获取我的邀请记录
```
GET /api/v1/merchant/invite/records
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| status | string | 否 | 状态：pending/completed |

**响应：**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "invitee_name": "隔壁小馆",
        "invitee_phone": "139****8000",
        "invite_code": "XM20240101ABCD",
        "status": "completed",
        "reward_type": "free_year",
        "created_at": "2024-01-01T10:00:00Z",
        "completed_at": "2024-01-15T10:00:00Z"
      }
    ],
    "pagination": {
      "total": 5,
      "page": 1,
      "page_size": 10
    }
  }
}
```

#### 3.3.14 商家入驻（支持邀请码）
```
POST /api/v1/merchant/register
```

**请求参数：**
```json
{
  "name": "商家名称",
  "contact_name": "联系人",
  "contact_phone": "联系电话",
  "contact_email": "联系邮箱",
  "address": "店铺地址",
  "business_category": "餐饮",
  "invite_code": "XM20240101ABCD",
  "license": {
    "license_no": "营业执照号",
    "license_name": "营业执照名称",
    "license_image": "营业执照图片URL",
    "legal_person": "法人姓名",
    "legal_person_id": "法人身份证号",
    "legal_person_id_front": "法人身份证正面图片URL",
    "legal_person_id_back": "法人身份证反面图片URL",
    "valid_from": "有效期开始",
    "valid_to": "有效期结束"
  },
  "bank_account": {
    "bank_name": "开户银行",
    "bank_branch": "开户支行",
    "account_no": "银行账号",
    "account_name": "账户名称",
    "account_type": "账户类型：1对公 2对私"
  },
  "store_info": {
    "store_name": "门店名称",
    "store_images": ["门头照URL", "内景照URL"]
  }
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "merchant_id": 1,
    "application_id": 1,
    "status": "draft",
    "invited_by": {
      "name": "邀请商家名称",
      "reward_eligible": true
    }
  }
}
```

### 3.3.15 服务号配置
```
GET /api/v1/admin/wechat-config
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "app_id": "wx1234567890",
    "enabled": true,
    "template_ids": {
      "order_new": "模板消息ID",
      "order_paid": "模板消息ID",
      "order_refund": "模板消息ID"
    }
  }
}
```

### 3.3.16 更新服务号配置
```
PUT /api/v1/admin/wechat-config
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "app_id": "wx1234567890",
  "app_secret": "AppSecret",
  "token": "自定义Token",
  "encoding_aes_key": "43位加密密钥",
  "template_ids": {
    "order_new": "模板消息ID",
    "order_paid": "模板消息ID",
    "order_refund": "模板消息ID"
  }
}
```

### 3.3.17 商家订阅配置
```
GET /api/v1/merchant/subscriptions
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "notify_type": "order_new",
      "enabled": true,
      "push_openid": "商家员工OpenID"
    }
  ]
}
```

### 3.3.18 更新商家订阅配置
```
PUT /api/v1/merchant/subscriptions
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "subscriptions": [
    {
      "notify_type": "order_new",
      "enabled": true,
      "push_openid": "商家员工OpenID"
    },
    {
      "notify_type": "order_paid",
      "enabled": true,
      "push_openid": "商家员工OpenID"
    }
  ]
}
```

### 3.3.19 获取云打印机列表
```
GET /api/v1/merchant/printers
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "name": "前台打印机",
      "type": "yilianyun",
      "device_no": "设备编号",
      "status": 1,
      "auto_print": true,
      "is_default": true,
      "print_count": 156
    }
  ]
}
```

### 3.3.20 添加云打印机
```
POST /api/v1/merchant/printers
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "name": "前台打印机",
  "type": "yilianyun",
  "device_no": "设备编号",
  "api_key": "API密钥",
  "api_url": "https://api.example.com",
  "auto_print": true,
  "is_default": false
}
```

### 3.3.21 更新云打印机
```
PUT /api/v1/merchant/printers/{printer_id}
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "name": "打印机名称",
  "auto_print": true,
  "is_default": true
}
```

### 3.3.22 删除云打印机
```
DELETE /api/v1/merchant/printers/{printer_id}
Authorization: Bearer {token}
```

### 3.3.23 测试云打印机
```
POST /api/v1/merchant/printers/{printer_id}/test
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "success": true,
    "message": "打印测试成功"
  }
}
```

### 3.3.24 获取打印记录
```
GET /api/v1/merchant/print-logs
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

### 3.4 商品分类接口

#### 3.4.1 获取分类列表
```
GET /api/v1/merchant/categories
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "name": "热销推荐",
      "sort": 1,
      "product_count": 10,
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 3.4.2 创建分类
```
POST /api/v1/merchant/categories
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "name": "分类名称",
  "sort": 1
}
```

#### 3.4.3 更新分类
```
PUT /api/v1/merchant/categories/{category_id}
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "name": "新分类名称",
  "sort": 2
}
```

#### 3.4.4 删除分类
```
DELETE /api/v1/merchant/categories/{category_id}
Authorization: Bearer {token}
```

#### 3.4.5 批量排序分类
```
POST /api/v1/merchant/categories/sort
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "orders": [
    {"id": 1, "sort": 1},
    {"id": 2, "sort": 2}
  ]
}
```

### 3.5 商品管理接口

#### 3.5.1 商品列表
```
GET /api/v1/merchant/products
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| category_id | int | 否 | 分类ID |
| status | string | 否 | 状态：on_sale/off_sale |
| keyword | string | 否 | 搜索关键词 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "招牌红烧肉",
        "images": ["图片URL"],
        "price": 58.00,
        "original_price": 68.00,
        "stock": 100,
        "sales": 256,
        "category_id": 1,
        "category_name": "热销推荐",
        "status": "on_sale",
        "sort": 1,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 10
  }
}
```

#### 3.5.2 创建商品
```
POST /api/v1/merchant/products
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "name": "商品名称",
  "description": "商品描述",
  "images": ["图片URL1", "图片URL2"],
  "category_id": 1,
  "price": 58.00,
  "original_price": 68.00,
  "stock": 100,
  "unit": "份",
  "sort": 1,
  "specs": [
    {
      "name": "规格",
      "options": [
        {"name": "小份", "price": 48.00, "stock": 50},
        {"name": "大份", "price": 68.00, "stock": 50}
      ]
    }
  ]
}
```

#### 3.5.3 更新商品
```
PUT /api/v1/merchant/products/{product_id}
Authorization: Bearer {token}
```

#### 3.5.4 商品上架
```
POST /api/v1/merchant/products/{product_id}/on-sale
Authorization: Bearer {token}
```

#### 3.5.5 商品下架
```
POST /api/v1/merchant/products/{product_id}/off-sale
Authorization: Bearer {token}
```

#### 3.5.6 批量上下架
```
POST /api/v1/merchant/products/batch-status
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "product_ids": [1, 2, 3],
  "status": "on_sale"
}
```

#### 3.5.7 删除商品
```
DELETE /api/v1/merchant/products/{product_id}
Authorization: Bearer {token}
```

#### 3.5.8 更新库存
```
PUT /api/v1/merchant/products/{product_id}/stock
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "stock": 100,
  "action": "set"
}
```

| action | 描述 |
|--------|------|
| set | 设置为指定值 |
| add | 增加指定数量 |
| subtract | 减少指定数量 |

### 3.6 订单管理接口

#### 3.6.1 订单列表
```
GET /api/v1/merchant/orders
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| status | string | 否 | 订单状态 |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |
| order_no | string | 否 | 订单号搜索 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "order_no": "202401010001",
        "user": {
          "id": 1,
          "nickname": "用户昵称",
          "avatar": "头像URL"
        },
        "total_amount": 116.00,
        "pay_amount": 106.00,
        "discount_amount": 10.00,
        "status": "paid",
        "items": [
          {
            "product_id": 1,
            "product_name": "招牌红烧肉",
            "image": "图片URL",
            "price": 58.00,
            "quantity": 2,
            "specs": "大份"
          }
        ],
        "remark": "少放辣",
        "created_at": "2024-01-01T12:00:00Z",
        "paid_at": "2024-01-01T12:01:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

#### 3.6.2 订单详情
```
GET /api/v1/merchant/orders/{order_id}
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "order_no": "202401010001",
    "user": {
      "id": 1,
      "nickname": "用户昵称",
      "avatar": "头像URL",
      "phone": "138****8000"
    },
    "merchant": {
      "id": 1,
      "name": "美味餐厅",
      "address": "店铺地址",
      "phone": "店铺电话"
    },
    "items": [
      {
        "product_id": 1,
        "product_name": "招牌红烧肉",
        "image": "图片URL",
        "price": 58.00,
        "quantity": 2,
        "specs": "大份",
        "subtotal": 116.00
      }
    ],
    "total_amount": 116.00,
    "pay_amount": 106.00,
    "discount_amount": 10.00,
    "delivery_fee": 0,
    "status": "paid",
    "remark": "少放辣",
    "transaction_id": "微信支付交易号",
    "created_at": "2024-01-01T12:00:00Z",
    "paid_at": "2024-01-01T12:01:00Z",
    "completed_at": null,
    "refunded_at": null
  }
}
```

#### 3.6.3 订单核销
```
POST /api/v1/merchant/orders/{order_id}/complete
Authorization: Bearer {token}
```

> **重要说明**：商家只能核销**属于自己店铺**的订单，不能跨商家核销。系统会验证订单的 merchant_id 是否与当前登录商家一致。

**请求参数：**
```json
{
  "verify_code": "核销码"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "核销成功",
  "data": {
    "order_id": 1,
    "order_no": "202401010001",
    "completed_at": "2024-01-01T14:00:00Z"
  }
}
```

**错误情况：**
| 错误码 | 描述 |
|--------|------|
| 5001 | 订单不存在 |
| 5002 | 订单状态错误（非已支付状态） |
| 5005 | 无权操作此订单（跨商家核销） |
| 5006 | 核销码错误 |

#### 3.6.4 订单退款
```
POST /api/v1/merchant/orders/{order_id}/refund
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "refund_amount": 106.00,
  "refund_reason": "退款原因"
}
```

#### 3.6.5 订单统计
```
GET /api/v1/merchant/orders/statistics
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "total_orders": 1000,
    "total_sales": 50000.00,
    "pending_payment": 10,
    "pending_complete": 50,
    "completed": 900,
    "refunded": 40,
    "cancelled": 0
  }
}
```

### 3.7 数据分析接口

#### 3.7.1 销售概览
```
GET /api/v1/merchant/analytics/overview
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| period | string | 否 | 周期：today/week/month/year |

**响应：**
```json
{
  "code": 0,
  "data": {
    "total_sales": 50000.00,
    "total_orders": 1000,
    "total_customers": 500,
    "avg_order_amount": 50.00,
    "sales_growth": 15.5,
    "orders_growth": 10.2,
    "customers_growth": 8.3
  }
}
```

#### 3.7.2 销售趋势
```
GET /api/v1/merchant/analytics/sales-trend
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| start_date | string | 是 | 开始日期 |
| end_date | string | 是 | 结束日期 |
| granularity | string | 否 | 粒度：day/week/month |

**响应：**
```json
{
  "code": 0,
  "data": [
    {
      "date": "2024-01-01",
      "sales": 5000.00,
      "orders": 100,
      "customers": 80
    }
  ]
}
```

#### 3.7.3 商品销量排行
```
GET /api/v1/merchant/analytics/product-ranking
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |
| limit | int | 否 | 返回数量，默认10 |
| sort_by | string | 否 | 排序字段：sales/amount |

**响应：**
```json
{
  "code": 0,
  "data": [
    {
      "rank": 1,
      "product_id": 1,
      "product_name": "招牌红烧肉",
      "image": "图片URL",
      "sales_count": 256,
      "sales_amount": 14848.00
    }
  ]
}
```

#### 3.7.4 时段分析
```
GET /api/v1/merchant/analytics/hourly
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| date | string | 否 | 日期，默认今天 |

**响应：**
```json
{
  "code": 0,
  "data": [
    {
      "hour": 9,
      "orders": 15,
      "sales": 750.00
    },
    {
      "hour": 10,
      "orders": 25,
      "sales": 1250.00
    }
  ]
}
```

#### 3.7.5 库存预警
```
GET /api/v1/merchant/analytics/stock-alert
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| threshold | int | 否 | 库存阈值，默认10 |

**响应：**
```json
{
  "code": 0,
  "data": [
    {
      "product_id": 1,
      "product_name": "招牌红烧肉",
      "stock": 5,
      "status": "low_stock"
    }
  ]
}
```

#### 3.7.6 商家用户分析（商家维度）
```
GET /api/v1/merchant/analytics/customers
Authorization: Bearer {token}
```

> **说明**：此接口从商家维度分析用户消费行为，仅统计在当前商家店铺消费的用户数据。

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "total_customers": 500,
    "new_customers": 50,
    "returning_customers": 450,
    "return_rate": 90.0,
    "customer_stats": {
      "avg_order_count": 3.2,
      "avg_order_amount": 156.00,
      "max_order_amount": 500.00,
      "min_order_amount": 20.00
    },
    "customer_distribution": [
      {
        "range": "1次",
        "count": 150,
        "percentage": 30.0
      },
      {
        "range": "2-5次",
        "count": 250,
        "percentage": 50.0
      },
      {
        "range": "6-10次",
        "count": 70,
        "percentage": 14.0
      },
      {
        "range": "10次以上",
        "count": 30,
        "percentage": 6.0
      }
    ],
    "top_customers": [
      {
        "user_id": 1,
        "nickname": "用户昵称",
        "avatar": "头像URL",
        "order_count": 25,
        "total_amount": 2500.00,
        "last_order_time": "2024-01-01T12:00:00Z"
      }
    ]
  }
}
```

#### 3.7.7 用户消费趋势
```
GET /api/v1/merchant/analytics/customer-trend
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| start_date | string | 是 | 开始日期 |
| end_date | string | 是 | 结束日期 |
| granularity | string | 否 | 粒度：day/week/month |

**响应：**
```json
{
  "code": 0,
  "data": [
    {
      "date": "2024-01-01",
      "new_customers": 10,
      "returning_customers": 40,
      "total_customers": 50
    }
  ]
}
```

### 3.8 C端用户接口（扫码进入商家店铺）

> **说明**：用户通过扫描商家二维码进入统一小程序，小程序携带商家ID参数，所有操作都在当前商家店铺上下文中进行。

#### 3.8.1 微信登录
```
POST /api/v1/user/auth/wechat-login
```

**请求参数：**
```json
{
  "code": "微信登录凭证",
  "nickname": "用户昵称",
  "avatar": "头像URL"
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "token": "JWT Token",
    "user": {
      "id": 1,
      "nickname": "用户昵称",
      "avatar": "头像URL"
    }
  }
}
```

#### 3.8.2 店铺首页（扫码进入）
```
GET /api/v1/store/{merchant_id}/home
```

> 用户扫描商家二维码后，小程序调用此接口获取店铺首页数据

**响应：**
```json
{
  "code": 0,
  "data": {
    "merchant": {
      "id": 1,
      "name": "美味餐厅",
      "logo": "店铺Logo",
      "images": ["店铺图片"],
      "address": "店铺地址",
      "phone": "店铺电话",
      "business_hours": "09:00-22:00",
      "announcement": "今日特惠：全场8折",
      "status": "open",
      "rating": 4.8,
      "sales_count": 1000
    },
    "categories": [
      {
        "id": 1,
        "name": "热销推荐",
        "sort": 1,
        "product_count": 10
      }
    ],
    "hot_products": [
      {
        "id": 1,
        "name": "招牌红烧肉",
        "image": "图片URL",
        "price": 58.00,
        "original_price": 68.00,
        "sales": 256
      }
    ]
  }
}
```

#### 3.8.3 商家商品列表
```
GET /api/v1/store/{merchant_id}/products
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| category_id | int | 否 | 分类ID，不传则返回全部分类商品 |

**响应：**
```json
{
  "code": 0,
  "data": [
    {
      "category": {
        "id": 1,
        "name": "热销推荐"
      },
      "products": [
        {
          "id": 1,
          "name": "招牌红烧肉",
          "image": "图片URL",
          "price": 58.00,
          "original_price": 68.00,
          "sales": 256,
          "stock": 100,
          "specs": [
            {"name": "小份", "price": 48.00},
            {"name": "大份", "price": 68.00}
          ]
        }
      ]
    }
  ]
}
```

#### 3.8.4 商品详情
```
GET /api/v1/store/{merchant_id}/products/{product_id}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "招牌红烧肉",
    "images": ["图片URL1", "图片URL2"],
    "description": "商品描述",
    "price": 58.00,
    "original_price": 68.00,
    "sales": 256,
    "stock": 100,
    "unit": "份",
    "specs": [
      {
        "id": 1,
        "name": "规格",
        "options": [
          {"id": 1, "name": "小份", "price": 48.00, "stock": 50},
          {"id": 2, "name": "大份", "price": 68.00, "stock": 50}
        ]
      }
    ]
  }
}
```

#### 3.8.5 获取配送费规则
```
GET /api/v1/store/{merchant_id}/delivery-rules
```

> **说明**：获取商家的配送费规则，用户选择配送距离后前端计算配送费

**响应：**
```json
{
  "code": 0,
  "data": {
    "enabled": true,
    "base_fee": 5.00,
    "free_delivery_amount": 50.00,
    "max_distance": 10,
    "rules": [
      {"min_distance": 0, "max_distance": 2, "fee": 0},
      {"min_distance": 2, "max_distance": 5, "fee": 3.00},
      {"min_distance": 5, "max_distance": 10, "fee": 6.00}
    ]
  }
}
```

**说明**：
- 用户选择配送距离（如：3公里）
- 前端根据 `rules` 匹配对应区间的配送费
- 如果订单金额 >= `free_delivery_amount`，配送费为 0

#### 3.8.6 创建订单
```
POST /api/v1/store/{merchant_id}/orders
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "items": [
    {
      "product_id": 1,
      "spec_option": "大份",
      "quantity": 2
    }
  ],
  "delivery_type": 1,
  "delivery_distance": 3,
  "remark": "少放辣"
}
```

**delivery_type 说明：**
| 值 | 描述 |
|------|------|
| 1 | 配送（需要填写配送距离） |
| 2 | 堂食（不需要配送费） |
| 3 | 自提（不需要配送费） |

**响应：**
```json
{
  "code": 0,
  "data": {
    "order_id": 1,
    "order_no": "202401010001",
    "total_amount": 116.00,
    "delivery_fee": 3.00,
    "discount_amount": 0,
    "pay_amount": 119.00,
    "delivery_info": {
      "distance": 3.5,
      "address": "收货地址"
    },
    "pay_params": {
      "timeStamp": "时间戳",
      "nonceStr": "随机字符串",
      "package": "prepay_id=xxx",
      "signType": "RSA",
      "paySign": "签名"
    }
  }
}
```

> **说明**：支付参数中的 `paySign` 由服务商模式下的服务商商户号签名，支付资金直接结算到商家的子商户账户。

#### 3.8.7 我的订单列表
```
GET /api/v1/user/orders
Authorization: Bearer {token}
```

**请求参数：**
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| merchant_id | int | 否 | 商家ID，筛选指定商家的订单 |
| status | string | 否 | 订单状态 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "order_no": "202401010001",
        "merchant": {
          "id": 1,
          "name": "美味餐厅",
          "logo": "店铺Logo"
        },
        "items": [
          {
            "product_name": "招牌红烧肉",
            "image": "图片URL",
            "quantity": 2
          }
        ],
        "total_amount": 116.00,
        "status": "paid",
        "status_text": "已支付",
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 10
  }
}
```

#### 3.8.8 订单详情
```
GET /api/v1/user/orders/{order_id}
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "order_no": "202401010001",
    "merchant": {
      "id": 1,
      "name": "美味餐厅",
      "logo": "店铺Logo",
      "address": "店铺地址",
      "phone": "店铺电话"
    },
    "items": [
      {
        "product_id": 1,
        "product_name": "招牌红烧肉",
        "image": "图片URL",
        "price": 58.00,
        "quantity": 2,
        "spec_info": "大份",
        "subtotal": 116.00
      }
    ],
    "total_amount": 116.00,
    "delivery_fee": 3.00,
    "pay_amount": 119.00,
    "discount_amount": 0,
    "delivery_info": {
      "type": "delivery",
      "address": "收货地址",
      "contact_name": "联系人",
      "contact_phone": "联系电话",
      "distance": 3.5
    },
    "status": "paid",
    "status_text": "已支付",
    "remark": "少放辣",
    "verify_code": "123456",
    "created_at": "2024-01-01T12:00:00Z",
    "paid_at": "2024-01-01T12:01:00Z"
  }
}
```

#### 3.8.9 取消订单
```
POST /api/v1/user/orders/{order_id}/cancel
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 0,
  "message": "订单已取消"
}
```

#### 3.8.10 申请退款
```
POST /api/v1/user/orders/{order_id}/refund
Authorization: Bearer {token}
```

**请求参数：**
```json
{
  "refund_reason": "退款原因"
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "refund_id": 1,
    "refund_no": "RF202401010001",
    "status": "processing"
  }
}
```

---

## 4. 数据模型设计

### 4.1 服务商表 (service_providers)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| name | VARCHAR(128) | 服务商名称 |
| contact_name | VARCHAR(64) | 联系人 |
| contact_phone | VARCHAR(20) | 联系电话 |
| mch_id | VARCHAR(32) | 服务商商户号 |
| api_key | VARCHAR(128) | API密钥（加密存储） |
| api_v3_key | VARCHAR(128) | APIv3密钥（加密存储） |
| cert_serial_no | VARCHAR(64) | 证书序列号 |
| private_key | TEXT | 商户私钥（加密存储） |
| public_key | TEXT | 平台公钥 |
| callback_url | VARCHAR(256) | 回调地址 |
| status | TINYINT | 状态：0禁用 1正常 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.2 服务商管理员表 (service_provider_admins)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| service_provider_id | BIGINT | 服务商ID |
| username | VARCHAR(64) | 用户名 |
| password | VARCHAR(128) | 密码（加密存储） |
| name | VARCHAR(64) | 姓名 |
| phone | VARCHAR(20) | 手机号 |
| role | VARCHAR(32) | 角色：admin/operator |
| status | TINYINT | 状态 |
| last_login_at | DATETIME | 最后登录时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.3 商家进件申请表 (merchant_applications)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| merchant_id | BIGINT | 商家ID |
| merchant_name | VARCHAR(128) | 商户名称 |
| business_license_info | JSON | 营业执照信息 |
| legal_person_info | JSON | 法人信息 |
| bank_account_info | JSON | 结算银行卡信息 |
| store_info | JSON | 门店信息 |
| contact_info | JSON | 联系人信息 |
| applyment_id | VARCHAR(64) | 微信支付申请单号 |
| sub_mch_id | VARCHAR(32) | 子商户号 |
| status | TINYINT | 状态：0草稿 1已提交 2审核中 3通过 4拒绝 |
| audit_detail | JSON | 审核详情 |
| submit_time | DATETIME | 提交时间 |
| audit_time | DATETIME | 审核时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**business_license_info JSON 结构：**
```json
{
  "license_no": "营业执照号",
  "license_name": "营业执照名称",
  "license_image": "营业执照图片URL",
  "valid_from": "有效期开始",
  "valid_to": "有效期结束"
}
```

**legal_person_info JSON 结构：**
```json
{
  "name": "法人姓名",
  "id_type": "证件类型",
  "id_no": "证件号码",
  "id_front": "身份证正面图片URL",
  "id_back": "身份证反面图片URL"
}
```

**bank_account_info JSON 结构：**
```json
{
  "bank_name": "开户银行",
  "bank_branch": "开户支行",
  "bank_code": "银行编码",
  "account_no": "银行账号",
  "account_name": "账户名称",
  "account_type": "账户类型：1对公 2对私"
}
```

**store_info JSON 结构：**
```json
{
  "store_name": "门店名称",
  "store_address": "门店地址",
  "store_images": ["门头照URL", "内景照URL"]
}
```

### 4.4 商家表 (merchants)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| service_provider_id | BIGINT | 服务商ID |
| name | VARCHAR(128) | 商家名称 |
| logo | VARCHAR(512) | 店铺Logo |
| contact_name | VARCHAR(64) | 联系人 |
| contact_phone | VARCHAR(20) | 联系电话 |
| contact_email | VARCHAR(128) | 联系邮箱 |
| address | VARCHAR(256) | 店铺地址 |
| lat | DECIMAL(10,6) | 纬度 |
| lng | DECIMAL(10,6) | 经度 |
| business_category | VARCHAR(64) | 经营类目 |
| business_hours | VARCHAR(64) | 营业时间 |
| announcement | TEXT | 门店公告 |
| min_order_amount | DECIMAL(10,2) | 最低起送金额 |
| takeout_enabled | BOOLEAN | 是否支持外卖 |
| dine_in_enabled | BOOLEAN | 是否支持堂食 |
| sub_mch_id | VARCHAR(32) | 微信支付子商户号 |
| sub_mch_status | TINYINT | 子商户绑定状态：0未绑定 1绑定中 2已绑定 3绑定失败 |
| applyment_status | TINYINT | 进件状态：0未进件 1进件中 2已通过 3已拒绝 |
| audit_status | TINYINT | 平台审核状态：0待审核 1通过 2拒绝 |
| audit_remark | VARCHAR(256) | 审核备注 |
| status | TINYINT | 状态：0关闭 1营业中 |
| rating | DECIMAL(2,1) | 评分 |
| sales_count | INT | 销量 |
| qrcode_url | VARCHAR(512) | 商家小程序码URL |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.5 商家配送设置表 (merchant_delivery_settings)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| merchant_id | BIGINT | 商家ID |
| enabled | BOOLEAN | 是否开启配送 |
| base_fee | DECIMAL(10,2) | 基础配送费 |
| free_delivery_amount | DECIMAL(10,2) | 满额免配送费金额 |
| max_distance | INT | 最大配送距离（公里） |
| distance_rules | JSON | 按距离收费规则 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**distance_rules JSON 结构：**
```json
[
  {"min_distance": 0, "max_distance": 2, "fee": 0},
  {"min_distance": 2, "max_distance": 5, "fee": 3.00},
  {"min_distance": 5, "max_distance": 10, "fee": 6.00}
]
```

### 4.6 商家营业执照表 (merchant_licenses)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| merchant_id | BIGINT | 商家ID |
| license_no | VARCHAR(64) | 营业执照号 |
| license_name | VARCHAR(128) | 营业执照名称 |
| license_image | VARCHAR(512) | 营业执照图片 |
| legal_person | VARCHAR(64) | 法人姓名 |
| legal_person_id | VARCHAR(32) | 法人身份证号 |
| legal_person_id_front | VARCHAR(512) | 法人身份证正面图片 |
| legal_person_id_back | VARCHAR(512) | 法人身份证反面图片 |
| valid_from | DATE | 有效期开始 |
| valid_to | DATE | 有效期结束 |
| status | TINYINT | 状态 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.7 商家员工表 (merchant_staffs)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| merchant_id | BIGINT | 商家ID |
| username | VARCHAR(64) | 用户名 |
| password | VARCHAR(128) | 密码 |
| name | VARCHAR(64) | 姓名 |
| phone | VARCHAR(20) | 手机号 |
| role | VARCHAR(32) | 角色：owner/manager/staff |
| status | TINYINT | 状态 |
| last_login_at | DATETIME | 最后登录时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.8 商品分类表 (categories)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| merchant_id | BIGINT | 商家ID |
| name | VARCHAR(64) | 分类名称 |
| sort | INT | 排序权重 |
| status | TINYINT | 状态：0禁用 1启用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.9 商品表 (products)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| merchant_id | BIGINT | 商家ID |
| category_id | BIGINT | 分类ID |
| name | VARCHAR(128) | 商品名称 |
| description | TEXT | 商品描述 |
| images | JSON | 图片数组 |
| price | DECIMAL(10,2) | 售价 |
| original_price | DECIMAL(10,2) | 原价 |
| stock | INT | 库存 |
| unit | VARCHAR(16) | 单位 |
| sales | INT | 销量 |
| sort | INT | 排序权重 |
| status | TINYINT | 状态：0下架 1上架 |
| deleted_at | DATETIME | 删除时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.10 商品规格表 (product_specs)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| product_id | BIGINT | 商品ID |
| name | VARCHAR(64) | 规格名称（如：份量） |
| options | JSON | 规格选项（存储规格和对应价格，如：[{"name":"小份","price":48.00},{"name":"大份","price":68.00}]） |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.12 C端用户表 (users)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| openid | VARCHAR(64) | 微信OpenID |
| union_id | VARCHAR(64) | 微信UnionID |
| nickname | VARCHAR(64) | 昵称 |
| avatar | VARCHAR(512) | 头像 |
| phone | VARCHAR(20) | 手机号 |
| status | TINYINT | 状态 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.11 订单表 (orders)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| order_no | VARCHAR(32) | 订单号 |
| user_id | BIGINT | 用户ID |
| merchant_id | BIGINT | 商家ID |
| total_amount | DECIMAL(10,2) | 商品总金额 |
| delivery_fee | DECIMAL(10,2) | 配送费 |
| discount_amount | DECIMAL(10,2) | 优惠金额 |
| pay_amount | DECIMAL(10,2) | 实付金额 |
| delivery_type | TINYINT | 配送类型：1配送 2堂食 3自提 |
| delivery_distance | DECIMAL(5,2) | 配送距离（公里） |
| delivery_address | VARCHAR(256) | 配送地址 |
| contact_name | VARCHAR(64) | 联系人姓名 |
| contact_phone | VARCHAR(20) | 联系人电话 |
| status | TINYINT | 状态 |
| remark | VARCHAR(256) | 备注 |
| verify_code | VARCHAR(16) | 核销码 |
| transaction_id | VARCHAR(64) | 微信支付交易号 |
| paid_at | DATETIME | 支付时间 |
| completed_at | DATETIME | 完成时间 |
| cancelled_at | DATETIME | 取消时间 |
| refunded_at | DATETIME | 退款时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**订单状态流转：**
```
pending_payment(待支付) → paid(已支付) → completed(已完成)
        ↓                    ↓
   cancelled(已取消)    refunding(退款中) → refunded(已退款)
```

### 4.12 订单商品表 (order_items)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| order_id | BIGINT | 订单ID |
| merchant_id | BIGINT | 商家ID |
| product_id | BIGINT | 商品ID |
| product_name | VARCHAR(128) | 商品名称 |
| image | VARCHAR(512) | 商品图片 |
| price | DECIMAL(10,2) | 单价 |
| quantity | INT | 数量 |
| spec_info | JSON | 规格信息（如：大份/小份） |
| subtotal | DECIMAL(10,2) | 小计 |
| created_at | DATETIME | 创建时间 |

### 4.13 退款记录表 (refunds)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| order_id | BIGINT | 订单ID |
| refund_no | VARCHAR(32) | 退款单号 |
| refund_amount | DECIMAL(10,2) | 退款金额 |
| refund_reason | VARCHAR(256) | 退款原因 |
| status | TINYINT | 状态：0处理中 1成功 2失败 |
| refund_id | VARCHAR(64) | 微信退款单号 |
| refunded_at | DATETIME | 退款完成时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.14 邀请记录表 (invite_records)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| inviter_id | BIGINT | 邀请人ID（商家ID） |
| invitee_id | BIGINT | 被邀请人ID（商家ID） |
| invite_code | VARCHAR(32) | 邀请码 |
| status | TINYINT | 状态：0待完成 1已完成 2已取消 |
| reward_type | VARCHAR(32) | 奖励类型：free_year/lowest_rate |
| reward_status | TINYINT | 奖励状态：0未发放 1已发放 |
| created_at | DATETIME | 创建时间 |
| completed_at | DATETIME | 完成时间 |

### 4.15 平台活动表 (activities)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| type | VARCHAR(16) | 类型：banner/announcement |
| title | VARCHAR(128) | 标题 |
| content | TEXT | 内容 |
| image | VARCHAR(512) | 图片URL |
| link_type | VARCHAR(16) | 跳转类型：invite/merchant/webview/none |
| link_value | VARCHAR(256) | 跳转值 |
| sort | INT | 排序 |
| status | TINYINT | 状态：0禁用 1启用 |
| start_time | DATETIME | 开始时间 |
| end_time | DATETIME | 结束时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.16 邀请奖励规则表 (invite_rewards)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| type | VARCHAR(32) | 奖励类型 |
| condition | VARCHAR(32) | 触发条件 |
| description | VARCHAR(256) | 奖励描述 |
| enabled | BOOLEAN | 是否启用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.17 服务号配置表 (wechat_config)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| service_provider_id | BIGINT | 服务商ID |
| app_id | VARCHAR(64) | 微信公众号AppID |
| app_secret | VARCHAR(128) | 微信公众号AppSecret |
| token | VARCHAR(64) | 验证Token |
| encoding_aes_key | VARCHAR(128) | 消息加密密钥 |
| template_ids | JSON | 模板消息ID集合 |
| enabled | BOOLEAN | 是否启用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**template_ids JSON结构**：
```json
{
  "order_new": "模板消息ID_新订单通知",
  "order_paid": "模板消息ID_订单支付通知",
  "order_refund": "模板消息ID_退款通知"
}
```

### 4.18 商家订阅配置表 (merchant_subscriptions)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| merchant_id | BIGINT | 商家ID |
| notify_type | VARCHAR(32) | 通知类型：order_new/order_paid/order_refund |
| enabled | BOOLEAN | 是否启用 |
| push_openid | VARCHAR(64) | 接收通知的OpenID |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.19 云打印机表 (cloud_printers)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| merchant_id | BIGINT | 商家ID |
| name | VARCHAR(64) | 打印机名称 |
| type | VARCHAR(32) | 打印机类型：yilianyun/feie/xp |
| device_no | VARCHAR(64) | 设备编号 |
| api_key | VARCHAR(128) | API密钥 |
| api_url | VARCHAR(256) | API接口地址 |
| status | TINYINT | 状态：0离线 1在线 |
| auto_print | BOOLEAN | 自动打印开关 |
| print_count | INT | 累计打印次数 |
| is_default | BOOLEAN | 是否默认打印机 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.20 打印记录表 (print_logs)
| 字段 | 类型 | 描述 |
|------|------|------|
| id | BIGINT | 主键 |
| merchant_id | BIGINT | 商家ID |
| printer_id | BIGINT | 打印机ID |
| order_id | BIGINT | 订单ID |
| status | TINYINT | 状态：0失败 1成功 |
| error_message | VARCHAR(256) | 错误信息 |
| print_time | DATETIME | 打印时间 |
| created_at | DATETIME | 创建时间 |

---

## 5. 技术架构

### 5.1 技术栈
| 层级 | 技术选型 | 说明 |
|------|----------|------|
| Web框架 | Gin | Go语言高性能Web框架 |
| ORM | GORM | Go语言ORM库 |
| 数据库 | MySQL 8.0 | 主数据库 |
| 缓存 | Redis | 缓存、Session、分布式锁 |
| 认证 | JWT | 用户认证 |
| 文档 | Swagger | API文档自动生成 |
| 支付 | wechatpay-go | 微信支付SDK |

### 5.2 项目结构

#### 后端项目 (server/)
```
server/                          # Go后端服务
├── cmd/
│   └── server/
│       └── main.go            # 程序入口
├── internal/
│   ├── config/                # 配置管理
│   ├── handlers/               # 处理器层
│   │   ├── admin/             # 服务商管理
│   │   ├── merchant/           # 商家管理
│   │   └── user/              # 用户功能
│   ├── middleware/            # 中间件
│   ├── models/                # 数据模型
│   └── utils/                 # 工具函数
├── pkg/
│   ├── database/              # 数据库连接
│   └── response/              # 统一响应
├── migrations/                # 数据库迁移
├── .env                       # 环境配置
├── .env.example
├── docker-compose.yml
└── go.mod
```

#### 小程序项目 (miniprogram/)
```
miniprogram/                    # 微信小程序
├── src/
│   ├── App.vue
│   ├── main.ts
│   ├── pages.json
│   ├── manifest.json
│   └── ...
├── package.json
└── ...
```

### 5.3 分层架构
```
┌─────────────────────────────────────────────┐
│              Handlers (Controller)           │  ← 处理HTTP请求/响应
├─────────────────────────────────────────────┤
│              Services (Business)             │  ← 业务逻辑处理
├─────────────────────────────────────────────┤
│           Repositories (Data Access)         │  ← 数据访问层
├─────────────────────────────────────────────┤
│              Models (Entity)                 │  ← 数据模型定义
└─────────────────────────────────────────────┘
```

### 5.4 微信支付服务商模式架构
```
┌─────────────────────────────────────────────────────┐
│                    服务商商户号                       │
│                  (主商户，负责清算)                    │
└─────────────────────┬───────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
        ▼             ▼             ▼
┌───────────┐  ┌───────────┐  ┌───────────┐
│ 子商户A   │  │ 子商户B   │  │ 子商户C   │
│ (商家1)   │  │ (商家2)   │  │ (商家3)   │
└───────────┘  └───────────┘  └───────────┘
```

---

## 6. 开发计划

### 6.1 第一阶段：基础架构 (P0)
- [ ] 项目初始化与依赖配置
- [ ] 数据库连接与迁移脚本
- [ ] 基础错误处理与响应格式
- [ ] JWT认证中间件
- [ ] 日志系统

### 6.2 第二阶段：服务商与商家管理 (P0)
- [ ] 服务商注册与配置
- [ ] 服务商管理员登录
- [ ] 商家入驻申请
- [ ] 商家审核流程
- [ ] 商家信息管理
- [ ] 商家自定义设置

### 6.3 第三阶段：商品管理 (P0)
- [ ] 商品分类CRUD
- [ ] 商品CRUD
- [ ] 商品上下架
- [ ] 商品规格管理
- [ ] 库存管理

### 6.4 第四阶段：订单与支付 (P0)
- [ ] 微信支付服务商对接
- [ ] 子商户绑定
- [ ] 订单创建
- [ ] 支付回调处理
- [ ] 订单状态管理
- [ ] 订单核销
- [ ] 退款处理

### 6.5 第五阶段：数据分析 (P1)
- [ ] 销售统计
- [ ] 商品销量排行
- [ ] 时段分析
- [ ] 库存预警

### 6.6 第六阶段：优化迭代 (P2)
- [ ] 员工管理
- [ ] 缓存优化
- [ ] 性能调优
- [ ] 监控告警

### 6.7 小程序开发计划

#### 技术选型
- **开发框架**：uni-app（Vue3 + TypeScript）
- **UI组件**：uView / Vant Weapp
- **状态管理**：Pinia
- **样式规范**：简洁实用型，支持商家自定义模板
- **发布平台**：微信小程序（可扩展至H5/APP）

#### 开发原则
1. **简洁实用** - 第一版本专注于功能打通，不追求复杂UI
2. **商家模板预留** - 组件设计预留样式扩展接口，便于后续商家店铺模板
3. **接口优先** - 先确保API对接，再优化界面

#### 页面优先级（按批次）

**第一批次：认证与入驻**
| 页面 | 功能描述 | 对接API |
|------|----------|---------|
| auth/login | 商家登录（账号密码） | POST /api/v1/auth/merchant/login |
| auth/register | 商家入驻申请 | POST /api/v1/merchant/register |
| auth/verify | 入驻状态查询 | GET /api/v1/merchant/application/status |

**第二批次：商户管理中心首页**
| 页面 | 功能描述 | 对接API |
|------|----------|---------|
| merchant/home | 商户后台首页、快捷入口、待处理事项 | GET /api/v1/merchant/profile |
| merchant/orders/statistics | 订单统计卡片 | GET /api/v1/merchant/orders/statistics |
| merchant/analytics/overview | 数据概览 | GET /api/v1/merchant/analytics/overview |

**第三批次：商品管理**
| 页面 | 功能描述 | 对接API |
|------|----------|---------|
| merchant/products/list | 商品列表、搜索、上下架 | GET /api/v1/merchant/products |
| merchant/products/edit | 商品编辑、规格设置 | POST/PUT /api/v1/merchant/products |
| merchant/categories | 分类管理、排序 | GET/POST /api/v1/merchant/categories |

**第四批次：订单管理**
| 页面 | 功能描述 | 对接API |
|------|----------|---------|
| merchant/orders/list | 订单列表、状态筛选 | GET /api/v1/merchant/orders |
| merchant/orders/detail | 订单详情、核销操作 | GET /api/v1/merchant/orders/{id} |

**第五批次：商家设置与邀请入驻**
| 页面 | 功能描述 | 对接API |
|------|----------|---------|
| merchant/settings | 商家信息、营业设置 | PUT /api/v1/merchant/profile |
| merchant/delivery-settings | 配送设置 | GET/PUT /api/v1/merchant/delivery-settings |
| merchant/invite | 邀请码生成、邀请统计 | POST/GET /api/v1/merchant/invite/* |

**第六批次：数据分析**
| 页面 | 功能描述 | 对接API |
|------|----------|---------|
| merchant/analytics/sales | 销售趋势图表 | GET /api/v1/merchant/analytics/sales-trend |
| merchant/analytics/products | 商品销量排行 | GET /api/v1/merchant/analytics/product-ranking |
| merchant/analytics/customers | 客户分析 | GET /api/v1/merchant/analytics/customers |

**第七批次：C端店铺购物（扫码进入）**
| 页面 | 功能描述 | 对接API |
|------|----------|---------|
| store/home | 店铺首页、分类商品 | GET /api/v1/store/{id}/home |
| store/products | 商品列表 | GET /api/v1/store/{id}/products |
| store/product | 商品详情 | GET /api/v1/store/{id}/products/{id} |
| store/confirm | 确认订单、支付 | POST /api/v1/user/orders |

**第八批次：微信支付集成**
| 功能 | 描述 | 对接API |
|------|------|---------|
| 支付下单 | 统一下单、获取支付参数 | POST /api/v1/notify/payment |
| 支付回调 | 支付结果通知 | POST /api/v1/notify/payment |
| 退款处理 | 申请退款 | POST /api/v1/merchant/orders/{id}/refund |

#### uni-app 项目结构
```
xm-mall/
├── src/
│   ├── App.vue
│   ├── main.ts
│   ├── pages.json
│   ├── manifest.json
│   ├── uni.scss
│   │
│   ├── api/                    # API接口封装
│   │   ├── index.ts           # 接口统一导出
│   │   ├── auth.ts            # 认证接口
│   │   ├── merchant.ts        # 商户接口
│   │   ├── product.ts         # 商品接口
│   │   ├── order.ts           # 订单接口
│   │   └── store.ts           # 店铺接口
│   │
│   ├── pages/                  # 主包页面
│   │   ├── auth/
│   │   │   ├── login.vue      # 登录页
│   │   │   └── register.vue   # 入驻申请
│   │   │
│   │   ├── merchant/
│   │   │   ├── home.vue       # 商户首页
│   │   │   ├── products/
│   │   │   ├── orders/
│   │   │   ├── analytics/
│   │   │   ├── invite/
│   │   │   └── settings/
│   │   │
│   │   └── store/
│   │       ├── home.vue       # 店铺首页
│   │       ├── products.vue
│   │       ├── product.vue
│   │       └── confirm.vue
│   │
│   ├── components/             # 公共组件
│   │   ├── merchant-card/
│   │   ├── product-card/
│   │   ├── order-card/
│   │   └── stats-card/
│   │
│   ├── composables/           # 组合式函数
│   │   ├── useAuth.ts
│   │   ├── useMerchant.ts
│   │   └── useOrder.ts
│   │
│   ├── stores/                # Pinia状态管理
│   │   ├── auth.ts
│   │   ├── merchant.ts
│   │   └── cart.ts
│   │
│   ├── utils/                 # 工具函数
│   │   ├── request.ts        # 请求封装
│   │   ├── auth.ts           # 认证工具
│   │   ├── payment.ts        # 支付工具
│   │   └── constants.ts       # 常量
│   │
│   ├── types/                 # TypeScript类型
│   │   ├── api.d.ts
│   │   ├── merchant.d.ts
│   │   ├── product.d.ts
│   │   └── order.d.ts
│   │
│   └── static/                # 静态资源
│       └── images/
│
├── package.json
├── tsconfig.json
├── vite.config.ts
└── README.md
```

#### 开发任务清单

**Phase 1：项目初始化与认证**
- [ ] uni-app 项目创建与配置
- [ ] API请求封装（request.ts）
- [ ] Pinia状态管理初始化
- [ ] 登录页面（auth/login）
- [ ] 商家入驻页面（auth/register）
- [ ] Token存储与自动登录

**Phase 2：商户中心基础**
- [ ] 商户首页框架
- [ ] 商家信息获取与展示
- [ ] 快捷入口组件
- [ ] 待处理订单统计卡片

**Phase 3：商品管理**
- [ ] 商品分类管理
- [ ] 商品列表（支持搜索、筛选）
- [ ] 商品编辑（基础信息）
- [ ] 商品规格设置
- [ ] 上下架操作

**Phase 4：订单管理**
- [ ] 订单列表（状态筛选）
- [ ] 订单详情
- [ ] 订单核销功能
- [ ] 订单统计

**Phase 5：商家设置**
- [ ] 商家信息编辑
- [ ] 配送设置
- [ ] 营业执照上传
- [ ] 邀请入驻功能

**Phase 6：通知与打印**
- [ ] 服务号订阅配置（接收订单通知）
- [ ] 商家员工OpenID绑定
- [ ] 云打印机配置（添加/编辑/删除）
- [ ] 自动打印开关
- [ ] 打印测试功能
- [ ] 打印记录查看

**Phase 7：数据分析**
- [ ] 销售概览
- [ ] 销售趋势图表
- [ ] 商品排行
- [ ] 客户分析

**Phase 8：C端购物**
- [ ] 店铺首页
- [ ] 商品浏览
- [ ] 商品详情
- [ ] 订单确认

**Phase 9：微信支付**
- [ ] 统一下单对接
- [ ] 支付签名
- [ ] 支付回调处理
- [ ] 退款功能

#### API对接表

| 功能模块 | 接口路径 | 方法 | 状态 |
|----------|----------|------|------|
| 商家登录 | /api/v1/auth/merchant/login | POST | ✅ 已实现 |
| 商家入驻 | /api/v1/merchant/register | POST | ⚠️ 待完善 |
| 商家信息 | /api/v1/merchant/profile | GET | ✅ 已实现 |
| 商家设置 | /api/v1/merchant/settings | PUT | ✅ 已实现 |
| 商品管理 | /api/v1/merchant/products/* | CRUD | ✅ 已实现 |
| 分类管理 | /api/v1/merchant/categories/* | CRUD | ✅ 已实现 |
| 订单管理 | /api/v1/merchant/orders/* | CRUD | ✅ 已实现 |
| 数据分析 | /api/v1/merchant/analytics/* | GET | ✅ 已实现 |
| 邀请入驻 | /api/v1/merchant/invite/* | CRUD | ✅ 已实现 |
| 店铺浏览 | /api/v1/store/{id}/* | GET | ✅ 已实现 |
| 服务号配置 | /api/v1/admin/wechat-config | GET/PUT | ✅ 已实现 |
| 订阅配置 | /api/v1/merchant/subscriptions | GET/PUT | ✅ 已实现 |
| 云打印机 | /api/v1/merchant/printers/* | CRUD | ✅ 已实现 |
| 打印记录 | /api/v1/merchant/print-logs | GET | ✅ 已实现 |
| 微信支付 | /api/v1/notify/payment | POST | ⚠️ 待完善 |

#### 服务号通知说明

**通知触发场景**：
| 通知类型 | 触发条件 | 接收人 |
|----------|----------|--------|
| order_new | 用户创建订单 | 商家管理员 |
| order_paid | 订单支付成功 | 商家管理员 |
| order_refund | 退款申请/成功 | 商家管理员 |

**订阅配置流程**：
```
1. 服务商在后台配置服务号AppID和模板消息ID
2. 商家员工关注服务号，获取OpenID
3. 商家在后台绑定员工的OpenID
4. 订单状态变化时，系统推送模板消息
```

#### 云打印集成说明

**支持的打印机品牌**：
- 易联云（yilianyun）
- 飞鹅打印机（feie）
- 芯烨打印机（xp）

**打印触发条件**：
- 商家开启自动打印
- 订单状态变为"已支付"
- 自动打印至商家设置的默认打印机

#### 微信支付集成说明

**支付流程**：
```
1. 用户下单 → 前端调用 POST /api/v1/user/orders
2. 获取支付参数 → 后端返回支付签名信息
3. 发起支付 → wx.requestPayment()
4. 支付回调 → 后端更新订单状态
5. 支付结果 → 前端展示支付结果
```

**需要后端配合**：
- [ ] 统一下单接口（已预留）
- [ ] 支付回调通知处理
- [ ] 退款接口

---

## 7. 非功能性需求

### 7.1 性能要求
- API 响应时间 < 200ms (P95)
- 支持并发 1000 QPS
- 数据库查询优化，避免 N+1 问题
- 合理使用缓存，热点数据缓存命中率 > 90%

### 7.2 安全要求
- 所有接口使用 HTTPS
- 敏感信息加密存储（密码、支付密钥等）
- SQL 注入防护
- XSS 防护
- 接口限流保护
- 支付签名验证

### 7.3 可维护性
- 完善的日志记录
- API 文档自动生成
- 单元测试覆盖率 > 70%
- 错误码规范化
- 数据库迁移版本管理

---

## 8. 附录

### 8.1 统一响应格式
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 8.2 错误码定义
| 错误码 | 描述 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未授权 |
| 1003 | 禁止访问 |
| 1004 | 资源不存在 |
| 2001 | 用户不存在 |
| 2002 | 用户已存在 |
| 2003 | 密码错误 |
| 3001 | 商家不存在 |
| 3002 | 商家未审核 |
| 3003 | 商家已禁用 |
| 3004 | 商家未开通配送 |
| 3005 | 超出配送范围 |
| 4001 | 商品不存在 |
| 4002 | 商品已下架 |
| 4003 | 库存不足 |
| 5001 | 订单不存在 |
| 5002 | 订单状态错误 |
| 5003 | 订单已支付 |
| 5004 | 订单已取消 |
| 5005 | 无权操作此订单（跨商家操作） |
| 5006 | 核销码错误 |
| 6001 | 支付失败 |
| 6002 | 退款失败 |
| 6003 | 进件申请失败 |
| 6004 | 进件审核中 |
| 7001 | 分类不存在 |
| 7002 | 分类下存在商品 |
| 8001 | 服务商配置错误 |
| 8002 | 七牛云上传失败 |
| 9001 | 服务器内部错误 |

### 8.3 订单状态枚举
| 状态 | 值 | 描述 |
|------|-----|------|
| pending_payment | 1 | 待支付 |
| paid | 2 | 已支付 |
| completed | 3 | 已完成 |
| cancelled | 4 | 已取消 |
| refunding | 5 | 退款中 |
| refunded | 6 | 已退款 |

### 8.4 分页参数
```json
{
  "page": 1,
  "page_size": 10,
  "total": 100,
  "total_pages": 10
}
```

### 8.5 微信支付回调通知
```json
{
  "id": "通知ID",
  "create_time": "通知创建时间",
  "resource_type": "encrypt-resource",
  "event_type": "TRANSACTION.SUCCESS",
  "summary": "支付成功",
  "resource": {
    "original_type": "transaction",
    "algorithm": "AEAD_AES_256_GCM",
    "ciphertext": "加密数据",
    "associated_data": "附加数据",
    "nonce": "随机串"
  }
}
```

---

## 9. 小程序测试指南

### 9.1 测试环境说明

在小程序正式发布前，由于微信小程序的限制，无法通过扫码直接进入特定商家的店铺页面进行测试。为此，我们提供了两种替代测试方法：

### 9.2 测试方法一：测试入口页面（推荐）

#### 功能说明
创建了专门的测试入口页面 `pages/store/test-entry.vue`，用于在小程序发布前测试用户下单流程。

#### 页面路径
- 路由：`/pages/store/test-entry`
- 导航标题：`商家店铺测试`
- 背景色：`#667eea`（紫色渐变）

#### 功能特性
1. **商家ID输入**：支持手动输入商家ID
2. **商家列表选择**：提供预设商家列表，点击即可快速选择
3. **直接跳转店铺**：输入商家ID后直接跳转至店铺首页

#### 使用步骤
1. 在微信开发者工具中打开小程序项目
2. 进入"商家店铺测试"页面
3. 在输入框中输入商家ID（默认：1）
4. 或直接点击商家列表中的商家
5. 点击"进入商家店铺"按钮
6. 进入店铺首页后可进行商品浏览、加入购物车、下单支付等操作

#### 测试商家列表
| 商家ID | 商家名称 | 状态 |
|--------|----------|------|
| 1 | 美味餐厅 | 营业中 |
| 2 | 示例商家B | 营业中 |
| 3 | 示例商家C | 营业中 |

#### 注意事项
- 支付功能需要在微信开发者工具中开启调试
- 测试支付可使用模拟支付模式
- 订单数据会真实写入数据库

### 9.3 测试方法二：编译模式直接进入

#### 功能说明
利用微信开发者工具的编译模式功能，直接编译到指定页面并携带参数。

#### 配置步骤
1. 打开微信开发者工具
2. 点击顶部"编译模式"下拉框
3. 选择"添加编译模式"
4. 配置以下参数：

| 参数 | 值 | 说明 |
|------|-----|------|
| 编译模式 | `pages/store/home` | 店铺首页 |
| 启动参数 | `merchant_id=1` | 商家ID参数 |

#### 启动参数说明
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| merchant_id | number | 是 | 商家ID，用于加载对应商家的店铺数据 |

#### 使用示例
```
启动参数：merchant_id=1
```
访问商家ID为1的店铺首页

```
启动参数：merchant_id=2
```
访问商家ID为2的店铺首页

### 9.4 测试功能清单

#### 可测试功能模块

| 模块 | 功能 | 测试方法一 | 测试方法二 |
|------|------|-----------|-----------|
| 店铺浏览 | 查看店铺信息、公告 | ✅ | ✅ |
| 商品展示 | 商品列表、分类筛选 | ✅ | ✅ |
| 商品详情 | 查看商品详情、规格选择 | ✅ | ✅ |
| 购物车 | 添加商品、修改数量、删除 | ✅ | ✅ |
| 下单流程 | 确认订单、选择配送方式 | ✅ | ✅ |
| 支付功能 | 微信支付（需开启调试） | ✅ | ✅ |
| 订单管理 | 查看订单列表、订单详情 | ✅ | ✅ |
| 退款申请 | 申请退款、查看退款状态 | ✅ | ✅ |

### 9.5 支付配置说明

#### 开发环境配置
1. 打开微信开发者工具
2. 进入"详情" → "本地设置"
3. 勾选"不校验合法域名"（开发阶段）
4. 勾选"不校验HTTPS证书"

#### 生产环境要求
- 需要配置已备案的域名
- 需要配置SSL证书（HTTPS）
- 需要在微信小程序后台添加服务器域名配置
- 需要配置微信支付相关参数

#### 微信支付测试
开发阶段测试微信支付：
1. 使用微信支付开发工具
2. 或配置测试商户号
3. 使用沙箱环境进行测试

### 9.6 调试技巧

#### 查看接口请求
1. 打开微信开发者工具
2. 切换到"Network"面板
3. 筛选接口请求
4. 查看请求参数和响应数据

#### 查看控制台日志
1. 在关键代码位置添加 `console.log`
2. 在"Console"面板查看日志输出
3. 检查接口返回的数据结构

#### 数据检查
- 使用后端接口直接查看数据库状态
- 使用 Postman 或 curl 测试 API 接口
- 检查订单状态是否正确更新
