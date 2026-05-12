-- ============================================
-- 寻梦私域管家 - 数据库初始化脚本
-- 数据库: MySQL 8.0
-- 版本: v2.0
-- 创建时间: 2024-01-01
-- 最后更新: 2026-05-11
-- 说明: 单文件初始化脚本，结构以当前代码模型为准，仅保留 PRD 中约定的服务商/商家登录账号
-- ============================================

CREATE DATABASE IF NOT EXISTS `fz_yyc_api` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `fz_yyc_api`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `print_logs`;
DROP TABLE IF EXISTS `cloud_printers`;
DROP TABLE IF EXISTS `coupon_records`;
DROP TABLE IF EXISTS `coupons`;
DROP TABLE IF EXISTS `user_addresses`;
DROP TABLE IF EXISTS `merchant_rates`;
DROP TABLE IF EXISTS `merchant_fees`;
DROP TABLE IF EXISTS `merchant_audit_records`;
DROP TABLE IF EXISTS `announcements`;
DROP TABLE IF EXISTS `invite_rewards`;
DROP TABLE IF EXISTS `activities`;
DROP TABLE IF EXISTS `invite_records`;
DROP TABLE IF EXISTS `refunds`;
DROP TABLE IF EXISTS `order_items`;
DROP TABLE IF EXISTS `orders`;
DROP TABLE IF EXISTS `user_behavior_events`;
DROP TABLE IF EXISTS `user_visits`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `product_specs`;
DROP TABLE IF EXISTS `products`;
DROP TABLE IF EXISTS `categories`;
DROP TABLE IF EXISTS `merchant_staffs`;
DROP TABLE IF EXISTS `merchant_licenses`;
DROP TABLE IF EXISTS `merchant_delivery_settings`;
DROP TABLE IF EXISTS `merchant_applications`;
DROP TABLE IF EXISTS `merchants`;
DROP TABLE IF EXISTS `service_provider_admins`;
DROP TABLE IF EXISTS `service_providers`;

-- 1. 服务商表
CREATE TABLE `service_providers` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(128) NOT NULL,
    `contact_name` VARCHAR(64) DEFAULT NULL,
    `contact_phone` VARCHAR(20) DEFAULT NULL,
    `mch_id` VARCHAR(32) DEFAULT NULL,
    `api_key` VARCHAR(128) DEFAULT NULL,
    `api_v3_key` VARCHAR(128) DEFAULT NULL,
    `cert_serial_no` VARCHAR(64) DEFAULT NULL,
    `private_key` TEXT DEFAULT NULL,
    `public_key` TEXT DEFAULT NULL,
    `callback_url` VARCHAR(256) DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_service_providers_mch_id` (`mch_id`),
    KEY `idx_service_providers_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务商表';

