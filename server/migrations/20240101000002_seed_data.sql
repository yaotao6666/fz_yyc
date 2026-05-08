-- ============================================
-- 寻梦私域管家 - 初始化数据
-- 用于开发测试环境
-- ============================================

-- 使用数据库
-- USE fz_yyc_api;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 1. 服务商数据
-- ============================================
INSERT INTO `service_providers` (`id`, `name`, `contact_name`, `contact_phone`, `mch_id`, `api_key`, `api_v3_key`, `cert_serial_no`, `private_key`, `public_key`, `callback_url`, `status`, `created_at`, `updated_at`)
VALUES
(1, '寻梦服务商', '张三', '13800138000', '1234567890', 'mock_api_key_12345', 'mock_api_v3_key_1234567890123456789012', 'mock_cert_serial_no', 'mock_private_key', 'mock_public_key', 'https://example.com/notify/payment', 1, NOW(), NOW());

-- ============================================
-- 2. 服务商管理员账号
-- 密码: admin123 (bcrypt加密)
-- ============================================
INSERT INTO `service_provider_admins` (`id`, `service_provider_id`, `username`, `password`, `name`, `phone`, `role`, `status`, `last_login_at`, `created_at`, `updated_at`)
VALUES
(1, 1, 'admin', '$2a$10$GCgIm2gqB7yPpx/w.pEVDeBj.xUzjOGFmsZdAZh2xLDr4RgYj2z6e', '超级管理员', '13800138000', 'admin', 1, NULL, NOW(), NOW());

-- ============================================
-- 3. 商家数据
-- 关联到服务商ID=1
-- ============================================
INSERT INTO `merchants` (`id`, `service_provider_id`, `name`, `logo`, `contact_name`, `contact_phone`, `contact_email`, `address`, `lat`, `lng`, `business_category`, `business_hours`, `announcement`, `min_order_amount`, `takeout_enabled`, `dine_in_enabled`, `sub_mch_id`, `sub_mch_status`, `applyment_status`, `audit_status`, `audit_remark`, `status`, `rating`, `sales_count`, `qrcode_url`, `created_at`, `updated_at`)
VALUES
(1, 1, '美味餐厅', 'https://example.com/images/merchant_logo_1.jpg', '李四', '13900139000', 'lisi@example.com', '北京市朝阳区建国路88号', 39.908823, 116.407470, '餐饮', '09:00-22:00', '欢迎光临！今日特惠：招牌菜8折', 20.00, 1, 1, '1500000001', 2, 2, 1, '资质审核通过', 1, 4.8, 256, 'https://example.com/qrcode/merchant_1.png', NOW(), NOW());

-- ============================================
-- 4. 商家员工账号（商家管理员）
-- 密码: merchant123 (bcrypt加密)
-- ============================================
INSERT INTO `merchant_staffs` (`id`, `merchant_id`, `username`, `password`, `name`, `phone`, `role`, `status`, `last_login_at`, `created_at`, `updated_at`)
VALUES
(1, 1, 'merchant', '$2a$10$mP89UzDWaHy0LVxdDqWhheUJ/UN4tVkArcEhTqW7kqScW7lk.558W', '商家管理员', '13900139000', 'owner', 1, NULL, NOW(), NOW());

-- ============================================
-- 5. 商家配送设置
-- ============================================
INSERT INTO `merchant_delivery_settings` (`id`, `merchant_id`, `enabled`, `base_fee`, `free_delivery_amount`, `max_distance`, `distance_rules`, `created_at`, `updated_at`)
VALUES
(1, 1, 1, 5.00, 50.00, 10, '[{"min_distance":0,"max_distance":2,"fee":0},{"min_distance":2,"max_distance":5,"fee":3.00},{"min_distance":5,"max_distance":10,"fee":6.00}]', NOW(), NOW());

-- ============================================
-- 6. 商家营业执照
-- ============================================
INSERT INTO `merchant_licenses` (`id`, `merchant_id`, `license_no`, `license_name`, `license_image`, `legal_person`, `legal_person_id`, `legal_person_id_front`, `legal_person_id_back`, `valid_from`, `valid_to`, `status`, `created_at`, `updated_at`)
VALUES
(1, 1, '91110000000000001X', '美味餐厅有限公司', 'https://example.com/images/license_1.jpg', '李四', '110101199001011234', 'https://example.com/images/id_front_1.jpg', 'https://example.com/images/id_back_1.jpg', '2020-01-01', '2030-12-31', 1, NOW(), NOW());

-- ============================================
-- 7. 商品分类
-- ============================================
INSERT INTO `categories` (`id`, `merchant_id`, `name`, `sort`, `status`, `created_at`, `updated_at`)
VALUES
(1, 1, '热销推荐', 1, 1, NOW(), NOW()),
(2, 1, '招牌菜', 2, 1, NOW(), NOW()),
(3, 1, '凉菜', 3, 1, NOW(), NOW()),
(4, 1, '主食', 4, 1, NOW(), NOW()),
(5, 1, '饮品', 5, 1, NOW(), NOW());

-- ============================================
-- 8. 商品数据
-- ============================================
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

