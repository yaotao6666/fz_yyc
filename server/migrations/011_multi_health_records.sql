-- 阶段五 8.2 多档案（一个账号绑定多位老人）
-- 改造：
--   1) health_records：user_id 唯一索引 → 普通索引；新增 relation 字段
--   2) health_assessments：新增 record_id（档案ID）列 + 索引
--   3) fitting_recommendations：新增 record_id（档案ID）列 + 索引
--   存量：health_records 为空，测试数据无需回填。
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 1) health_records 多档案化
ALTER TABLE health_records
  DROP INDEX uk_health_records_user_id,
  ADD INDEX idx_health_records_user_id (user_id),
  ADD COLUMN relation TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '与账号关系:1本人2父母3其他亲属' AFTER user_id;

-- 2) health_assessments 挂档案维度
ALTER TABLE health_assessments
  ADD COLUMN record_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联档案ID（多档案）' AFTER user_id,
  ADD INDEX idx_health_assessments_record_id (record_id);

-- 3) fitting_recommendations 挂档案维度
ALTER TABLE fitting_recommendations
  ADD COLUMN record_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联档案ID（多档案）' AFTER user_id,
  ADD INDEX idx_fitting_recommendations_record_id (record_id);

SET FOREIGN_KEY_CHECKS = 1;
