-- ============================================================
-- 乐享辅具（云南财旭商贸）医疗辅具商城：分类清洗 + 店铺重建种子
-- 说明：
--   1) 最小化补齐表列：categories.category_type、products.product_type/service_content，
--      并让 categories/products 的 merchant_id 有默认值（单店架构 DefaultMerchantID=1，GORM 不写该列）。
--   2) 清空旧的餐厅 demo 商品/分类及关联数据，重建医疗辅具分类树（一级/二级/三级）。
--   3) 生成全品类二类医疗辅具商品 + 共享租赁低价套餐，并挂到对应分类。
-- 约束：纯 SQL 单条执行（SET @var+PREPARE+EXECUTE 模式 / 直接 SQL）。
-- ============================================================

-- ---- 1. 补列（防御式） ----
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND COLUMN_NAME='category_type');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE categories ADD COLUMN category_type TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '分类类型: 1=商品分类 2=服务分类' AFTER name");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND COLUMN_NAME='product_type');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE products ADD COLUMN product_type TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '商品类型: 1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务 5=科普活动 6=长护险服务' AFTER unit");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND COLUMN_NAME='service_content');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE products ADD COLUMN service_content JSON NULL COMMENT '服务型商品的内容描述JSON' AFTER product_type");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 让 categories/products 的 merchant_id 可默认写入（单店架构）
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1',
  "ALTER TABLE categories MODIFY COLUMN merchant_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '商家ID(单店默认1)'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1',
  "ALTER TABLE products MODIFY COLUMN merchant_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '商家ID(单店默认1)'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---- 2. 清空旧餐厅 demo 数据（先清关联表，再清商品与分类） ----
SET FOREIGN_KEY_CHECKS = 0;
DELETE FROM product_specs;
DELETE FROM order_items;
DELETE FROM refunds;
DELETE FROM orders;
DELETE FROM products;
DELETE FROM categories;
ALTER TABLE products AUTO_INCREMENT = 1;
ALTER TABLE categories AUTO_INCREMENT = 1;
SET FOREIGN_KEY_CHECKS = 1;

-- ---- 3. 分类树（一级） ----
INSERT INTO categories (id, name, category_type, parent_id, level, sort, status, created_at, updated_at) VALUES
(1, '轮椅', 1, NULL, 1, 1, 1, NOW(), NOW()),
(2, '助行器具', 1, NULL, 1, 2, 1, NOW(), NOW()),
(3, '无障碍扶手', 1, NULL, 1, 3, 1, NOW(), NOW()),
(4, '康复护理床', 1, NULL, 1, 4, 1, NOW(), NOW()),
(5, '康复理疗器材', 1, NULL, 1, 5, 1, NOW(), NOW()),
(6, '护理耗材配件', 1, NULL, 1, 6, 1, NOW(), NOW()),
(7, '共享租赁专区', 2, NULL, 1, 7, 1, NOW(), NOW());

-- ---- 4. 分类树（二级） ----
INSERT INTO categories (id, name, category_type, parent_id, level, sort, status, created_at, updated_at) VALUES
(11, '手动轮椅', 1, 1, 2, 1, 1, NOW(), NOW()),
(12, '电动轮椅', 1, 1, 2, 2, 1, NOW(), NOW()),
(21, '助行拐杖', 1, 2, 2, 1, 1, NOW(), NOW()),
(22, '助行器/学步车', 1, 2, 2, 2, 1, NOW(), NOW()),
(31, '马桶安全扶手', 1, 3, 2, 1, 1, NOW(), NOW()),
(32, '走廊/浴室扶手', 1, 3, 2, 2, 1, NOW(), NOW()),
(41, '手动护理床', 1, 4, 2, 1, 1, NOW(), NOW()),
(42, '电动护理床', 1, 4, 2, 2, 1, NOW(), NOW()),
(51, '四肢训练器', 1, 5, 2, 1, 1, NOW(), NOW()),
(52, '理疗按摩仪', 1, 5, 2, 2, 1, NOW(), NOW()),
(71, '轮椅租赁', 2, 7, 2, 1, 1, NOW(), NOW()),
(72, '助行器租赁', 2, 7, 2, 2, 1, NOW(), NOW()),
(73, '护理床租赁', 2, 7, 2, 3, 1, NOW(), NOW()),
(74, '康复器械租用套餐', 2, 7, 2, 4, 1, NOW(), NOW());

-- ---- 5. 分类树（三级） ----
INSERT INTO categories (id, name, category_type, parent_id, level, sort, status, created_at, updated_at) VALUES
(111, '轻便折叠款', 1, 11, 3, 1, 1, NOW(), NOW()),
(112, '高背全躺款', 1, 11, 3, 2, 1, NOW(), NOW()),
(121, '标准电动', 1, 12, 3, 1, 1, NOW(), NOW()),
(122, '高续航电动', 1, 12, 3, 2, 1, NOW(), NOW()),
(311, '可调节落地扶手', 1, 31, 3, 1, 1, NOW(), NOW()),
(711, '周租低价套餐', 2, 71, 3, 1, 1, NOW(), NOW()),
(712, '月租低价套餐', 2, 71, 3, 2, 1, NOW(), NOW()),
(741, '四肢训练器租用', 2, 74, 3, 1, 1, NOW(), NOW()),
(742, '气压按摩仪租用', 2, 74, 3, 2, 1, NOW(), NOW());

