-- 018_drop_announcements.sql
-- 移除游离的系统公告表（首页公告已使用 merchants.announcement 文本字段）
SET NAMES utf8mb4;
DROP TABLE IF EXISTS announcements;