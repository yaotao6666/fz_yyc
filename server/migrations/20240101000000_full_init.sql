-- ============================================
-- 寻梦私域管家 - 数据库初始化脚本
-- 数据库: MySQL 8.0
-- 版本: v1.0
-- 创建时间: 2024-01-01
-- 最后更新: 2026-05-10
-- ============================================

-- ============================================
-- 第一部分：数据库设置
-- ============================================
CREATE DATABASE IF NOT EXISTS fz_yyc_api CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE fz_yyc_api;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 第二部分：表结构定义
-- ============================================

-- 1. 服务商表 (service_providers)
DROP TABLE IF EXISTS `service_providers`;
CREATE TABLE `service_providers` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `name` VARCHAR(128) NOT NULL COMMENT '服务商名称',
    `contact_name` VARCHAR(64) DEFAULT NULL COMMENT '联系人',
    `contact_phone` VARCHAR(20) DEFAULT NULL COMMENT '联系电话',
    `mch_id` VARCHAR(32) DEFAULT NULL COMMENT '服务商商户号',
    `api_key` VARCHAR(128) DEFAULT NULL COMMENT 'API密钥（加密存储）',
    `api_v3_key` VARCHAR(128) DEFAULT NULL COMMENT 'APIv3密钥（加密存储）',
    `cert_serial_no` VARCHAR(64) DEFAULT NULL COMMENT '证书序列号',
    `private_key` TEXT DEFAULT NULL COMMENT '商户私钥（加密存储）',
    `public_key` TEXT DEFAULT NULL COMMENT '平台公钥',
    `callback_url` VARCHAR(256) DEFAULT NULL COMMENT '支付回调地址',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0禁用 1正常',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_service_providers_mch_id` (`mch_id`),
    INDEX `idx_service_providers_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务商表';

-- 2. 服务商管理员表 (service_provider_admins)
DROP TABLE IF EXISTS `service_provider_admins`;
CREATE TABLE `service_provider_admins` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `service_provider_id` BIGINT UNSIGNED NOT NULL COMMENT '服务商ID',
    `username` VARCHAR(64) NOT NULL COMMENT '用户名',
    `password` VARCHAR(128) NOT NULL COMMENT '密码（bcrypt加密）',
    `name` VARCHAR(64) DEFAULT NULL COMMENT '姓名',
    `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    `role` VARCHAR(32) NOT NULL DEFAULT 'operator' COMMENT '角色：admin/operator',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0禁用 1正常',
    `last_login_at` DATETIME DEFAULT NULL COMMENT '最后登录时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_service_provider_admins_username` (`username`),
    INDEX `idx_service_provider_admins_sp_id` (`service_provider_id`),
    CONSTRAINT `fk_service_provider_admins_sp` FOREIGN KEY (`service_provider_id`) REFERENCES `service_providers` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务商管理员表';