-- 2. 服务商管理员表
CREATE TABLE `service_provider_admins` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `service_provider_id` BIGINT UNSIGNED NOT NULL,
    `username` VARCHAR(64) NOT NULL,
    `password` VARCHAR(128) NOT NULL,
    `name` VARCHAR(64) DEFAULT NULL,
    `phone` VARCHAR(20) DEFAULT NULL,
    `role` VARCHAR(32) NOT NULL DEFAULT 'operator',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `last_login_at` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_service_provider_admins_username` (`username`),
    KEY `idx_service_provider_admins_sp_id` (`service_provider_id`),
    CONSTRAINT `fk_service_provider_admins_sp` FOREIGN KEY (`service_provider_id`) REFERENCES `service_providers` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务商管理员表';

-- 3. 商家表
CREATE TABLE `merchants` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `service_provider_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(128) NOT NULL,
    `logo` VARCHAR(512) DEFAULT NULL,
    `contact_name` VARCHAR(64) DEFAULT NULL,
    `contact_phone` VARCHAR(20) DEFAULT NULL,
    `contact_email` VARCHAR(128) DEFAULT NULL,
    `address` VARCHAR(256) DEFAULT NULL,
    `lat` DECIMAL(10,6) DEFAULT NULL,
    `lng` DECIMAL(10,6) DEFAULT NULL,
    `business_category` VARCHAR(64) DEFAULT NULL,
    `business_hours` VARCHAR(64) DEFAULT NULL,
    `announcement` TEXT DEFAULT NULL,
    `min_order_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `takeout_enabled` TINYINT(1) NOT NULL DEFAULT 1,
    `dine_in_enabled` TINYINT(1) NOT NULL DEFAULT 1,
    `sub_mch_id` VARCHAR(32) DEFAULT NULL,
    `sub_mch_status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
    `applyment_status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
    `audit_status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
    `audit_remark` VARCHAR(256) DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `rating` DECIMAL(2,1) NOT NULL DEFAULT 5.0,
    `sales_count` INT UNSIGNED NOT NULL DEFAULT 0,
    `qrcode_url` VARCHAR(512) DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_merchants_sp_id` (`service_provider_id`),
    KEY `idx_merchants_sub_mch_id` (`sub_mch_id`),
    KEY `idx_merchants_audit_status` (`audit_status`),
    KEY `idx_merchants_status` (`status`),
    KEY `idx_merchants_location` (`lat`, `lng`),
    CONSTRAINT `fk_merchants_sp` FOREIGN KEY (`service_provider_id`) REFERENCES `service_providers` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家表';

-- 4. 商家进件申请表
CREATE TABLE `merchant_applications` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `merchant_name` VARCHAR(128) NOT NULL,
    `business_license_info` JSON DEFAULT NULL,
    `legal_person_info` JSON DEFAULT NULL,
    `bank_account_info` JSON DEFAULT NULL,
    `store_info` JSON DEFAULT NULL,
    `contact_info` JSON DEFAULT NULL,
    `applyment_id` VARCHAR(64) DEFAULT NULL,
    `sub_mch_id` VARCHAR(32) DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
    `audit_detail` JSON DEFAULT NULL,
    `submit_time` DATETIME DEFAULT NULL,
    `audit_time` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_merchant_applications_merchant_id` (`merchant_id`),
    KEY `idx_merchant_applications_applyment_id` (`applyment_id`),
    KEY `idx_merchant_applications_sub_mch_id` (`sub_mch_id`),
    KEY `idx_merchant_applications_status` (`status`),
    CONSTRAINT `fk_merchant_applications_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家进件申请表';

-- 5. 商家配送设置表
CREATE TABLE `merchant_delivery_settings` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `enabled` TINYINT(1) NOT NULL DEFAULT 1,
    `base_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `free_delivery_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `max_distance` INT UNSIGNED NOT NULL DEFAULT 10,
    `distance_rules` JSON DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_merchant_delivery_settings_merchant_id` (`merchant_id`),
    CONSTRAINT `fk_merchant_delivery_settings_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家配送设置表';

-- 6. 商家营业执照表
CREATE TABLE `merchant_licenses` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `license_no` VARCHAR(64) DEFAULT NULL,
    `license_name` VARCHAR(128) DEFAULT NULL,
    `license_image` VARCHAR(512) DEFAULT NULL,
    `legal_person` VARCHAR(64) DEFAULT NULL,
    `legal_person_id` VARCHAR(32) DEFAULT NULL,
    `legal_person_id_front` VARCHAR(512) DEFAULT NULL,
    `legal_person_id_back` VARCHAR(512) DEFAULT NULL,
    `valid_from` DATE DEFAULT NULL,
    `valid_to` DATE DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_merchant_licenses_merchant_id` (`merchant_id`),
    CONSTRAINT `fk_merchant_licenses_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家营业执照表';

-- 7. 商家员工表
CREATE TABLE `merchant_staffs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `username` VARCHAR(64) NOT NULL,
    `password` VARCHAR(128) NOT NULL,
    `name` VARCHAR(64) DEFAULT NULL,
    `phone` VARCHAR(20) DEFAULT NULL,
    `openid` VARCHAR(64) DEFAULT NULL,
    `unionid` VARCHAR(64) DEFAULT NULL,
    `wechat_bound_at` DATETIME DEFAULT NULL,
    `role` VARCHAR(32) NOT NULL DEFAULT 'staff',
    `notify_enabled` TINYINT(1) NOT NULL DEFAULT 1,
    `browse_notify_enabled` TINYINT(1) NOT NULL DEFAULT 1,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `last_login_at` DATETIME DEFAULT NULL,
    `last_wechat_login_at` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_merchant_staffs_merchant_username` (`merchant_id`, `username`),
    KEY `idx_merchant_staffs_openid` (`openid`),
    KEY `idx_merchant_staffs_phone` (`phone`),
    CONSTRAINT `fk_merchant_staffs_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家员工表';

-- 8. 商品分类表
CREATE TABLE `categories` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(64) NOT NULL,
    `sort` INT UNSIGNED NOT NULL DEFAULT 0,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_categories_merchant_id` (`merchant_id`),
    KEY `idx_categories_sort` (`merchant_id`, `sort`),
    CONSTRAINT `fk_categories_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品分类表';

