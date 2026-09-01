-- ============================================================
-- 008_service_safety.sql — 阶段三：服务过程安全
-- ------------------------------------------------------------
-- 说明：
--  1. service_staffs 增列 service_region / quality_score（幂等）
--  2. 新建 service_records / service_location_tracks / service_alert_events（幂等）
--  3. 新建 agreements / agreement_consents 协议管理表（幂等）
--  4. RBAC：预警中心（顶级）+ 系统管理下协议管理菜单与按钮权限节点
-- ============================================================

SET NAMES utf8mb4;

-- ============================================================
-- 1. service_staffs 增列
-- ============================================================
SET @_col = (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_staffs' AND COLUMN_NAME='service_region');
SET @_sql = IF(@_col = 0,
  "ALTER TABLE service_staffs ADD COLUMN service_region VARCHAR(256) DEFAULT NULL COMMENT '服务区域(区县,逗号分隔,空=不限)' AFTER qualifications",
  'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_staffs' AND COLUMN_NAME='quality_score');
SET @_sql = IF(@_col = 0,
  "ALTER TABLE service_staffs ADD COLUMN quality_score DECIMAL(3,1) NOT NULL DEFAULT 5.0 COMMENT '服务质量分(阶段四写入)' AFTER service_region",
  'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. 服务记录表 service_records
-- ============================================================
SET @_tbl = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_records');
SET @_sql = IF(@_tbl > 0, 'SELECT 1',
"CREATE TABLE service_records (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '服务记录ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  staff_id BIGINT UNSIGNED NOT NULL COMMENT '服务人员ID',
  start_time DATETIME DEFAULT NULL COMMENT '签到时间(orders快照)',
  end_time DATETIME DEFAULT NULL COMMENT '签退时间(orders快照)',
  gps_track_url VARCHAR(512) DEFAULT NULL COMMENT '轨迹聚合文件URL(可选)',
  audio_url VARCHAR(512) DEFAULT NULL COMMENT '服务录音URL(七牛)',
  audio_uploaded_at DATETIME DEFAULT NULL COMMENT '录音上传时间(30天清理依据)',
  audio_deleted_at DATETIME DEFAULT NULL COMMENT '录音删除标记时间',
  sos_triggered TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否触发SOS: 0=否 1=是',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=异常',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY uk_service_records_order_id (order_id),
  KEY idx_service_records_staff_id (staff_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务记录表'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 3. 轨迹点表 service_location_tracks（高频写）
-- ============================================================
SET @_tbl = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_location_tracks');
SET @_sql = IF(@_tbl > 0, 'SELECT 1',
"CREATE TABLE service_location_tracks (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '轨迹点ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  staff_id BIGINT UNSIGNED NOT NULL COMMENT '服务人员ID',
  lat DECIMAL(10,6) NOT NULL COMMENT '纬度',
  lng DECIMAL(10,6) NOT NULL COMMENT '经度',
  reported_at DATETIME NOT NULL COMMENT '上报时间',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  KEY idx_tracks_order_id (order_id),
  KEY idx_tracks_staff_id (staff_id),
  KEY idx_tracks_reported_at (reported_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务轨迹点表'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 4. 预警事件表 service_alert_events
-- ============================================================
SET @_tbl = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_alert_events');
SET @_sql = IF(@_tbl > 0, 'SELECT 1',
"CREATE TABLE service_alert_events (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '预警事件ID',
  order_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联订单ID(可空)',
  staff_id BIGINT UNSIGNED NOT NULL COMMENT '服务人员ID',
  alert_type TINYINT UNSIGNED NOT NULL COMMENT '预警类型: 1=SOS求助 2=服务超时未结束',
  lat DECIMAL(10,6) DEFAULT NULL COMMENT '触发位置纬度',
  lng DECIMAL(10,6) DEFAULT NULL COMMENT '触发位置经度',
  address VARCHAR(256) DEFAULT NULL COMMENT '触发位置地址',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=待处理 2=处理中 3=已处理',
  handler_id BIGINT UNSIGNED DEFAULT NULL COMMENT '处理人ID(merchant_staffs.id)',
  handler_name VARCHAR(64) DEFAULT NULL COMMENT '处理人姓名快照',
  handle_remark VARCHAR(512) DEFAULT NULL COMMENT '处理备注留痕',
  handled_at DATETIME DEFAULT NULL COMMENT '处理时间',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  KEY idx_alert_order_id (order_id),
  KEY idx_alert_staff_id (staff_id),
  KEY idx_alert_status (status),
  KEY idx_alert_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务预警事件表'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 5. 协议表 agreements
-- ============================================================
SET @_tbl = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='agreements');
SET @_sql = IF(@_tbl > 0, 'SELECT 1',
"CREATE TABLE agreements (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '协议ID',
  type TINYINT UNSIGNED NOT NULL COMMENT '协议类型: 1=用户协议 2=隐私政策 3=录音/定位授权协议',
  title VARCHAR(128) NOT NULL COMMENT '协议标题',
  content TEXT COMMENT '协议正文(富文本)',
  version VARCHAR(32) NOT NULL COMMENT '版本号(同类型递增,如v1.2)',
  status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 1=已发布(当前生效) 0=草稿/停用',
  published_at DATETIME DEFAULT NULL COMMENT '发布时间',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  KEY idx_agreements_type (type),
  KEY idx_agreements_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='协议表'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 6. 协议同意留痕表 agreement_consents
-- ============================================================
SET @_tbl = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='agreement_consents');
SET @_sql = IF(@_tbl > 0, 'SELECT 1',
"CREATE TABLE agreement_consents (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '留痕ID',
  agreement_id BIGINT UNSIGNED NOT NULL COMMENT '协议ID',
  user_type TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '用户类型: 1=C端用户 2=服务人员',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  version VARCHAR(32) NOT NULL COMMENT '同意时协议版本号快照',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '同意时间',
  KEY idx_consents_agreement_id (agreement_id),
  KEY idx_consents_user (user_type, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='协议同意留痕表'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 7. RBAC 菜单节点
-- 预警中心（顶级 id=97）；协议管理（系统管理 id=7 下 id=76）
-- ============================================================
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(97, 0, 1, '预警中心', '/alert-events', NULL, 10, 1, 1, 'alert-events:view'),
(971, 97, 2, '处理预警', NULL, NULL, 1, 1, 1, 'alert-events:update'),
(76, 7, 1, '协议管理', '/system/agreements', NULL, 5, 1, 1, 'agreements:view'),
(761, 76, 2, '新建协议', NULL, NULL, 1, 1, 1, 'agreements:create'),
(762, 76, 2, '编辑协议', NULL, NULL, 2, 1, 1, 'agreements:update'),
(763, 76, 2, '发布协议', NULL, NULL, 3, 1, 1, 'agreements:update');

-- 授予超管角色（role_id=1）新菜单权限
INSERT IGNORE INTO sys_role_menus (role_id, menu_id)
SELECT 1, id FROM sys_menus WHERE id IN (97, 971, 76, 761, 762, 763);