-- 3. 商家表 (merchants)
DROP TABLE IF EXISTS `merchants`;
CREATE TABLE `merchants` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `service_provider_id` BIGINT UNSIGNED NOT NULL COMMENT '服务商ID',
    `name` VARCHAR(128) NOT NULL COMMENT '商家名称',
    `logo` VARCHAR(512) DEFAULT NULL COMMENT '店铺Logo',
    `contact_name` VARCHAR(64) DEFAULT NULL COMMENT '联系人',
    `contact_phone` VARCHAR(20) DEFAULT NULL COMMENT '联系电话',
    `contact_email` VARCHAR(128) DEFAULT NULL COMMENT '联系邮箱',
    `address` VARCHAR(256) DEFAULT NULL COMMENT '店铺地址',
    `lat` DECIMAL(10,6) DEFAULT NULL COMMENT '纬度',
    `lng` DECIMAL(10,6) DEFAULT NULL COMMENT '经度',
    `business_category` VARCHAR(64) DEFAULT NULL COMMENT '经营类目',
    `business_hours` VARCHAR(64) DEFAULT NULL COMMENT '营业时间',
    `announcement` TEXT DEFAULT NULL COMMENT '门店公告',
    `min_order_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '最低起送金额',
    `takeout_enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '是否支持外卖：0否 1是',
    `dine_in_enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '是否支持堂食：0否 1是',
    `sub_mch_id` VARCHAR(32) DEFAULT NULL COMMENT '微信支付子商户号',
    `sub_mch_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '子商户绑定状态：0未绑定 1绑定中 2已绑定 3绑定失败',
    `applyment_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '进件状态：0未进件 1进件中 2已通过 3已拒绝',
    `audit_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '平台审核状态：0待审核 1通过 2拒绝',
    `audit_remark` VARCHAR(256) DEFAULT NULL COMMENT '审核备注',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0关闭 1营业中',
    `rating` DECIMAL(2,1) NOT NULL DEFAULT 5.0 COMMENT '评分',
    `sales_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销量',
    `qrcode_url` VARCHAR(512) DEFAULT NULL COMMENT '商家小程序码URL',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_merchants_sp_id` (`service_provider_id`),
    INDEX `idx_merchants_sub_mch_id` (`sub_mch_id`),
    INDEX `idx_merchants_audit_status` (`audit_status`),
    INDEX `idx_merchants_status` (`status`),
    INDEX `idx_merchants_location` (`lat`, `lng`),
    CONSTRAINT `fk_merchants_sp` FOREIGN KEY (`service_provider_id`) REFERENCES `service_providers` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家表';

-- 4. 商家进件申请表 (merchant_applications)
DROP TABLE IF EXISTS `merchant_applications`;
CREATE TABLE `merchant_applications` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `merchant_id` BIGINT UNSIGNED NOT NULL COMMENT '商家ID',
    `merchant_name` VARCHAR(128) NOT NULL COMMENT '商户名称',
    `business_license_info` JSON DEFAULT NULL COMMENT '营业执照信息',
    `legal_person_info` JSON DEFAULT NULL COMMENT '法人信息',
    `bank_account_info` JSON DEFAULT NULL COMMENT '结算银行卡信息',
    `store_info` JSON DEFAULT NULL COMMENT '门店信息',
    `contact_info` JSON DEFAULT NULL COMMENT '联系人信息',
    `applyment_id` VARCHAR(64) DEFAULT NULL COMMENT '微信支付申请单号',
    `sub_mch_id` VARCHAR(32) DEFAULT NULL COMMENT '子商户号',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态：0草稿 1已提交 2审核中 3通过 4拒绝',
    `audit_detail` JSON DEFAULT NULL COMMENT '审核详情',
    `submit_time` DATETIME DEFAULT NULL COMMENT '提交时间',
    `audit_time` DATETIME DEFAULT NULL COMMENT '审核时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_merchant_applications_merchant_id` (`merchant_id`),
    INDEX `idx_merchant_applications_applyment_id` (`applyment_id`),
    INDEX `idx_merchant_applications_sub_mch_id` (`sub_mch_id`),
    INDEX `idx_merchant_applications_status` (`status`),
    CONSTRAINT `fk_merchant_applications_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家进件申请表';

-- 5. 商家配送设置表 (merchant_delivery_settings)
DROP TABLE IF EXISTS `merchant_delivery_settings`;
CREATE TABLE `merchant_delivery_settings` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `merchant_id` BIGINT UNSIGNED NOT NULL COMMENT '商家ID',
    `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '是否开启配送：0否 1是',
    `base_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '基础配送费',
    `free_delivery_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '满额免配送费金额',
    `max_distance` INT UNSIGNED NOT NULL DEFAULT 10 COMMENT '最大配送距离（公里）',
    `distance_rules` JSON NOT NULL COMMENT '按距离收费规则',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_merchant_delivery_settings_merchant_id` (`merchant_id`),
    CONSTRAINT `fk_merchant_delivery_settings_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家配送设置表';