-- ============================================
-- 9. 商品规格
-- ============================================
INSERT INTO `product_specs` (`id`, `product_id`, `name`, `options`, `created_at`, `updated_at`)
VALUES
(1, 1, '份量', '[{"name":"小份","price":48.00},{"name":"大份","price":68.00}]', NOW(), NOW()),
(2, 2, '份量', '[{"name":"小份","price":32.00},{"name":"大份","price":38.00}]', NOW(), NOW()),
(3, 3, '份量', '[{"name":"小份","price":36.00},{"name":"大份","price":42.00}]', NOW(), NOW());

-- ============================================
-- 10. C端用户数据
-- ============================================
INSERT INTO `users` (`id`, `openid`, `union_id`, `nickname`, `avatar`, `phone`, `status`, `created_at`, `updated_at`)
VALUES
(1, 'mock_openid_001', 'mock_union_id_001', '小明', 'https://example.com/avatar/user1.jpg', '13811112222', 1, NOW(), NOW()),
(2, 'mock_openid_002', 'mock_union_id_002', '小红', 'https://example.com/avatar/user2.jpg', '13811113333', 1, NOW(), NOW()),
(3, 'mock_openid_003', 'mock_union_id_003', '小张', 'https://example.com/avatar/user3.jpg', '13811114444', 1, NOW(), NOW()),
(4, 'mock_openid_004', 'mock_union_id_004', '小李', 'https://example.com/avatar/user4.jpg', '13811115555', 1, NOW(), NOW()),
(5, 'mock_openid_005', 'mock_union_id_005', '小王', 'https://example.com/avatar/user5.jpg', '13811116666', 1, NOW(), NOW());

-- ============================================
-- 11. 订单数据
-- ============================================
INSERT INTO `orders` (`id`, `order_no`, `user_id`, `merchant_id`, `total_amount`, `delivery_fee`, `discount_amount`, `pay_amount`, `delivery_type`, `delivery_distance`, `delivery_address`, `contact_name`, `contact_phone`, `status`, `remark`, `verify_code`, `transaction_id`, `paid_at`, `completed_at`, `cancelled_at`, `refunded_at`, `created_at`, `updated_at`)
VALUES
(1, 'ORD202401010001', 1, 1, 116.00, 3.00, 0.00, 119.00, 1, 3.50, '北京市朝阳区建国路100号', '小明', '13811112222', 2, '少放辣', '123456', 'WX2024010100001', '2024-01-01 12:01:00', NULL, NULL, NULL, '2024-01-01 12:00:00', NOW()),
(2, 'ORD202401010002', 2, 1, 180.00, 0.00, 10.00, 170.00, 2, NULL, NULL, '小红', '13811113333', 3, '打包带走', '234567', 'WX2024010100002', '2024-01-01 13:05:00', '2024-01-01 14:30:00', NULL, NULL, '2024-01-01 13:00:00', NOW()),
(3, 'ORD202401010003', 3, 1, 58.00, 0.00, 0.00, 58.00, 3, NULL, '北京市朝阳区建国路200号自提', '小张', '13811114444', 2, '', '345678', 'WX2024010100003', '2024-01-01 18:30:00', NULL, NULL, NULL, '2024-01-01 18:00:00', NOW()),
(4, 'ORD202401020001', 4, 1, 89.00, 6.00, 0.00, 95.00, 1, 5.50, '北京市朝阳区东三环中路50号', '小李', '13811115555', 2, '请带餐具', '456789', 'WX2024010200001', '2024-01-02 19:15:00', NULL, NULL, NULL, '2024-01-02 19:00:00', NOW()),
(5, 'ORD202401030001', 5, 1, 220.00, 3.00, 20.00, 203.00, 1, 3.00, '北京市朝阳区西大望路1号', '小王', '13811116666', 1, '晚上8点送到', '567890', NULL, NULL, NULL, NULL, NULL, '2024-01-03 19:30:00', NOW());

-- ============================================
-- 12. 订单商品数据
-- ============================================
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
-- 13. 平台活动数据
-- ============================================
INSERT INTO `activities` (`id`, `type`, `title`, `content`, `image`, `link_type`, `link_value`, `sort`, `status`, `start_time`, `end_time`, `created_at`, `updated_at`)
VALUES
(1, 'banner', '新商家入驻优惠', '新商家入驻享受0.2%超低费率', 'https://example.com/images/banner_1.jpg', 'invite', '', 1, 1, '2024-01-01 00:00:00', '2024-12-31 23:59:59', NOW(), NOW()),
(2, 'banner', '春节特惠活动', '春节期间所有商家商品8折起', 'https://example.com/images/banner_2.jpg', 'webview', 'https://example.com/activity/spring', 2, 1, '2024-01-20 00:00:00', '2024-02-10 23:59:59', NOW(), NOW()),
(3, 'announcement', '平台升级通知', '平台将于本周日凌晨2:00-6:00进行系统升级，届时服务暂时不可用。', NULL, 'none', '', 1, 1, NULL, NULL, NOW(), NOW());

-- ============================================
-- 14. 邀请奖励规则
-- ============================================
INSERT INTO `invite_rewards` (`id`, `type`, `condition`, `description`, `enabled`, `created_at`, `updated_at`)
VALUES
(1, 'free_year', 'first_transaction', '被邀请商家完成首次交易后，邀请人可获得一定期限的免年费', 1, NOW(), NOW()),
(2, 'lowest_rate', 'merchant_joined', '被邀请商家入驻后，邀请人可申请最低0.2%交易费率', 1, NOW(), NOW());

-- ============================================
-- 恢复外键检查
-- ============================================
SET FOREIGN_KEY_CHECKS = 1;