-- ---- 6. 商品（零售 sale_type=1 / 租赁 sale_type=2） ----
INSERT INTO products
(id, merchant_id, category_id, name, description, images, price, original_price, stock, unit,
 product_type, sale_type, rental_unit, rental_price, deposit, max_rental_duration, sales, sort, status, created_at, updated_at)
VALUES
-- 轮椅·轻便折叠
(10001, 1, 111, '铝合金折叠轮椅', '轻量铝合金车架，一键折叠，坐宽45cm，承重100kg，适合居家出行。', '[]', 1280.00, 1580.00, 20, '台', 1, 1, 0, 0.00, 0.00, 0, 12, 1, 1, NOW(), NOW()),
(10002, 1, 111, '轻便便携旅行轮椅', '超轻6kg，可放入后备箱，配旅行袋，出行便携首选。', '[]', 1580.00, 1880.00, 15, '台', 1, 1, 0, 0.00, 0.00, 0, 8, 2, 1, NOW(), NOW()),
-- 轮椅·高背全躺
(10003, 1, 112, '高背全躺轮椅', '大轮高背，靠背可调至全躺，适合长时间坐卧者。', '[]', 1680.00, 2080.00, 12, '台', 1, 1, 0, 0.00, 0.00, 0, 6, 1, 1, NOW(), NOW()),
(10004, 1, 112, '全躺式看护轮椅', '座便两用，可全躺可拆卸腿托，方便护理。', '[]', 1880.00, 2280.00, 10, '台', 1, 1, 0, 0.00, 0.00, 0, 5, 2, 1, NOW(), NOW()),
-- 电动轮椅
(10005, 1, 121, '基础款电动轮椅', '锂电池 12km，可折叠，遥控/手推双模式。', '[]', 3990.00, 4590.00, 8, '台', 1, 1, 0, 0.00, 0.00, 0, 9, 1, 1, NOW(), NOW()),
(10006, 1, 122, '高续航电动轮椅', '锂电池续航20km，防后倾，上下肢驱动助力。', '[]', 5500.00, 6290.00, 6, '台', 1, 1, 0, 0.00, 0.00, 0, 4, 1, 1, NOW(), NOW()),
-- 助行拐杖
(10007, 1, 21, '四脚助行拐杖', '铝合金四脚，10档高度可调，承重125kg。', '[]', 89.00, 109.00, 80, '支', 1, 1, 0, 0.00, 0.00, 0, 20, 1, 1, NOW(), NOW()),
(10008, 1, 21, '可调高度单拐', '轻便单拐，磨砂握把防滑，出街轻巧。', '[]', 59.00, 79.00, 100, '支', 1, 1, 0, 0.00, 0.00, 0, 25, 2, 1, NOW(), NOW()),
(10009, 1, 21, '铝合金肘拐', '肘托承重，适合单侧下肢支撑，康复期常用。', '[]', 129.00, 159.00, 60, '支', 1, 1, 0, 0.00, 0.00, 0, 15, 3, 1, NOW(), NOW()),
-- 助行器/学步车
(10010, 1, 22, '四轮助行器带刹车', '一键刹车，带坐垫可歇脚，适老助行。', '[]', 299.00, 359.00, 30, '个', 1, 1, 0, 0.00, 0.00, 0, 18, 1, 1, NOW(), NOW()),
(10011, 1, 22, '老人三脚助行架', '三角稳定结构，防滑底脚，轻便稳固。', '[]', 249.00, 299.00, 25, '个', 1, 1, 0, 0.00, 0.00, 0, 10, 2, 1, NOW(), NOW()),
(10012, 1, 22, '宝宝学步车', '护栏可拆，静音万向轮，安全学步。', '[]', 199.00, 239.00, 20, '辆', 1, 1, 0, 0.00, 0.00, 0, 14, 3, 1, NOW(), NOW()),
-- 无障碍扶手
(10013, 1, 311, '可调节马桶助力架', '免打孔可调宽，起身支撑更省力。', '[]', 359.00, 419.00, 20, '个', 1, 1, 0, 0.00, 0.00, 0, 11, 1, 1, NOW(), NOW()),
(10014, 1, 31, 'U型马桶安全扶手', '带防滑盖板，不锈钢承重，适老化改造。', '[]', 399.00, 469.00, 18, '个', 1, 1, 0, 0.00, 0.00, 0, 9, 2, 1, NOW(), NOW()),
(10015, 1, 32, '浴室L型扶手', '304不锈钢，承重200kg，浴缸/淋浴墙面安装。', '[]', 129.00, 169.00, 50, '个', 1, 1, 0, 0.00, 0.00, 0, 16, 1, 1, NOW(), NOW()),
(10016, 1, 32, '走廊连续扶手', '墙面长扶手，适老化走廊/过道安装。', '[]', 59.00, 79.00, 100, '米', 1, 1, 0, 0.00, 0.00, 0, 13, 2, 1, NOW(), NOW()),
-- 护理床
(10017, 1, 41, '手动三折护理床', '背/腿/脚多段升降，配护栏，家庭护理。', '[]', 2680.00, 3180.00, 10, '台', 1, 1, 0, 0.00, 0.00, 0, 7, 1, 1, NOW(), NOW()),
(10018, 1, 42, '五功能电动护理床', '背腿升降+翻身+便孔，配防褥疮床垫。', '[]', 6800.00, 7980.00, 6, '台', 1, 1, 0, 0.00, 0.00, 0, 5, 1, 1, NOW(), NOW()),
(10019, 1, 42, '翻身防压疮电动床', '定时翻身，减轻护理负担，适合长期卧床。', '[]', 8990.00, 10900.00, 4, '台', 1, 1, 0, 0.00, 0.00, 0, 3, 2, 1, NOW(), NOW()),
-- 康复理疗器材
(10020, 1, 51, '手指握力康复训练器', '五档阻力，手部精细动作康复。', '[]', 79.00, 99.00, 90, '个', 1, 1, 0, 0.00, 0.00, 0, 22, 1, 1, NOW(), NOW()),
(10021, 1, 51, '肩关节滑轮训练器', '家用滑轮，爬墙训练，肩部术后康复。', '[]', 299.00, 359.00, 30, '个', 1, 1, 0, 0.00, 0.00, 0, 12, 2, 1, NOW(), NOW()),
(10022, 1, 51, '弹力带康复套装', '五色拉力带，全身肌力训练。', '[]', 49.00, 69.00, 120, '套', 1, 1, 0, 0.00, 0.00, 0, 26, 3, 1, NOW(), NOW()),
(10023, 1, 52, '四肢气压按摩仪', '多档气压，促进血液循环，缓解浮肿。', '[]', 1290.00, 1590.00, 15, '台', 1, 1, 0, 0.00, 0.00, 0, 8, 1, 1, NOW(), NOW()),
-- 护理耗材配件
(10024, 1, 6, '防褥疮充气床垫', '交替充气，分散压力，预防压疮。', '[]', 499.00, 599.00, 30, '个', 1, 1, 0, 0.00, 0.00, 0, 17, 1, 1, NOW(), NOW()),
(10025, 1, 6, '轮椅防压疮坐垫', '凝胶减压，透气久坐不闷。', '[]', 129.00, 159.00, 60, '个', 1, 1, 0, 0.00, 0.00, 0, 19, 2, 1, NOW(), NOW()),
(10026, 1, 6, '透气护理垫', '加厚防水，一次性医疗级护理垫。', '[]', 39.00, 49.00, 200, '包', 1, 1, 0, 0.00, 0.00, 0, 30, 3, 1, NOW(), NOW()),