-- 6. 商家营业执照表 (merchant_licenses)
DROP TABLE IF EXISTS `merchant_licenses`;
CREATE TABLE `merchant_licenses` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `merchant_id` BIGINT UNSIGNED NOT NULL COMMENT '商家ID',
    `license_no` VARCHAR(64) DEFAULT NULL COMMENT '营业执照号',
    `license_name` VARCHAR(128) DEFAULT NULL COMMENT '营业执照名称',
    `license_image` VARCHAR(512) DEFAULT NULL COMMENT '营业执照图片',
    `legal_person` VARCHAR(64) DEFAULT NULL COMMENT '法人姓名',
    `legal_person_id` VARCHAR(32) DEFAULT NULL COMMENT '法人身份证号',
    `legal_person_id_front` VARCHAR(512) DEFAULT NULL COMMENT '法人身份证正面图片',
    `legal_person_id_back` VARCHAR(512) DEFAULT NULL COMMENT '法人身份证反面图片',
    `valid_from` DATE DEFAULT NULL COMMENT '有效期开始',
    `valid_to` DATE DEFAULT NULL COMMENT '有效期结束',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0失效 1有效',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_merchant_licenses_merchant_id` (`merchant_id`),
    CONSTRAINT `fk_merchant_licenses_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家营业执照表';

-- 7. 商家员工表 (merchant_staffs)
DROP TABLE IF EXISTS `merchant_staffs`;
CREATE TABLE `merchant_staffs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `merchant_id` BIGINT UNSIGNED NOT NULL COMMENT '商家ID',
    `username` VARCHAR(64) NOT NULL COMMENT '用户名',
    `password` VARCHAR(128) NOT NULL COMMENT '密码（bcrypt加密）',
    `name` VARCHAR(64) DEFAULT NULL COMMENT '姓名',
    `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    `role` VARCHAR(32) NOT NULL DEFAULT 'staff' COMMENT '角色：owner/manager/staff',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0禁用 1正常',
    `last_login_at` DATETIME DEFAULT NULL COMMENT '最后登录时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_merchant_staffs_merchant_username` (`merchant_id`, `username`),
    INDEX `idx_merchant_staffs_phone` (`phone`),
    CONSTRAINT `fk_merchant_staffs_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家员工表';

-- 8. 系统公告表 (announcements)
DROP TABLE IF EXISTS `announcements`;
CREATE TABLE `announcements` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `service_provider_id` BIGINT UNSIGNED NOT NULL COMMENT '服务商ID',
    `title` VARCHAR(128) NOT NULL COMMENT '公告标题',
    `content` TEXT DEFAULT NULL COMMENT '公告内容',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0禁用 1正常',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_announcements_sp_id` (`service_provider_id`),
    INDEX `idx_announcements_status` (`status`),
    INDEX `idx_announcements_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统公告表';

-- 9. 商家审核记录表 (merchant_audit_records)
DROP TABLE IF EXISTS `merchant_audit_records`;
CREATE TABLE `merchant_audit_records` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `merchant_id` BIGINT UNSIGNED NOT NULL COMMENT '商家ID',
    `auditor_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '审核员ID（服务商管理员ID）',
    `action` VARCHAR(32) NOT NULL COMMENT '操作类型：submit/approve/reject',
    `status` VARCHAR(32) NOT NULL COMMENT '状态：pending/approved/rejected',
    `remark` TEXT DEFAULT NULL COMMENT '审核备注',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_merchant_audit_records_merchant_id` (`merchant_id`),
    INDEX `idx_merchant_audit_records_auditor_id` (`auditor_id`),
    INDEX `idx_merchant_audit_records_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家审核记录表';

