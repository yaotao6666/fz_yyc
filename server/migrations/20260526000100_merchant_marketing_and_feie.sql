-- 商家满减规则与飞鹅打印机字段扩展

CREATE TABLE IF NOT EXISTS merchant_full_reduction_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '规则ID',
  merchant_id BIGINT UNSIGNED NOT NULL COMMENT '所属商家ID',
  threshold_amount DECIMAL(10,2) NOT NULL COMMENT '满减门槛金额(元)',
  discount_amount DECIMAL(10,2) NOT NULL COMMENT '减免金额(元)',
  sort INT NOT NULL DEFAULT 0 COMMENT '排序值',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=停用',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_full_reduction_merchant_id (merchant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家自动满减规则表';

ALTER TABLE cloud_printers
  ADD COLUMN feie_user VARCHAR(64) NULL COMMENT '飞鹅账号' AFTER api_url,
  ADD COLUMN feie_ukey VARCHAR(128) NULL COMMENT '飞鹅UKey' AFTER feie_user,
  ADD COLUMN feie_sn VARCHAR(64) NULL COMMENT '飞鹅打印机终端号' AFTER feie_ukey;