-- 共享租赁·低价套餐（sale_type=2 租赁）
(20001, 1, 711, '手动轮椅·周租套餐', '共享租赁低价体验，按周计费，含押金，到期归还。', '[]', 50.00, 0.00, 5, '台/周', 2, 2, 2, 50.00, 200.00, 12, 6, 1, 1, NOW(), NOW()),
(20002, 1, 712, '手动轮椅·月租套餐', '共享租赁低价月租，长期使用更划算。', '[]', 80.00, 0.00, 8, '台/月', 2, 2, 3, 80.00, 200.00, 6, 9, 1, 1, NOW(), NOW()),
(20003, 1, 72, '四轮助行器·月租套餐', '带刹车助行器按月租，康复期安心用。', '[]', 60.00, 0.00, 6, '个/月', 2, 2, 3, 60.00, 100.00, 6, 7, 1, 1, NOW(), NOW()),
(20004, 1, 73, '手动护理床·月租套餐', '护理床按月租，含安装指导，居家照护。', '[]', 300.00, 0.00, 4, '台/月', 2, 2, 3, 300.00, 800.00, 6, 4, 1, 1, NOW(), NOW()),
(20005, 1, 741, '四肢联动训练器·月租套餐', '共享康复器械，按月租用，配合康复指导。', '[]', 200.00, 0.00, 5, '台/月', 2, 2, 3, 200.00, 500.00, 6, 3, 1, 1, NOW(), NOW()),
(20006, 1, 742, '气压按摩仪·周租套餐', '康复理疗器械低价租用，促血液循环。', '[]', 60.00, 0.00, 6, '台/周', 2, 2, 2, 60.00, 300.00, 12, 5, 1, 1, NOW(), NOW());

-- ---- 7. 更新商家为医疗商（单店 id=1） ----
UPDATE merchants
SET name = '乐享辅具（云南财旭商贸）',
    business_category = '二类医疗器械/康复辅具 与 共享租赁',
    business_hours = '08:00-21:00',
    announcement = '全品类二类医疗辅具，支持零售与共享租赁低价套餐，专业适老康复服务。',
    status = 1
WHERE id = 1;