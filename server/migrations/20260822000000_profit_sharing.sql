SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 服务商分账功能：分账接收方 / 分账单 / 分账明细 + 自动分账开关 + 分账菜单权限
-- 说明：固定 sub_app 模式，分账 appid 取特约商户主体小程序 SubAppID
-- 说明：追加列一律使用 information_schema + PREPARE 动态判断，
--       兼容不支持 `ADD COLUMN IF NOT EXISTS` 的 MySQL 版本。
-- ============================================================

-- ============================================================
-- 1. merchants 新增自动分账总开关
-- ============================================================
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND COLUMN_NAME='profit_sharing_enabled');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE `merchants` ADD COLUMN `profit_sharing_enabled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启自动分账' AFTER `payment_config_status`");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. orders 新增分账字段（0未分账 1分账中 2分账成功 3分账失败 4跳过）
-- ============================================================
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='profit_sharing_status');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE `orders` ADD COLUMN `profit_sharing_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分账状态:0未分账1分账中2分账成功3分账失败4已跳过' AFTER `pay_notify_payload`");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='profit_sharing_amount');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE `orders` ADD COLUMN `profit_sharing_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '分账总额(元)' AFTER `profit_sharing_status`");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='profit_sharing_order_no');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE `orders` ADD COLUMN `profit_sharing_order_no` VARCHAR(64) DEFAULT NULL COMMENT '微信分账单号' AFTER `profit_sharing_amount`");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='profit_sharing_at');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE `orders` ADD COLUMN `profit_sharing_at` DATETIME DEFAULT NULL COMMENT '分账完成时间' AFTER `profit_sharing_order_no`");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='profit_sharing_error');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE `orders` ADD COLUMN `profit_sharing_error` VARCHAR(512) DEFAULT NULL COMMENT '分账失败原因' AFTER `profit_sharing_at`");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 3. 分账接收方表
-- ============================================================
CREATE TABLE IF NOT EXISTS `profit_sharing_receivers` (
  `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `merchant_id`    BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '商家ID(单商户恒为1)',
  `receiver_type`  TINYINT UNSIGNED NOT NULL COMMENT '接收方类型: 1=商户号 2=个人微信openid',
  `name`           VARCHAR(64)  NOT NULL COMMENT '显示名称',
  `account`        VARCHAR(64)  NOT NULL COMMENT '接收方账号(商户号 或 个人openid)',
  `personal_name`  VARCHAR(64)  DEFAULT NULL COMMENT '个人真实姓名(个人类型时, 微信实名校验, 可空)',
  `relation_type`  VARCHAR(32)  NOT NULL DEFAULT 'SERVICE_PROVIDER' COMMENT '与特约商户关系, 默认SERVICE_PROVIDER',
  `default_ratio`  DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '自动分账默认比例(%)',
  `wechat_bound`   TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '微信接收方关系是否已建立: 0=未建立 1=已建立',
  `wechat_error`   VARCHAR(256) DEFAULT NULL COMMENT '微信建立接收方关系失败原因',
  `status`         TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=停用',
  `sort`           INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序值(越小越靠前)',
  `remark`         VARCHAR(256) DEFAULT NULL COMMENT '备注',
  `created_at`     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_profit_sharing_receivers_merchant_id` (`merchant_id`),
  KEY `idx_profit_sharing_receivers_account` (`account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账接收方表';

-- ============================================================
-- 4. 分账单表（一个支付单对应一个分账单）
-- ============================================================
CREATE TABLE IF NOT EXISTS `profit_sharing_records` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `merchant_id`       BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '商家ID(单商户恒为1)',
  `order_id`          BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `order_no`          VARCHAR(32)     NOT NULL COMMENT '订单编号',
  `sp_mchid`          VARCHAR(32)     DEFAULT NULL COMMENT '服务商商户号',
  `sub_mchid`         VARCHAR(32)     DEFAULT NULL COMMENT '特约商户号(分账出资方)',
  `appid`             VARCHAR(64)     DEFAULT NULL COMMENT '分账请求使用的appid(特约商户主体小程序)',
  `transaction_id`    VARCHAR(64)     DEFAULT NULL COMMENT '微信支付交易单号',
  `out_order_no`      VARCHAR(64)     NOT NULL COMMENT '微信分账单号',
  `total_amount`      DECIMAL(10,2)   NOT NULL DEFAULT 0.00 COMMENT '订单实付金额(元)',
  `total_share_amount` DECIMAL(10,2)  NOT NULL DEFAULT 0.00 COMMENT '本次分账总额(元)',
  `status`            TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 0=待分账 1=分账中 2=分账成功 3=分账失败 4=已跳过',
  `share_time`        DATETIME DEFAULT NULL COMMENT '分账完成时间',
  `error_message`     VARCHAR(512) DEFAULT NULL COMMENT '失败原因',
  `created_at`        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_profit_sharing_records_merchant_id` (`merchant_id`),
  KEY `idx_profit_sharing_records_order_id` (`order_id`),
  KEY `idx_profit_sharing_records_order_no` (`order_no`),
  KEY `idx_profit_sharing_records_out_order_no` (`out_order_no`),
  KEY `idx_profit_sharing_records_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账单表';

-- ============================================================
-- 5. 分账单明细表（一个分账单内的各方）
-- ============================================================
CREATE TABLE IF NOT EXISTS `profit_sharing_record_receivers` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `record_id`       BIGINT UNSIGNED NOT NULL COMMENT '分账单ID',
  `receiver_id`     BIGINT UNSIGNED DEFAULT NULL COMMENT '接收方ID',
  `receiver_type`   TINYINT UNSIGNED NOT NULL COMMENT '接收方类型快照: 1=商户号 2=个人微信openid',
  `receiver_name`   VARCHAR(64)     NOT NULL COMMENT '接收方名称快照',
  `account`         VARCHAR(64)     NOT NULL COMMENT '接收方账号快照',
  `amount`          DECIMAL(10,2)   NOT NULL DEFAULT 0.00 COMMENT '分账金额(元)',
  `result_status`   VARCHAR(32)     DEFAULT NULL COMMENT '微信分账结果: PROCESSING/SUCCESS/CLOSED/FAILED/FINISHED',
  `detail_id`       VARCHAR(64)     DEFAULT NULL COMMENT '微信分账明细单号',
  `fail_reason`     VARCHAR(128)    DEFAULT NULL COMMENT '分账失败原因',
  `finish_time`     DATETIME DEFAULT NULL COMMENT '分账完成时间',
  `created_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_profit_sharing_record_receivers_record_id` (`record_id`),
  KEY `idx_profit_sharing_record_receivers_receiver_id` (`receiver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账单明细表';

-- ============================================================
-- 6. 分账菜单权限种子（admin 角色一键绑定）
-- 注意：菜单 id 需避开已占用号段(1-8 顶级 / 81-88 等)，此处使用 9/91/92 号段
-- ============================================================
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(9, 0, 1, '分账管理', '/settlement', 'Money', 9, 1, 1, NULL),
(91, 9, 1, '分账接收方', '/settlement/receivers', NULL, 1, 1, 1, 'profit:view'),
(92, 9, 1, '分账记录', '/settlement/records', NULL, 2, 1, 1, 'profit:view'),
(911, 91, 2, '新增/编辑/删除/同步接收方', NULL, NULL, 1, 1, 1, 'profit:receiver:manage'),
(921, 92, 2, '分账/重试', NULL, NULL, 1, 1, 1, 'profit:share'),
(922, 92, 2, '分账配置', NULL, NULL, 2, 1, 1, 'profit:config');

INSERT IGNORE INTO sys_role_menus (role_id, menu_id)
SELECT 1, id FROM sys_menus WHERE id IN (9, 91, 92, 911, 921, 922);

SET FOREIGN_KEY_CHECKS = 1;