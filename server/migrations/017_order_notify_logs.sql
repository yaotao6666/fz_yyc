-- 017_order_notify_logs.sql
-- 订单通知日志表：服务人员接单等订单类通知的发送留痕（幂等去重依据），channel 区分小程序订阅消息/短信
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS order_notify_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '接收用户ID',
  channel VARCHAR(16) NOT NULL DEFAULT 'wechat' COMMENT '发送渠道(wechat/sms)',
  content VARCHAR(255) NOT NULL DEFAULT '' COMMENT '通知内容',
  success TINYINT NOT NULL DEFAULT 0 COMMENT '是否成功: 0=失败 1=成功',
  message VARCHAR(255) NOT NULL DEFAULT '' COMMENT '结果说明',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  KEY idx_onn_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单通知日志表';