-- 10. 商品分类表 (categories)
DROP TABLE IF EXISTS `categories`;
CREATE TABLE `categories` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `merchant_id` BIGINT UNSIGNED NOT NULL COMMENT '商家ID',
    `name` VARCHAR(64) NOT NULL COMMENT '分类名称',
    `sort` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序权重',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0禁用 1启用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_categories_merchant_id` (`merchant_id`),
    INDEX `idx_categories_sort` (`merchant_id`, `sort`),
    CONSTRAINT `fk_categories_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品分类表';

-- 11. 商品表 (products)
DROP TABLE IF EXISTS `products`;
CREATE TABLE `products` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `merchant_id` BIGINT UNSIGNED NOT NULL COMMENT '商家ID',
    `category_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '分类ID',
    `name` VARCHAR(128) NOT NULL COMMENT '商品名称',
    `description` TEXT DEFAULT NULL COMMENT '商品描述',
    `images` JSON NOT NULL COMMENT '图片数组',
    `price` DECIMAL(10,2) NOT NULL COMMENT '售价',
    `original_price` DECIMAL(10,2) DEFAULT NULL COMMENT '原价',
    `stock` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '库存',
    `unit` VARCHAR(16) NOT NULL DEFAULT '份' COMMENT '单位',
    `sales` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销量',
    `sort` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序权重',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0下架 1上架',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间（软删除）',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_products_merchant_id` (`merchant_id`),
    INDEX `idx_products_category_id` (`category_id`),
    INDEX `idx_products_status` (`status`),
    INDEX `idx_products_sales` (`merchant_id`, `sales` DESC),
    INDEX `idx_products_sort` (`merchant_id`, `sort`),
    CONSTRAINT `fk_products_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';

-- 12. 商品规格表 (product_specs)
DROP TABLE IF EXISTS `product_specs`;
CREATE TABLE `product_specs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `name` VARCHAR(64) NOT NULL COMMENT '规格名称（如：规格、尺寸）',
    `options` JSON NOT NULL COMMENT '规格选项',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_product_specs_product_id` (`product_id`),
    CONSTRAINT `fk_product_specs_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品规格表';