-- 9. 商品表
CREATE TABLE `products` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `category_id` BIGINT UNSIGNED DEFAULT NULL,
    `name` VARCHAR(128) NOT NULL,
    `description` TEXT DEFAULT NULL,
    `images` JSON DEFAULT NULL,
    `price` DECIMAL(10,2) NOT NULL,
    `original_price` DECIMAL(10,2) DEFAULT NULL,
    `stock` INT UNSIGNED NOT NULL DEFAULT 0,
    `unit` VARCHAR(16) NOT NULL DEFAULT '份',
    `sales` INT UNSIGNED NOT NULL DEFAULT 0,
    `sort` INT UNSIGNED NOT NULL DEFAULT 0,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `deleted_at` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_products_merchant_id` (`merchant_id`),
    KEY `idx_products_category_id` (`category_id`),
    KEY `idx_products_status` (`status`),
    KEY `idx_products_sales` (`merchant_id`, `sales`),
    KEY `idx_products_sort` (`merchant_id`, `sort`),
    CONSTRAINT `fk_products_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';

-- 10. 商品规格表
CREATE TABLE `product_specs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `product_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(64) NOT NULL,
    `options` JSON DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_product_specs_product_id` (`product_id`),
    CONSTRAINT `fk_product_specs_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品规格表';

-- 11. C端用户表
CREATE TABLE `users` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `openid` VARCHAR(64) DEFAULT NULL,
    `union_id` VARCHAR(64) DEFAULT NULL,
    `nickname` VARCHAR(64) DEFAULT NULL,
    `avatar` VARCHAR(512) DEFAULT NULL,
    `phone` VARCHAR(20) DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `first_visit_at` DATETIME DEFAULT NULL,
    `last_visit_at` DATETIME DEFAULT NULL,
    `visit_count` INT UNSIGNED NOT NULL DEFAULT 1,
    `has_ordered` TINYINT(1) NOT NULL DEFAULT 0,
    `total_orders` INT UNSIGNED NOT NULL DEFAULT 0,
    `total_spent` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `has_paid` TINYINT(1) NOT NULL DEFAULT 0,
    `first_paid_at` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_users_openid` (`openid`),
    KEY `idx_users_union_id` (`union_id`),
    KEY `idx_users_phone` (`phone`),
    KEY `idx_users_first_visit_at` (`first_visit_at`),
    KEY `idx_users_has_paid` (`has_paid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='C端用户表';

-- 12. 用户收货地址表
CREATE TABLE `user_addresses` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(64) NOT NULL,
    `phone` VARCHAR(20) NOT NULL,
    `province` VARCHAR(32) DEFAULT NULL,
    `city` VARCHAR(32) DEFAULT NULL,
    `district` VARCHAR(32) DEFAULT NULL,
    `address` VARCHAR(256) NOT NULL,
    `lat` DECIMAL(10,6) DEFAULT NULL,
    `lng` DECIMAL(10,6) DEFAULT NULL,
    `is_default` TINYINT(1) NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_addresses_user_id` (`user_id`),
    CONSTRAINT `fk_user_addresses_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户收货地址表';

