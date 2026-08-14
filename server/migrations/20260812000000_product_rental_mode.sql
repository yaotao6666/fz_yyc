-- 商品租赁模式升级
-- 为商品增加销售类型属性，区分一口价和租赁两种模式
-- 租赁模式支持：计费周期、单位租金、押金、最大租赁时长

-- 商品表新增租赁字段
ALTER TABLE `products`
    ADD COLUMN `sale_type` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '销售类型: 1=一口价 2=租赁' AFTER `unit`,
    ADD COLUMN `rental_unit` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '租赁计费周期: 0=非租赁 1=按天 2=按周 3=按月' AFTER `sale_type`,
    ADD COLUMN `rental_price` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '单位租金(元)' AFTER `rental_unit`,
    ADD COLUMN `deposit` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '押金(元)' AFTER `rental_price`,
    ADD COLUMN `max_rental_duration` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最大租赁时长(0=不限)' AFTER `deposit`,
    ADD KEY `idx_products_sale_type` (`merchant_id`, `sale_type`);

-- 订单商品表新增租赁快照字段
ALTER TABLE `order_items`
    ADD COLUMN `sale_type` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '销售类型快照: 1=一口价 2=租赁' AFTER `subtotal`,
    ADD COLUMN `rental_unit` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '租赁计费周期快照' AFTER `sale_type`,
    ADD COLUMN `rental_duration` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '租赁时长(下单时选择)' AFTER `rental_unit`,
    ADD COLUMN `unit_rental_price` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '单位租金快照' AFTER `rental_duration`,
    ADD COLUMN `rental_subtotal` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '租金小计=单价×时长' AFTER `unit_rental_price`,
    ADD COLUMN `deposit` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '单商品押金快照' AFTER `rental_subtotal`,
    ADD COLUMN `deposit_deduct` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '押金扣除金额(损坏赔偿)' AFTER `deposit`;

-- 订单表新增押金与租赁状态字段
ALTER TABLE `orders`
    ADD COLUMN `total_deposit` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '总押金(元)' AFTER `pay_amount`,
    ADD COLUMN `deposit_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '押金状态: 0=无押金 1=已收 2=已退 3=部分扣除' AFTER `total_deposit`,
    ADD COLUMN `deposit_refund_amount` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '押金退还金额' AFTER `deposit_status`,
    ADD COLUMN `deposit_deduct_amount` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '押金扣除总额(损坏赔偿)' AFTER `deposit_refund_amount`,
    ADD COLUMN `deposit_refunded_at` DATETIME DEFAULT NULL COMMENT '押金退还时间' AFTER `deposit_deduct_amount`,
    ADD COLUMN `rental_returned_at` DATETIME DEFAULT NULL COMMENT '租赁归还时间' AFTER `deposit_refunded_at`,
    ADD COLUMN `rental_return_remark` VARCHAR(256) DEFAULT NULL COMMENT '归还备注(验机情况)' AFTER `rental_returned_at`,
    ADD KEY `idx_orders_deposit_status` (`deposit_status`);
