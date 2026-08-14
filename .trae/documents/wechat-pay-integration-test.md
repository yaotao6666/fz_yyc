# 接入微信支付服务商真实密钥与测试商户计划

## 背景

用户已拥有微信支付服务商的真实资料和密钥，且已线下进件完成1个商户。需要将这些真实配置接入系统，使支付链路可以真实跑通测试。

## 当前状态

- 数据库中的服务商和商户数据均为 mock 值
- `.env` 文件中微信支付配置为空或 mock 值
- `docker-compose.yml` 未传递 `WECHAT_PAY_SP_*` 新版环境变量
- 支付回调地址需要公网可访问（微信支付服务器需回调通知）
- 小程序 AppID/AppSecret 需要配置（用于获取用户 OpenID 拉起支付）

## 实施步骤

### 步骤1：配置 `.env` 文件中的微信支付服务商密钥

修改 `server/.env` 文件，填入真实的服务商支付配置：

```env
# 微信支付服务商配置（优先使用 SP_ 前缀变量）
WECHAT_PAY_SP_MCH_ID=你的服务商商户号
WECHAT_PAY_SP_API_V3_KEY=你的APIv3密钥（32字节）
WECHAT_PAY_SP_CERT_SERIAL_NO=你的证书序列号
WECHAT_PAY_SP_PRIVATE_KEY=你的服务商私钥PEM内容（或文件路径）
WECHAT_PAY_SP_PUBLIC_KEY=微信支付平台公钥/证书PEM内容（或文件路径）
WECHAT_PAY_SP_CALLBACK_URL=你的公网回调地址（如 https://yourdomain.com/api/v1/notify/payment）

# 微信小程序配置
WECHAT_APP_ID=你的小程序AppID
WECHAT_APP_SECRET=你的小程序AppSecret
```

**关于私钥/公钥的填写方式**：
- 方式A：直接粘贴 PEM 内容（包含 `-----BEGIN ...-----` 和 `-----END ...-----`）
- 方式B：填写服务器上的文件路径（如 `/app/certs/apiclient_key.pem`）

### 步骤2：更新 `docker-compose.yml` 传递新版环境变量
  
```yaml
WECHAT_PAY_SP_MCH_ID: ${WECHAT_PAY_SP_MCH_ID:-}
WECHAT_PAY_SP_API_V3_KEY: ${WECHAT_PAY_SP_API_V3_KEY:-}
WECHAT_PAY_SP_CERT_SERIAL_NO: ${WECHAT_PAY_SP_CERT_SERIAL_NO:-}
WECHAT_PAY_SP_PRIVATE_KEY: ${WECHAT_PAY_SP_PRIVATE_KEY:-}
WECHAT_PAY_SP_PUBLIC_KEY: ${WECHAT_PAY_SP_PUBLIC_KEY:-}
WECHAT_PAY_SP_CALLBACK_URL: ${WECHAT_PAY_SP_CALLBACK_URL:-}
```

### 步骤3：更新数据库中的服务商支付信息

通过 SQL 更新 `service_providers` 表中的真实信息：

```sql
UPDATE service_providers SET
  mch_id = '你的服务商商户号',
  api_v3_key = '你的APIv3密钥',
  cert_serial_no = '你的证书序列号',
  private_key = '你的私钥PEM',
  public_key = '你的公钥/证书PEM',
  callback_url = '你的公网回调地址'
WHERE id = 1;
```

> 注意：`service_providers` 表中的字段与 `.env` 配置存在双重存储。当前支付逻辑使用的是 `.env` 中的配置（通过 `config.Config.WechatPay`），数据库中的字段主要用于 SP 设置页面展示。两边需要保持一致。

### 步骤4：更新已进件商户的子商户号

通过 SQL 或 SP 小程序商家编辑页面，回填已进件商户的 `sub_mch_id`：

**方式A：SQL 直接更新**
```sql
UPDATE merchants SET
  sub_mch_id = '你的子商户号',
  payment_config_status = 1
WHERE id = 你的商家ID;
```

**方式B：通过 SP 小程序操作**
1. 登录 SP 端（admin/admin123）
2. 进入商家列表 → 选择商家 → 编辑
3. 在“支付配置”区域填写子商户号
4. 保存

### 步骤5：配置公网回调地址（内网穿透）

微信支付回调需要公网可访问的 URL。开发环境需使用内网穿透工具：

**推荐方案**：使用 ngrok / frp / 花生壳等工具

```bash
# 示例：使用 ngrok
ngrok http 8080
# 获得公网地址如 https://xxxx.ngrok.io
# 回调地址配置为 https://xxxx.ngrok.io/api/v1/notify/payment
```

将公网地址填入：
1. `.env` 的 `WECHAT_PAY_SP_CALLBACK_URL`
2. 数据库 `service_providers.callback_url`

### 步骤6：重启服务并验证配置

```bash
# 重新构建并启动
cd server && docker compose up -d --build

# 验证配置是否加载
curl -s http://localhost:8080/api/v1/sp/settings -H "Authorization: Bearer <SP_TOKEN>"
```

### 步骤7：端到端支付测试

1. **C端用户登录**：在小程序中扫码进入店铺，获取 OpenID
2. **下单**：选择商品 → 确认订单 → 提交
3. **支付**：拉起微信支付 → 完成支付
4. **回调验证**：检查后端日志确认支付回调收到并处理
5. **商家端验证**：商家端订单列表查看订单状态变为“已支付”

## 需要用户确认的信息

| 信息项 | 说明 | 示例 |
|--------|------|------|
| 服务商商户号 | 微信支付服务商平台的商户号 | 14xxxxx |
| APIv3密钥 | 32字节字符串 | xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
| 证书序列号 | 服务商API证书序列号 | 5A7Bxxxxx |
| 服务商私钥 | PEM格式，可粘贴内容或提供文件路径 | -----BEGIN PRIVATE KEY----- ... |
| 微信支付平台公钥/证书 | PEM格式，用于验签回调 | -----BEGIN CERTIFICATE----- ... |
| 小程序AppID | 微信小程序的AppID | wx1234567890 |
| 小程序AppSecret | 微信小程序的AppSecret | xxxxxxxxxxxxxxxx |
| 子商户号 | 已进件商户的子商户号 | 15xxxxx |
| 公网回调地址 | 内网穿透后的公网URL | https://xxx.ngrok.io/api/v1/notify/payment |

## 涉及修改的文件

| 文件 | 修改内容 |
|------|----------|
| `server/.env` | 填入真实微信支付和小程序配置 |
| `server/docker-compose.yml` | 补充 `WECHAT_PAY_SP_*` 环境变量传递 |
| 数据库 `service_providers` | 更新服务商支付信息 |
| 数据库 `merchants` | 回填子商户号 |

## 注意事项

1. **密钥安全**：私钥和密钥属于敏感信息，不应提交到 Git。`.env` 文件已在 `.gitignore` 中排除
2. **双存储一致性**：`.env` 配置和数据库 `service_providers` 表都存储了支付信息，需保持一致
3. **回调地址必须HTTPS**：微信支付要求回调地址为 HTTPS
4. **小程序需关联服务商**：小程序的 AppID 需要在微信支付服务商平台关联
5. **测试金额**：微信支付沙箱环境已下线，测试使用真实金额，建议使用0.01元测试
