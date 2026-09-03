-- ============================================================
-- 027_mock_rental_orders.sql — 生成几条已支付的租赁订单（mock 数据）
-- ------------------------------------------------------------
-- 目标用户：users.id=64（已有健康档案/地址/历史订单的演示用户）
-- 商品：20001 / 20004 / 20006（product_type=2 辅具租赁，含租金与押金）
-- 说明：
--  1. order_type=2 租赁、status=2 已支付、deposit_status=1 已收押金、biz_status=0
--  2. pay_amount = total_amount(租金小计) + total_deposit(押金)
--  3. rental_end_at = paid_at + 租期
--  4. 订单项从 products 表取名称/租金/押金快照，避免手工中文乱码
--  5. 幂等：按 order_no 去重；多记录用事务保证原子性
-- ============================================================
SET NAMES utf8mb4;

START TRANSACTION;

-- 订单 A：20001，月租50×1期，押金200，实付250
INSERT INTO `orders` (`order_no`,`user_id`,`order_type`,`total_amount`,`delivery_fee`,`discount_amount`,`pay_amount`,`total_deposit`,`deposit_status`,`status`,`biz_status`,`verify_code`,`paid_at`,`rental_end_at`,`delivery_district`,`delivery_address`,`contact_name`,`contact_phone`,`profit_sharing_status`)
SELECT '20260901100000821034',64,2,50.00,0,0,250.00,200.00,1,2,0,'821034','2026-09-01 10:00:00','2026-10-01 10:00:00',a.district,a.address,a.name,a.phone,0
FROM `user_addresses` a
WHERE a.id = (SELECT id FROM `user_addresses` WHERE user_id=64 ORDER BY is_default DESC,id LIMIT 1)
  AND NOT EXISTS (SELECT 1 FROM `orders` WHERE order_no='20260901100000821034');

-- 订单 B：20004，月租300×1期，押金800，实付1100
INSERT INTO `orders` (`order_no`,`user_id`,`order_type`,`total_amount`,`delivery_fee`,`discount_amount`,`pay_amount`,`total_deposit`,`deposit_status`,`status`,`biz_status`,`verify_code`,`paid_at`,`rental_end_at`,`delivery_district`,`delivery_address`,`contact_name`,`contact_phone`,`profit_sharing_status`)
SELECT '20260902110000410278',64,2,300.00,0,0,1100.00,800.00,1,2,0,'410278','2026-09-02 11:00:00','2026-10-02 11:00:00',a.district,a.address,a.name,a.phone,0
FROM `user_addresses` a
WHERE a.id = (SELECT id FROM `user_addresses` WHERE user_id=64 ORDER BY is_default DESC,id LIMIT 1)
  AND NOT EXISTS (SELECT 1 FROM `orders` WHERE order_no='20260902110000410278');

-- 订单 C：20006，月租60×2期，租金小计120，押金300，实付420
INSERT INTO `orders` (`order_no`,`user_id`,`order_type`,`total_amount`,`delivery_fee`,`discount_amount`,`pay_amount`,`total_deposit`,`deposit_status`,`status`,`biz_status`,`verify_code`,`paid_at`,`rental_end_at`,`delivery_district`,`delivery_address`,`contact_name`,`contact_phone`,`profit_sharing_status`)
SELECT '20260815120000663901',64,2,120.00,0,0,420.00,300.00,1,2,0,'663901','2026-08-15 12:00:00','2026-10-15 12:00:00',a.district,a.address,a.name,a.phone,0
FROM `user_addresses` a
WHERE a.id = (SELECT id FROM `user_addresses` WHERE user_id=64 ORDER BY is_default DESC,id LIMIT 1)
  AND NOT EXISTS (SELECT 1 FROM `orders` WHERE order_no='20260815120000663901');

-- 订单项（名称/租金/押金取自 products 快照）
INSERT INTO `order_items` (`order_id`,`product_id`,`product_name`,`image`,`price`,`quantity`,`subtotal`,`sale_type`,`rental_unit`,`rental_duration`,`unit_rental_price`,`rental_subtotal`,`deposit`)
SELECT o.id,p.id,p.name,NULL,p.rental_price,1,p.rental_price*1,2,p.rental_unit,1,p.rental_price,p.rental_price*1,p.deposit
FROM `orders` o JOIN `products` p ON p.id=20001
WHERE o.order_no='20260901100000821034'
  AND NOT EXISTS (SELECT 1 FROM `order_items` WHERE order_id=o.id AND product_id=20001);

INSERT INTO `order_items` (`order_id`,`product_id`,`product_name`,`image`,`price`,`quantity`,`subtotal`,`sale_type`,`rental_unit`,`rental_duration`,`unit_rental_price`,`rental_subtotal`,`deposit`)
SELECT o.id,p.id,p.name,NULL,p.rental_price,1,p.rental_price*1,2,p.rental_unit,1,p.rental_price,p.rental_price*1,p.deposit
FROM `orders` o JOIN `products` p ON p.id=20004
WHERE o.order_no='20260902110000410278'
  AND NOT EXISTS (SELECT 1 FROM `order_items` WHERE order_id=o.id AND product_id=20004);

INSERT INTO `order_items` (`order_id`,`product_id`,`product_name`,`image`,`price`,`quantity`,`subtotal`,`sale_type`,`rental_unit`,`rental_duration`,`unit_rental_price`,`rental_subtotal`,`deposit`)
SELECT o.id,p.id,p.name,NULL,p.rental_price,2,p.rental_price*2,2,p.rental_unit,2,p.rental_price,p.rental_price*2,p.deposit
FROM `orders` o JOIN `products` p ON p.id=20006
WHERE o.order_no='20260815120000663901'
  AND NOT EXISTS (SELECT 1 FROM `order_items` WHERE order_id=o.id AND product_id=20006);

COMMIT;