-- 清理分账相关表与字段
-- 删除分账记录表
DROP TABLE IF EXISTS merchant_profit_sharing_records;

-- 删除 orders 表分账字段
ALTER TABLE orders DROP COLUMN IF EXISTS profit_sharing_status;
ALTER TABLE orders DROP COLUMN IF EXISTS profit_sharing_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS profit_sharing_order_no;
ALTER TABLE orders DROP COLUMN IF EXISTS profit_sharing_at;
ALTER TABLE orders DROP COLUMN IF EXISTS profit_sharing_error;

-- 删除 merchants 表分账字段（保留 sub_mch_id、payment_config_status）
ALTER TABLE merchants DROP COLUMN IF EXISTS profit_sharing_enabled;
ALTER TABLE merchants DROP COLUMN IF EXISTS profit_sharing_ratio;
