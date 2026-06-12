-- 新增商家自提点表 + 订单自提点字段

CREATE TABLE IF NOT EXISTS `merchant_pickup_points` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `merchant_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(64) NOT NULL,
    `address` VARCHAR(256) NOT NULL,
    `lat` DECIMAL(10,6) NOT NULL,
    `lng` DECIMAL(10,6) NOT NULL,
    `is_default` TINYINT(1) NOT NULL DEFAULT 0,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `sort` INT UNSIGNED NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_merchant_pickup_points_merchant_id` (`merchant_id`),
    KEY `idx_merchant_pickup_points_default` (`merchant_id`, `is_default`),
    CONSTRAINT `fk_merchant_pickup_points_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家自提点';

ALTER TABLE `orders`
    ADD COLUMN `pickup_point_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '自提点ID(自提订单)' AFTER `contact_phone`,
    ADD COLUMN `pickup_point_name` VARCHAR(64) DEFAULT NULL COMMENT '自提点名称快照' AFTER `pickup_point_id`,
    ADD COLUMN `pickup_point_address` VARCHAR(256) DEFAULT NULL COMMENT '自提点地址快照' AFTER `pickup_point_name`,
    ADD COLUMN `pickup_point_lat` DECIMAL(10,6) DEFAULT NULL COMMENT '自提点纬度快照' AFTER `pickup_point_address`,
    ADD COLUMN `pickup_point_lng` DECIMAL(10,6) DEFAULT NULL COMMENT '自提点经度快照' AFTER `pickup_point_lat`,
    ADD KEY `idx_orders_pickup_point_id` (`pickup_point_id`);

