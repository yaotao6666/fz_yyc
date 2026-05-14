-- 服务商支付与分账重构迁移
SET FOREIGN_KEY_CHECKS = 0;

ALTER TABLE `merchants`
  ADD COLUMN IF NOT EXISTS `profit_sharing_enabled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启分账' AFTER `sub_mch_id`,
  ADD COLUMN IF NOT EXISTS `profit_sharing_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '分账比例' AFTER `profit_sharing_enabled`,
  ADD COLUMN IF NOT EXISTS `payment_config_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付配置状态' AFTER `profit_sharing_ratio`;

UPDATE `merchants`
SET
  `profit_sharing_enabled` = COALESCE(`profit_sharing_enabled`, 0),
  `profit_sharing_ratio` = COALESCE(`profit_sharing_ratio`, 0),
  `payment_config_status` = CASE
    WHEN COALESCE(`sub_mch_id`, '') <> '' AND (COALESCE(`profit_sharing_enabled`, 0) = 0 OR COALESCE(`profit_sharing_ratio`, 0) > 0) THEN 1
    ELSE 0
  END;

ALTER TABLE `orders`
  ADD COLUMN IF NOT EXISTS `pay_notify_payload` JSON DEFAULT NULL COMMENT '支付回调原始数据' AFTER `paid_at`,
  ADD COLUMN IF NOT EXISTS `profit_sharing_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分账状态' AFTER `refunded_at`,
  ADD COLUMN IF NOT EXISTS `profit_sharing_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '分账金额' AFTER `profit_sharing_status`,
  ADD COLUMN IF NOT EXISTS `profit_sharing_order_no` VARCHAR(64) DEFAULT NULL COMMENT '分账单号' AFTER `profit_sharing_amount`,
  ADD COLUMN IF NOT EXISTS `profit_sharing_at` DATETIME DEFAULT NULL COMMENT '分账时间' AFTER `profit_sharing_order_no`,
  ADD COLUMN IF NOT EXISTS `profit_sharing_error` VARCHAR(256) DEFAULT NULL COMMENT '分账错误信息' AFTER `profit_sharing_at`;

CREATE TABLE IF NOT EXISTS `merchant_profit_sharing_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `service_provider_id` BIGINT UNSIGNED NOT NULL,
  `merchant_id` BIGINT UNSIGNED NOT NULL,
  `order_id` BIGINT UNSIGNED NOT NULL,
  `order_no` VARCHAR(32) NOT NULL,
  `transaction_id` VARCHAR(64) DEFAULT NULL,
  `profit_sharing_order_no` VARCHAR(64) NOT NULL,
  `profit_sharing_date` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `pay_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  `profit_sharing_ratio` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `profit_sharing_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  `merchant_received_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  `status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
  `error_message` VARCHAR(256) DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_merchant_profit_sharing_records_sp_id` (`service_provider_id`),
  KEY `idx_merchant_profit_sharing_records_merchant_id` (`merchant_id`),
  KEY `idx_merchant_profit_sharing_records_order_id` (`order_id`),
  KEY `idx_merchant_profit_sharing_records_order_no` (`order_no`),
  KEY `idx_merchant_profit_sharing_records_status` (`status`),
  KEY `idx_merchant_profit_sharing_records_date` (`profit_sharing_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家分账记录表';

DROP TABLE IF EXISTS `merchant_applications`;
DROP TABLE IF EXISTS `merchant_audit_records`;
DROP TABLE IF EXISTS `merchant_licenses`;
DROP TABLE IF EXISTS `invite_records`;
DROP TABLE IF EXISTS `invite_rewards`;

ALTER TABLE `merchants`
  DROP COLUMN IF EXISTS `sub_mch_status`,
  DROP COLUMN IF EXISTS `applyment_status`,
  DROP COLUMN IF EXISTS `audit_status`,
  DROP COLUMN IF EXISTS `audit_remark`;

SET FOREIGN_KEY_CHECKS = 1;
