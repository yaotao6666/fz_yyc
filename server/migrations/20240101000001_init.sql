-- ============================================
-- 微信支付服务商商家管理系统 - 数据库表结构
-- 数据库: MySQL 8.0
-- 版本: v1.0
-- 创建时间: 2024-01-01
-- ============================================

-- 创建数据库（如果不存在）
-- CREATE DATABASE IF NOT EXISTS fz_yyc_api CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
-- USE fz_yyc_api;

-- 设置默认字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 1. 服务商表 (service_providers)
-- ============================================
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

-- ============================================
-- 2. 服务商管理员表 (service_provider_admins)
-- ============================================
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

-- ============================================
-- 3. 商家表 (merchants)
-- ============================================
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

-- ============================================
-- 4. 商家进件申请表 (merchant_applications)
-- ============================================
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

-- ============================================
-- 5. 商家配送设置表 (merchant_delivery_settings)
-- ============================================
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

-- ============================================
-- 6. 商家营业执照表 (merchant_licenses)
-- ============================================
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

-- ============================================
-- 7. 商家员工表 (merchant_staffs)
-- ============================================
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

-- ============================================
-- 9. 商品分类表 (categories)
-- ============================================
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

-- ============================================
-- 10. 商品表 (products)
-- ============================================
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

-- ============================================
-- 11. 商品规格表 (product_specs)
-- ============================================
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

-- ============================================
-- 12. C端用户表 (users)
-- ============================================
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

-- ============================================
-- 14. 订单表 (orders)
-- ============================================
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

-- ============================================
-- 15. 订单商品表 (order_items)
-- ============================================
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

-- ============================================
-- 16. 退款记录表 (refunds)
-- ============================================
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
-- 恢复外键检查
-- ============================================
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 初始化数据示例（可选）
-- ============================================

-- 插入服务商数据
-- INSERT INTO `service_providers` (`name`, `contact_name`, `contact_phone`, `mch_id`, `status`)
-- VALUES ('示例服务商', '张三', '13800138000', '1234567890', 1);

-- 插入服务商管理员（密码：admin123，需要先加密）
-- INSERT INTO `service_provider_admins` (`service_provider_id`, `username`, `password`, `name`, `role`, `status`)
-- VALUES (1, 'admin', '$2b$12$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx', '超级管理员', 'admin', 1);