-- 13. 用户访问记录表
CREATE TABLE `user_visits` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `openid` VARCHAR(64) DEFAULT NULL,
    `visit_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `source` VARCHAR(32) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_user_visits_user_id` (`user_id`),
    KEY `idx_user_visits_merchant_id` (`merchant_id`),
    KEY `idx_user_visits_openid` (`openid`),
    CONSTRAINT `fk_user_visits_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_visits_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户访问记录表';

-- 14. 用户行为事件表
CREATE TABLE `user_behavior_events` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `openid` VARCHAR(64) DEFAULT NULL,
    `event_type` VARCHAR(32) NOT NULL,
    `page` VARCHAR(64) DEFAULT NULL,
    `product_id` BIGINT UNSIGNED DEFAULT NULL,
    `order_id` BIGINT UNSIGNED DEFAULT NULL,
    `source` VARCHAR(32) DEFAULT NULL,
    `payload` JSON DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_behavior_events_merchant_id` (`merchant_id`),
    KEY `idx_user_behavior_events_user_id` (`user_id`),
    KEY `idx_user_behavior_events_openid` (`openid`),
    KEY `idx_user_behavior_events_event_type` (`event_type`),
    KEY `idx_user_behavior_events_product_id` (`product_id`),
    KEY `idx_user_behavior_events_order_id` (`order_id`),
    KEY `idx_user_behavior_events_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户行为事件表';

-- 15. 订单表
CREATE TABLE `orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `order_no` VARCHAR(32) NOT NULL,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `total_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `delivery_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `discount_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `pay_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `delivery_type` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `delivery_distance` DECIMAL(5,2) DEFAULT NULL,
    `delivery_address` VARCHAR(256) DEFAULT NULL,
    `contact_name` VARCHAR(64) DEFAULT NULL,
    `contact_phone` VARCHAR(20) DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `remark` VARCHAR(256) DEFAULT NULL,
    `verify_code` VARCHAR(16) DEFAULT NULL,
    `transaction_id` VARCHAR(64) DEFAULT NULL,
    `paid_at` DATETIME DEFAULT NULL,
    `completed_at` DATETIME DEFAULT NULL,
    `completed_by_name` VARCHAR(64) DEFAULT NULL,
    `cancelled_at` DATETIME DEFAULT NULL,
    `refunded_at` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_orders_order_no` (`order_no`),
    KEY `idx_orders_user_id` (`user_id`),
    KEY `idx_orders_merchant_id` (`merchant_id`),
    KEY `idx_orders_status` (`status`),
    KEY `idx_orders_created_at` (`created_at`),
    KEY `idx_orders_paid_at` (`paid_at`),
    KEY `idx_orders_verify_code` (`verify_code`),
    CONSTRAINT `fk_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_orders_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 16. 订单商品表
CREATE TABLE `order_items` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `order_id` BIGINT UNSIGNED NOT NULL,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `product_id` BIGINT UNSIGNED NOT NULL,
    `product_name` VARCHAR(128) NOT NULL,
    `image` VARCHAR(512) DEFAULT NULL,
    `price` DECIMAL(10,2) NOT NULL,
    `quantity` INT UNSIGNED NOT NULL DEFAULT 1,
    `spec_info` JSON DEFAULT NULL,
    `subtotal` DECIMAL(10,2) NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_order_items_order_id` (`order_id`),
    KEY `idx_order_items_merchant_id` (`merchant_id`),
    KEY `idx_order_items_product_id` (`product_id`),
    CONSTRAINT `fk_order_items_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_order_items_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_order_items_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品表';

-- 17. 退款记录表
CREATE TABLE `refunds` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `order_id` BIGINT UNSIGNED NOT NULL,
    `refund_no` VARCHAR(32) NOT NULL,
    `refund_amount` DECIMAL(10,2) NOT NULL,
    `refund_reason` VARCHAR(256) DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
    `refund_id` VARCHAR(64) DEFAULT NULL,
    `refunded_at` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_refunds_refund_no` (`refund_no`),
    KEY `idx_refunds_order_id` (`order_id`),
    KEY `idx_refunds_status` (`status`),
    CONSTRAINT `fk_refunds_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款记录表';

