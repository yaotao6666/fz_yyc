-- 021_drop_activities.sql
-- 下线游离的平台活动表 activities（无任何路由/handler/service 引用，促销已统一走优惠券体系 coupon_templates）
SET NAMES utf8mb4;
DROP TABLE IF EXISTS activities;