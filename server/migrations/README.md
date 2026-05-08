# 初始化数据使用说明

## 📋 数据概述

本初始化数据包含以下内容：

### 1. 服务商账号
- **服务商名称**: 寻梦服务商
- **联系人**: 张三
- **联系电话**: 13800138000
- **服务商商户号**: 1234567890

### 2. 服务商管理员账号
- **用户名**: `admin`
- **密码**: `admin123`
- **角色**: 超级管理员 (admin)
- **登录地址**: POST /api/v1/admin/auth/login

### 3. 商家账号
- **商家名称**: 美味餐厅
- **联系人**: 李四
- **联系电话**: 13900139000
- **经营类目**: 餐饮
- **商家地址**: 北京市朝阳区建国路88号
- **关联服务商**: 寻梦服务商 (ID=1)

### 4. 商家管理员账号
- **用户名**: `merchant`
- **密码**: `merchant123`
- **角色**: 商家管理员 (owner)
- **登录地址**: POST /api/v1/merchant/auth/login

## 🚀 使用步骤

### 1. 执行数据库迁移
```bash
# 确保已经执行了基础表结构迁移
mysql -u root -p fz_yyc_api < migrations/20240101000001_init.sql

# 执行初始化数据迁移
mysql -u root -p fz_yyc_api < migrations/20240101000002_seed_data.sql
```

或者使用 Go 程序自动执行：
```bash
# 编译验证程序
cd server
go build -o bin/verify_seed cmd/verify_seed/main.go

# 运行验证（会检查数据库连接和数据完整性）
./bin/verify_seed
```

### 2. 验证账号登录

#### 服务商管理员登录
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

#### 商家管理员登录
```bash
curl -X POST http://localhost:8080/api/v1/merchant/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"merchant","password":"merchant123"}'
```

## 📊 Mock 数据详情

### 商品分类 (5个)
1. 热销推荐
2. 招牌菜
3. 凉菜
4. 主食
5. 饮品

### 商品 (10个)
- 招牌红烧肉 - ¥58.00
- 宫保鸡丁 - ¥38.00
- 糖醋里脊 - ¥42.00
- 水煮鱼 - ¥88.00
- 凉拌黄瓜 - ¥18.00
- 凉拌木耳 - ¥22.00
- 米饭 - ¥3.00
- 馒头 - ¥2.00
- 可乐 - ¥5.00
- 鲜榨橙汁 - ¥12.00

### C端用户 (5个)
1. 小明 - 13811112222
2. 小红 - 13811113333
3. 小张 - 13811114444
4. 小李 - 13811115555
5. 小王 - 13811116666

### 订单 (5个)
- 待支付订单: 1个
- 已支付订单: 3个
- 已完成订单: 1个

### 平台活动 (3个)
1. 新商家入驻优惠 (Banner)
2. 春节特惠活动 (Banner)
3. 平台升级通知 (公告)

## ⚠️ 注意事项

1. **仅用于开发测试**: 此数据仅用于开发和测试环境
2. **密码安全**: 所有密码均为测试密码，请勿在生产环境使用
3. **数据独立性**: 商家关联到 ID=1 的服务商，如删除该服务商，商家数据也会被删除
4. **外键约束**: 订单数据依赖用户和商家，删除用户或商家可能导致订单查询异常

## 🔧 自定义数据

如需修改初始数据，请编辑 `migrations/20240101000002_seed_data.sql` 文件。

如需重新生成密码哈希：
```bash
cd server/cmd/gen_password
go run main.go
```

## 📞 技术支持

如有问题，请查看：
- 数据库迁移日志
- API 服务日志
- MySQL 错误日志