-- 18. 邀请记录表
CREATE TABLE `invite_records` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `inviter_id` BIGINT UNSIGNED NOT NULL,
    `invitee_id` BIGINT UNSIGNED DEFAULT NULL,
    `invite_code` VARCHAR(32) DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
    `reward_type` VARCHAR(32) DEFAULT NULL,
    `reward_status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `completed_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_invite_records_inviter_id` (`inviter_id`),
    KEY `idx_invite_records_invitee_id` (`invitee_id`),
    KEY `idx_invite_records_invite_code` (`invite_code`),
    CONSTRAINT `fk_invite_records_inviter` FOREIGN KEY (`inviter_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_invite_records_invitee` FOREIGN KEY (`invitee_id`) REFERENCES `merchants` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请记录表';

-- 19. 平台活动表
CREATE TABLE `activities` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `type` VARCHAR(16) NOT NULL,
    `title` VARCHAR(128) DEFAULT NULL,
    `content` TEXT DEFAULT NULL,
    `image` VARCHAR(512) DEFAULT NULL,
    `link_type` VARCHAR(16) DEFAULT NULL,
    `link_value` VARCHAR(256) DEFAULT NULL,
    `sort` INT UNSIGNED NOT NULL DEFAULT 0,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `start_time` DATETIME DEFAULT NULL,
    `end_time` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_activities_type` (`type`),
    KEY `idx_activities_status` (`status`),
    KEY `idx_activities_sort` (`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台活动表';

-- 20. 邀请奖励规则表
CREATE TABLE `invite_rewards` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `type` VARCHAR(32) NOT NULL,
    `condition` VARCHAR(32) NOT NULL,
    `description` VARCHAR(256) DEFAULT NULL,
    `enabled` TINYINT(1) NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请奖励规则表';

-- 21. 系统公告表
CREATE TABLE `announcements` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `service_provider_id` BIGINT UNSIGNED NOT NULL,
    `title` VARCHAR(128) NOT NULL,
    `content` TEXT DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_announcements_sp_id` (`service_provider_id`),
    KEY `idx_announcements_status` (`status`),
    CONSTRAINT `fk_announcements_sp` FOREIGN KEY (`service_provider_id`) REFERENCES `service_providers` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统公告表';

-- 22. 商家审核记录表
CREATE TABLE `merchant_audit_records` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `auditor_id` BIGINT UNSIGNED NOT NULL,
    `action` VARCHAR(32) NOT NULL,
    `before_status` TINYINT UNSIGNED NOT NULL,
    `after_status` TINYINT UNSIGNED NOT NULL,
    `remark` VARCHAR(256) DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_merchant_audit_records_merchant_id` (`merchant_id`),
    KEY `idx_merchant_audit_records_auditor_id` (`auditor_id`),
    CONSTRAINT `fk_merchant_audit_records_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家审核记录表';

-- 23. 商家年费表
CREATE TABLE `merchant_fees` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `year` INT UNSIGNED NOT NULL,
    `amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `status` VARCHAR(16) NOT NULL,
    `pay_time` DATETIME DEFAULT NULL,
    `free_reason` VARCHAR(256) DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_merchant_fees_merchant_id` (`merchant_id`),
    CONSTRAINT `fk_merchant_fees_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家年费表';

-- 24. 商家手续费率表
CREATE TABLE `merchant_rates` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `rate_type` VARCHAR(32) NOT NULL,
    `rate` DECIMAL(5,4) NOT NULL,
    `effective_time` DATETIME NOT NULL,
    `expire_time` DATETIME DEFAULT NULL,
    `remark` VARCHAR(256) DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_merchant_rates_merchant_id` (`merchant_id`),
    KEY `idx_merchant_rates_status` (`status`),
    CONSTRAINT `fk_merchant_rates_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家手续费率表';

-- 25. 优惠券表
CREATE TABLE `coupons` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(64) NOT NULL,
    `type` VARCHAR(16) NOT NULL,
    `discount_amount` DECIMAL(10,2) DEFAULT NULL,
    `min_order_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `total_count` INT NOT NULL,
    `remaining_count` INT NOT NULL,
    `per_user_limit` INT NOT NULL DEFAULT 1,
    `start_time` DATETIME NOT NULL,
    `end_time` DATETIME NOT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_coupons_merchant_id` (`merchant_id`),
    KEY `idx_coupons_status` (`status`),
    CONSTRAINT `fk_coupons_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券表';

-- 26. 优惠券领取记录表
CREATE TABLE `coupon_records` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `coupon_id` BIGINT UNSIGNED NOT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
    `used_at` DATETIME DEFAULT NULL,
    `order_id` BIGINT UNSIGNED DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_coupon_records_user_id` (`user_id`),
    KEY `idx_coupon_records_coupon_id` (`coupon_id`),
    KEY `idx_coupon_records_order_id` (`order_id`),
    CONSTRAINT `fk_coupon_records_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_coupon_records_coupon` FOREIGN KEY (`coupon_id`) REFERENCES `coupons` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券领取记录表';

-- 27. 云打印机表
CREATE TABLE `cloud_printers` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(64) NOT NULL,
    `brand` VARCHAR(32) DEFAULT NULL,
    `device_no` VARCHAR(64) NOT NULL,
    `api_key` VARCHAR(64) DEFAULT NULL,
    `print_types` JSON DEFAULT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `last_print_at` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_cloud_printers_merchant_id` (`merchant_id`),
    CONSTRAINT `fk_cloud_printers_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云打印机表';

-- 28. 打印记录表
CREATE TABLE `print_logs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `printer_id` BIGINT UNSIGNED NOT NULL,
    `order_id` BIGINT UNSIGNED DEFAULT NULL,
    `type` VARCHAR(16) NOT NULL,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0,
    `error_message` VARCHAR(256) DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_print_logs_merchant_id` (`merchant_id`),
    KEY `idx_print_logs_printer_id` (`printer_id`),
    KEY `idx_print_logs_order_id` (`order_id`),
    CONSTRAINT `fk_print_logs_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_print_logs_printer` FOREIGN KEY (`printer_id`) REFERENCES `cloud_printers` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打印记录表';

-- ============================================
-- 初始化数据（仅保留 PRD 中约定的服务商/商家账号及最小联调数据）
-- ============================================

INSERT INTO `service_providers` (
    `id`, `name`, `contact_name`, `contact_phone`, `mch_id`, `api_key`, `api_v3_key`,
    `cert_serial_no`, `private_key`, `public_key`, `callback_url`, `status`, `created_at`, `updated_at`
) VALUES (
    1, '寻梦服务商', '张三', '13800138000', '1234567890', 'mock_api_key_12345',
    'mock_api_v3_key_1234567890123456789012', 'mock_cert_serial_no', 'mock_private_key',
    'mock_public_key', 'https://example.com/notify/payment', 1, NOW(), NOW()
);

INSERT INTO `service_provider_admins` (
    `id`, `service_provider_id`, `username`, `password`, `name`, `phone`, `role`, `status`, `last_login_at`, `created_at`, `updated_at`
) VALUES (
    1, 1, 'admin', '$2a$10$GCgIm2gqB7yPpx/w.pEVDeBj.xUzjOGFmsZdAZh2xLDr4RgYj2z6e',
    '超级管理员', '13800138000', 'admin', 1, NULL, NOW(), NOW()
);

INSERT INTO `merchants` (
    `id`, `service_provider_id`, `name`, `logo`, `contact_name`, `contact_phone`, `contact_email`, `address`,
    `lat`, `lng`, `business_category`, `business_hours`, `announcement`, `min_order_amount`, `takeout_enabled`,
    `dine_in_enabled`, `sub_mch_id`, `sub_mch_status`, `applyment_status`, `audit_status`, `audit_remark`,
    `status`, `rating`, `sales_count`, `qrcode_url`, `created_at`, `updated_at`
) VALUES (
    1, 1, '美味餐厅', 'https://example.com/images/merchant_logo_1.jpg', '李四', '13900139000',
    'lisi@example.com', '北京市朝阳区建国路88号', 39.908823, 116.407470, '餐饮', '09:00-22:00',
    '欢迎光临！', 20.00, 1, 1, '1500000001', 2, 2, 1, '资质审核通过', 1, 5.0, 0,
    'https://example.com/qrcode/merchant_1.png', NOW(), NOW()
);

INSERT INTO `merchant_staffs` (
    `id`, `merchant_id`, `username`, `password`, `name`, `phone`, `openid`, `unionid`, `wechat_bound_at`, `role`,
    `notify_enabled`, `browse_notify_enabled`, `status`, `last_login_at`, `last_wechat_login_at`, `created_at`, `updated_at`
) VALUES (
    1, 1, 'merchant', '$2a$10$mP89UzDWaHy0LVxdDqWhheUJ/UN4tVkArcEhTqW7kqScW7lk.558W',
    '商家管理员', '13900139000', NULL, NULL, NULL, 'owner', 1, 1, 1, NULL, NULL, NOW(), NOW()
);

INSERT INTO `merchant_delivery_settings` (
    `id`, `merchant_id`, `enabled`, `base_fee`, `free_delivery_amount`, `max_distance`, `distance_rules`, `created_at`, `updated_at`
) VALUES (
    1, 1, 1, 5.00, 50.00, 10,
    '[{"min_distance":0,"max_distance":2,"fee":0},{"min_distance":2,"max_distance":5,"fee":3},{"min_distance":5,"max_distance":10,"fee":6}]',
    NOW(), NOW()
);

INSERT INTO `merchant_licenses` (
    `id`, `merchant_id`, `license_no`, `license_name`, `license_image`, `legal_person`, `legal_person_id`,
    `legal_person_id_front`, `legal_person_id_back`, `valid_from`, `valid_to`, `status`, `created_at`, `updated_at`
) VALUES (
    1, 1, '91110000000000001X', '美味餐厅有限公司', 'https://example.com/images/license_1.jpg',
    '李四', '110101199001011234', 'https://example.com/images/id_front_1.jpg',
    'https://example.com/images/id_back_1.jpg', '2020-01-01', '2030-12-31', 1, NOW(), NOW()
);

SET FOREIGN_KEY_CHECKS = 1;