-- 13. C端用户表 (users)
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `openid` VARCHAR(64) DEFAULT NULL COMMENT '微信OpenID',
    `union_id` VARCHAR(64) DEFAULT NULL COMMENT '微信UnionID',
    `nickname` VARCHAR(64) DEFAULT NULL COMMENT '昵称',
    `avatar` VARCHAR(512) DEFAULT NULL COMMENT '头像',
    `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0禁用 1正常',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_users_openid` (`openid`),
    INDEX `idx_users_union_id` (`union_id`),
    INDEX `idx_users_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='C端用户表';

-- 14. 订单表 (orders)
DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `order_no` VARCHAR(32) NOT NULL COMMENT '订单号',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `merchant_id` BIGINT UNSIGNED NOT NULL COMMENT '商家ID',
    `total_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '商品总金额',
    `delivery_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '配送费',
    `discount_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '优惠金额',
    `pay_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '实付金额',
    `delivery_type` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '配送类型：1配送 2堂食 3自提',
    `delivery_distance` DECIMAL(5,2) DEFAULT NULL COMMENT '配送距离（公里）',
    `delivery_address` VARCHAR(256) DEFAULT NULL COMMENT '配送地址',
    `contact_name` VARCHAR(64) DEFAULT NULL COMMENT '联系人姓名',
    `contact_phone` VARCHAR(20) DEFAULT NULL COMMENT '联系人电话',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1待支付 2已支付 3已完成 4已取消 5退款中 6已退款',
    `remark` VARCHAR(256) DEFAULT NULL COMMENT '备注',
    `verify_code` VARCHAR(16) DEFAULT NULL COMMENT '核销码',
    `transaction_id` VARCHAR(64) DEFAULT NULL COMMENT '微信支付交易号',
    `paid_at` DATETIME DEFAULT NULL COMMENT '支付时间',
    `completed_at` DATETIME DEFAULT NULL COMMENT '完成时间',
    `cancelled_at` DATETIME DEFAULT NULL COMMENT '取消时间',
    `refunded_at` DATETIME DEFAULT NULL COMMENT '退款时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_orders_order_no` (`order_no`),
    INDEX `idx_orders_user_id` (`user_id`),
    INDEX `idx_orders_merchant_id` (`merchant_id`),
    INDEX `idx_orders_status` (`status`),
    INDEX `idx_orders_created_at` (`created_at` DESC),
    INDEX `idx_orders_paid_at` (`paid_at` DESC),
    INDEX `idx_orders_verify_code` (`verify_code`),
    CONSTRAINT `fk_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_orders_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 15. 订单商品表 (order_items)
DROP TABLE IF EXISTS `order_items`;
CREATE TABLE `order_items` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `merchant_id` BIGINT UNSIGNED NOT NULL COMMENT '商家ID（冗余，便于查询）',
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `product_name` VARCHAR(128) NOT NULL COMMENT '商品名称（冗余）',
    `image` VARCHAR(512) DEFAULT NULL COMMENT '商品图片（冗余）',
    `price` DECIMAL(10,2) NOT NULL COMMENT '单价（冗余）',
    `quantity` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '数量',
    `spec_info` JSON DEFAULT NULL COMMENT '规格信息（如：大份/小份）',
    `subtotal` DECIMAL(10,2) NOT NULL COMMENT '小计',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_order_items_order_id` (`order_id`),
    INDEX `idx_order_items_merchant_id` (`merchant_id`),
    INDEX `idx_order_items_product_id` (`product_id`),
    CONSTRAINT `fk_order_items_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_order_items_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_order_items_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品表';

-- 16. 退款记录表 (refunds)
DROP TABLE IF EXISTS `refunds`;
CREATE TABLE `refunds` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `refund_no` VARCHAR(32) NOT NULL COMMENT '退款单号',
    `refund_amount` DECIMAL(10,2) NOT NULL COMMENT '退款金额',
    `refund_reason` VARCHAR(256) DEFAULT NULL COMMENT '退款原因',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态：0处理中 1成功 2失败',
    `refund_id` VARCHAR(64) DEFAULT NULL COMMENT '微信退款单号',
    `refunded_at` DATETIME DEFAULT NULL COMMENT '退款完成时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_refunds_refund_no` (`refund_no`),
    INDEX `idx_refunds_order_id` (`order_id`),
    INDEX `idx_refunds_status` (`status`),
    CONSTRAINT `fk_refunds_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款记录表';

-- ============================================
-- 第三部分：测试数据
-- ============================================

-- 1. 服务商数据
INSERT INTO `service_providers` (`id`, `name`, `contact_name`, `contact_phone`, `mch_id`, `api_key`, `api_v3_key`, `cert_serial_no`, `private_key`, `public_key`, `callback_url`, `status`, `created_at`, `updated_at`)
VALUES
(1, '寻梦服务商', '张三', '13800138000', '1234567890', 'mock_api_key_12345', 'mock_api_v3_key_1234567890123456789012', 'mock_cert_serial_no', 'mock_private_key', 'mock_public_key', 'https://example.com/notify/payment', 1, NOW(), NOW());

-- 2. 服务商管理员账号
INSERT INTO `service_provider_admins` (`id`, `service_provider_id`, `username`, `password`, `name`, `phone`, `role`, `status`, `last_login_at`, `created_at`, `updated_at`)
VALUES
(1, 1, 'admin', '$2a$10$GCgIm2gqB7yPpx/w.pEVDeBj.xUzjOGFmsZdAZh2xLDr4RgYj2z6e', '超级管理员', '13800138000', 'admin', 1, NULL, NOW(), NOW());

-- 3. 商家数据
INSERT INTO `merchants` (`id`, `service_provider_id`, `name`, `logo`, `contact_name`, `contact_phone`, `contact_email`, `address`, `lat`, `lng`, `business_category`, `business_hours`, `announcement`, `min_order_amount`, `takeout_enabled`, `dine_in_enabled`, `sub_mch_id`, `sub_mch_status`, `applyment_status`, `audit_status`, `audit_remark`, `status`, `rating`, `sales_count`, `qrcode_url`, `created_at`, `updated_at`)
VALUES
(1, 1, '美味餐厅', 'https://example.com/images/merchant_logo_1.jpg', '李四', '13900139000', 'lisi@example.com', '北京市朝阳区建国路88号', 39.908823, 116.407470, '餐饮', '09:00-22:00', '欢迎光临！今日特惠：招牌菜8折', 20.00, 1, 1, '1500000001', 2, 2, 1, '资质审核通过', 1, 4.8, 256, 'https://example.com/qrcode/merchant_1.png', NOW(), NOW());

-- 4. 商家员工账号
INSERT INTO `merchant_staffs` (`id`, `merchant_id`, `username`, `password`, `name`, `phone`, `role`, `status`, `last_login_at`, `created_at`, `updated_at`)
VALUES
(1, 1, 'merchant', '$2a$10$mP89UzDWaHy0LVxdDqWhheUJ/UN4tVkArcEhTqW7kqScW7lk.558W', '商家管理员', '13900139000', 'owner', 1, NULL, NOW(), NOW());

-- 5. 商家配送设置
INSERT INTO `merchant_delivery_settings` (`id`, `merchant_id`, `enabled`, `base_fee`, `free_delivery_amount`, `max_distance`, `distance_rules`, `created_at`, `updated_at`)
VALUES
(1, 1, 1, 5.00, 50.00, 10, '[{"min_distance":0,"max_distance":2,"fee":0},{"min_distance":2,"max_distance":5,"fee":3.00},{"min_distance":5,"max_distance":10,"fee":6.00}]', NOW(), NOW());

-- 6. 商家营业执照
INSERT INTO `merchant_licenses` (`id`, `merchant_id`, `license_no`, `license_name`, `license_image`, `legal_person`, `legal_person_id`, `legal_person_id_front`, `legal_person_id_back`, `valid_from`, `valid_to`, `status`, `created_at`, `updated_at`)
VALUES
(1, 1, '91110000000000001X', '美味餐厅有限公司', 'https://example.com/images/license_1.jpg', '李四', '110101199001011234', 'https://example.com/images/id_front_1.jpg', 'https://example.com/images/id_back_1.jpg', '2020-01-01', '2030-12-31', 1, NOW(), NOW());

-- 7. 商品分类
INSERT INTO `categories` (`id`, `merchant_id`, `name`, `sort`, `status`, `created_at`, `updated_at`)
VALUES
(1, 1, '热销推荐', 1, 1, NOW(), NOW()),
(2, 1, '招牌菜', 2, 1, NOW(), NOW()),
(3, 1, '凉菜', 3, 1, NOW(), NOW()),
(4, 1, '主食', 4, 1, NOW(), NOW()),
(5, 1, '饮品', 5, 1, NOW(), NOW());

-- 8. 商品数据
INSERT INTO `products` (`id`, `merchant_id`, `category_id`, `name`, `description`, `images`, `price`, `original_price`, `stock`, `unit`, `sales`, `sort`, `status`, `deleted_at`, `created_at`, `updated_at`)
VALUES
(1, 1, 1, '招牌红烧肉', '精选五花肉，慢火炖煮3小时，入口即化', '["https://example.com/images/dish_1.jpg","https://example.com/images/dish_1_2.jpg"]', 58.00, 68.00, 100, '份', 256, 1, 1, NULL, NOW(), NOW()),
(2, 1, 1, '宫保鸡丁', '经典川菜，鸡丁滑嫩，花生酥脆', '["https://example.com/images/dish_2.jpg"]', 38.00, 45.00, 80, '份', 189, 2, 1, NULL, NOW(), NOW()),
(3, 1, 2, '糖醋里脊', '外酥里嫩，酸甜可口，老少皆宜', '["https://example.com/images/dish_3.jpg"]', 42.00, 50.00, 60, '份', 156, 1, 1, NULL, NOW(), NOW()),
(4, 1, 2, '水煮鱼', '新鲜草鱼，麻辣鲜香', '["https://example.com/images/dish_4.jpg"]', 88.00, 108.00, 40, '份', 98, 2, 1, NULL, NOW(), NOW()),
(5, 1, 3, '凉拌黄瓜', '清脆爽口，开胃小菜', '["https://example.com/images/dish_5.jpg"]', 18.00, NULL, 200, '份', 145, 3, 1, NULL, NOW(), NOW()),
(6, 1, 3, '凉拌木耳', '东北黑木耳，爽滑Q弹', '["https://example.com/images/dish_6.jpg"]', 22.00, NULL, 150, '份', 112, 1, 1, NULL, NOW(), NOW()),
(7, 1, 4, '米饭', '东北大米，香糯可口', '["https://example.com/images/rice.jpg"]', 3.00, NULL, 1000, '碗', 500, 4, 1, NULL, NOW(), NOW()),
(8, 1, 4, '馒头', '手工馒头，蓬松柔软', '["https://example.com/images/bun.jpg"]', 2.00, NULL, 500, '个', 300, 2, 1, NULL, NOW(), NOW()),
(9, 1, 5, '可乐', '冰镇可乐，清凉解渴', '["https://example.com/images/coke.jpg"]', 5.00, 6.00, 300, '瓶', 280, 5, 1, NULL, NOW(), NOW()),
(10, 1, 5, '鲜榨橙汁', '新鲜橙子鲜榨，不加水不加糖', '["https://example.com/images/orange_juice.jpg"]', 12.00, 15.00, 100, '杯', 156, 1, 1, NULL, NOW(), NOW());

-- 9. 商品规格
INSERT INTO `product_specs` (`id`, `product_id`, `name`, `options`, `created_at`, `updated_at`)
VALUES
(1, 1, '份量', '[{"name":"小份","price":48.00},{"name":"大份","price":68.00}]', NOW(), NOW()),
(2, 2, '份量', '[{"name":"小份","price":32.00},{"name":"大份","price":38.00}]', NOW(), NOW()),
(3, 3, '份量', '[{"name":"小份","price":36.00},{"name":"大份","price":42.00}]', NOW(), NOW());

-- 10. C端用户数据
INSERT INTO `users` (`id`, `openid`, `union_id`, `nickname`, `avatar`, `phone`, `status`, `created_at`, `updated_at`)
VALUES
(1, 'mock_openid_001', 'mock_union_id_001', '小明', 'https://example.com/avatar/user1.jpg', '13811112222', 1, NOW(), NOW()),
(2, 'mock_openid_002', 'mock_union_id_002', '小红', 'https://example.com/avatar/user2.jpg', '13811113333', 1, NOW(), NOW()),
(3, 'mock_openid_003', 'mock_union_id_003', '小张', 'https://example.com/avatar/user3.jpg', '13811114444', 1, NOW(), NOW()),
(4, 'mock_openid_004', 'mock_union_id_004', '小李', 'https://example.com/avatar/user4.jpg', '13811115555', 1, NOW(), NOW()),
(5, 'mock_openid_005', 'mock_union_id_005', '小王', 'https://example.com/avatar/user5.jpg', '13811116666', 1, NOW(), NOW());

-- 11. 订单数据
INSERT INTO `orders` (`id`, `order_no`, `user_id`, `merchant_id`, `total_amount`, `delivery_fee`, `discount_amount`, `pay_amount`, `delivery_type`, `delivery_distance`, `delivery_address`, `contact_name`, `contact_phone`, `status`, `remark`, `verify_code`, `transaction_id`, `paid_at`, `completed_at`, `cancelled_at`, `refunded_at`, `created_at`, `updated_at`)
VALUES
(1, 'ORD202401010001', 1, 1, 116.00, 3.00, 0.00, 119.00, 1, 3.50, '北京市朝阳区建国路100号', '小明', '13811112222', 2, '少放辣', '123456', 'WX2024010100001', '2024-01-01 12:01:00', NULL, NULL, NULL, '2024-01-01 12:00:00', NOW()),
(2, 'ORD202401010002', 2, 1, 180.00, 0.00, 10.00, 170.00, 2, NULL, NULL, '小红', '13811113333', 3, '打包带走', '234567', 'WX2024010100002', '2024-01-01 13:05:00', '2024-01-01 14:30:00', NULL, NULL, '2024-01-01 13:00:00', NOW()),
(3, 'ORD202401010003', 3, 1, 58.00, 0.00, 0.00, 58.00, 3, NULL, '北京市朝阳区建国路200号自提', '小张', '13811114444', 2, '', '345678', 'WX2024010100003', '2024-01-01 18:30:00', NULL, NULL, NULL, '2024-01-01 18:00:00', NOW()),
(4, 'ORD202401020001', 4, 1, 89.00, 6.00, 0.00, 95.00, 1, 5.50, '北京市朝阳区东三环中路50号', '小李', '13811115555', 2, '请带餐具', '456789', 'WX2024010200001', '2024-01-02 19:15:00', NULL, NULL, NULL, '2024-01-02 19:00:00', NOW()),
(5, 'ORD202401030001', 5, 1, 220.00, 3.00, 20.00, 203.00, 1, 3.00, '北京市朝阳区西大望路1号', '小王', '13811116666', 1, '晚上8点送到', '567890', NULL, NULL, NULL, NULL, NULL, '2024-01-03 19:30:00', NOW());

-- 12. 订单商品数据
INSERT INTO `order_items` (`id`, `order_id`, `merchant_id`, `product_id`, `product_name`, `image`, `price`, `quantity`, `spec_info`, `subtotal`, `created_at`)
VALUES
(1, 1, 1, 1, '招牌红烧肉', 'https://example.com/images/dish_1.jpg', 58.00, 2, '{"份量":"大份"}', 116.00, '2024-01-01 12:00:00'),
(2, 2, 1, 3, '糖醋里脊', 'https://example.com/images/dish_3.jpg', 42.00, 2, '{"份量":"大份"}', 84.00, '2024-01-01 13:00:00'),
(3, 2, 1, 7, '米饭', 'https://example.com/images/rice.jpg', 3.00, 2, NULL, 6.00, '2024-01-01 13:00:00'),
(4, 2, 1, 9, '可乐', 'https://example.com/images/coke.jpg', 5.00, 2, NULL, 10.00, '2024-01-01 13:00:00'),
(5, 2, 1, 10, '鲜榨橙汁', 'https://example.com/images/orange_juice.jpg', 12.00, 5, NULL, 60.00, '2024-01-01 13:00:00'),
(6, 3, 1, 1, '招牌红烧肉', 'https://example.com/images/dish_1.jpg', 58.00, 1, NULL, 58.00, '2024-01-01 18:00:00'),
(7, 4, 1, 2, '宫保鸡丁', 'https://example.com/images/dish_2.jpg', 38.00, 1, NULL, 38.00, '2024-01-02 19:00:00'),
(8, 4, 1, 4, '水煮鱼', 'https://example.com/images/dish_4.jpg', 88.00, 1, NULL, 88.00, '2024-01-02 19:00:00'),
(9, 5, 1, 4, '水煮鱼', 'https://example.com/images/dish_4.jpg', 88.00, 2, '{"份量":"大份"}', 176.00, '2024-01-03 19:30:00'),
(10, 5, 1, 7, '米饭', 'https://example.com/images/rice.jpg', 3.00, 2, NULL, 6.00, '2024-01-03 19:30:00'),
(11, 5, 1, 10, '鲜榨橙汁', 'https://example.com/images/orange_juice.jpg', 12.00, 3, NULL, 36.00, '2024-01-03 19:30:00');

-- ============================================
-- 第四部分：恢复外键检查
-- ============================================
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 初始化完成
-- ============================================
