-- ============================================================
-- 001_schema.sql  — 财旭商贸（乐享辅具）数据库结构（合并最终版）
-- ------------------------------------------------------------
-- 说明：由历史迁移脚本全部合并后生成，以代码模型 + 最新脚本对齐，
--       含去 merchant_id / service_provider_id 后的最终表结构。
-- 表数量：35 张（本文件仅含 CREATE TABLE，不含基础数据）
-- 执行：先执行本文件建表，再执行 002_seed.sql 导入基础种子数据。
-- 约束：幂等（每张表 DROP IF EXISTS + CREATE），可重复执行。
-- 已清理残留列：各业务表 merchant_id、activities/announcements/
--       merchants 的 service_provider_id（与原脚本 remove_merchant_id、
--       cxsm_clean_and_expand 的最终意图一致）。
-- ============================================================
-- 源库 MySQL 8.0.46

-- MySQL dump 起始标记

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `activities`
--

DROP TABLE IF EXISTS `activities`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `activities` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL,
  `title` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `image` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `link_type` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `link_value` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort` int unsigned NOT NULL DEFAULT '0',
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_activities_type` (`type`),
  KEY `idx_activities_status` (`status`),
  KEY `idx_activities_sort` (`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台活动表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `announcements`
--

DROP TABLE IF EXISTS `announcements`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `announcements` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_announcements_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统公告表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `care_plans`
--

DROP TABLE IF EXISTS `care_plans`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `care_plans` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '居民用户ID',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '计划名称',
  `plan_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '类型:1生活照料2基础护理3康复训练4综合康养',
  `start_date` date DEFAULT NULL COMMENT '开始日期',
  `end_date` date DEFAULT NULL COMMENT '结束日期',
  `frequency` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '照护频次',
  `goals` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '照护目标',
  `items` json DEFAULT NULL COMMENT '护理项配置[{name,desc}]',
  `assigned_staff_id` bigint unsigned DEFAULT NULL COMMENT '指派服务人员ID',
  `order_id` bigint unsigned DEFAULT NULL COMMENT '关联服务订单ID(可空)',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态:0草稿1执行中2已暂停3已完成',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_care_plans_user_id` (`user_id`),
  KEY `idx_care_plans_assigned_staff_id` (`assigned_staff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='照护计划表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `care_visits`
--

DROP TABLE IF EXISTS `care_visits`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `care_visits` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `plan_id` bigint unsigned DEFAULT NULL COMMENT '关联照护计划ID(可空)',
  `order_id` bigint unsigned DEFAULT NULL COMMENT '关联服务订单ID(可空)',
  `user_id` bigint unsigned NOT NULL COMMENT '居民用户ID',
  `staff_id` bigint unsigned NOT NULL COMMENT '录入服务人员ID',
  `visit_at` datetime DEFAULT NULL COMMENT '到访时间',
  `nursing_items` json DEFAULT NULL COMMENT '完成的护理项[{name,done,remark}]',
  `vitals` json DEFAULT NULL COMMENT '生命体征{blood_pressure,blood_glucose,heart_rate,oxygen,weight}',
  `photos` json DEFAULT NULL COMMENT '照片URL数组',
  `remark` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `follow_up_advice` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '下次随访建议',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_care_visits_plan_id` (`plan_id`),
  KEY `idx_care_visits_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上门照护记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `categories`
--

DROP TABLE IF EXISTS `categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `category_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '分类类型: 1=商品分类 2=服务分类',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父分类ID(空=一级)',
  `level` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '层级: 1=一级 2=二级 3=三级',
  `sort` int unsigned NOT NULL DEFAULT '0',
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_categories_parent` (`parent_id`),
  KEY `idx_categories_sort` (`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=743 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品分类表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fitting_recommendations`
--

DROP TABLE IF EXISTS `fitting_recommendations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `fitting_recommendations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '居民用户ID',
  `assessment_id` bigint unsigned DEFAULT NULL COMMENT '关联评估记录ID',
  `symptom_desc` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '症状/需求描述',
  `fitting_result` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '适配结论',
  `recommended_products` json DEFAULT NULL COMMENT '推荐商品快照[{product_id,name,reason,sale_type}]',
  `staff_id` bigint unsigned DEFAULT NULL COMMENT '生成建议的服务人员ID',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态:0草稿1已确认2已下单',
  `order_id` bigint unsigned DEFAULT NULL COMMENT '关联订单ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_fitting_recommendations_user_id` (`user_id`),
  KEY `idx_fitting_recommendations_staff_id` (`staff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='康复辅具适配建议表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `follow_up_tasks`
--

DROP TABLE IF EXISTS `follow_up_tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `follow_up_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '居民用户ID',
  `task_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '类型:1康复随访2租后回访3慢病随访4评估回访',
  `source_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '来源:1服务完成2租赁归还3评估完成4手动',
  `source_id` bigint unsigned DEFAULT NULL COMMENT '来源ID(订单/评估ID)',
  `plan_follow_time` datetime DEFAULT NULL COMMENT '计划随访时间',
  `staff_id` bigint unsigned DEFAULT NULL COMMENT '执行服务人员ID(可空=待认领)',
  `contact_method` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '随访方式:1电话2上门3微信',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态:0待执行1已完成2已跳过',
  `result` json DEFAULT NULL COMMENT '随访结果{contact_method,content,education_article_ids,satisfaction,remark}',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `remark` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_follow_up_tasks_user_id` (`user_id`),
  KEY `idx_follow_up_tasks_staff_id` (`staff_id`),
  KEY `idx_follow_up_tasks_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='随访任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `health_assessment_forms`
--

DROP TABLE IF EXISTS `health_assessment_forms`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_assessment_forms` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '量表名称',
  `dimension` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '评估维度',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '量表说明',
  `questions` json DEFAULT NULL COMMENT '题目数组[{key,title,options:[{label,score}]}]',
  `score_rule` json DEFAULT NULL COMMENT '评分规则[{min,max,level,conclusion}]',
  `version` int unsigned NOT NULL DEFAULT '1' COMMENT '版本号',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '0草稿1启用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='健康评估量表表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `health_assessments`
--

DROP TABLE IF EXISTS `health_assessments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_assessments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `form_id` bigint unsigned NOT NULL COMMENT '量表ID',
  `form_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '量表名称快照',
  `assessor_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '1自助2服务人员',
  `staff_id` bigint unsigned DEFAULT NULL COMMENT '评估服务人员ID',
  `answers` json DEFAULT NULL COMMENT '答案{key:选中label}',
  `total_score` decimal(6,1) DEFAULT NULL COMMENT '总分',
  `level` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '评估等级',
  `conclusion` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '评估结论',
  `suggestions` json DEFAULT NULL COMMENT '建议数组',
  `symptom_desc` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '症状描述',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_health_assessments_user_id` (`user_id`),
  KEY `idx_health_assessments_form_id` (`form_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='健康评估记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `health_education_articles`
--

DROP TABLE IF EXISTS `health_education_articles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_education_articles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标题',
  `category` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分类',
  `cover` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '封面图URL',
  `content` text COLLATE utf8mb4_unicode_ci COMMENT '正文',
  `tags` json DEFAULT NULL COMMENT '定向慢病标签数组',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态:0草稿1发布',
  `publish_at` datetime DEFAULT NULL COMMENT '发布时间',
  `views` int unsigned NOT NULL DEFAULT '0' COMMENT '浏览量',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='健康宣教内容表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `health_monitoring`
--

DROP TABLE IF EXISTS `health_monitoring`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_monitoring` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `record_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '类型:1血压2血糖3心率4血氧5体重',
  `value` decimal(8,2) NOT NULL DEFAULT '0.00' COMMENT '测量值',
  `unit` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '单位',
  `extra` json DEFAULT NULL COMMENT '扩展(如血压高低压)',
  `recorded_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT '录入人(0=用户本人)',
  `recorded_at` datetime DEFAULT NULL COMMENT '测量时间',
  `remark` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_health_monitoring_user_id` (`user_id`),
  KEY `idx_health_monitoring_record_type` (`record_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生命体征监测表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `health_records`
--

DROP TABLE IF EXISTS `health_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `real_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '真实姓名',
  `gender` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '性别:1男2女',
  `birth_date` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '出生日期',
  `id_card` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '身份证号',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
  `emergency_contact` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '紧急联系人',
  `emergency_phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '紧急联系电话',
  `address` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '常住地址',
  `height_cm` decimal(5,1) DEFAULT NULL COMMENT '身高(cm)',
  `weight_kg` decimal(5,1) DEFAULT NULL COMMENT '体重(kg)',
  `blood_type` varchar(8) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '血型',
  `past_history` json DEFAULT NULL COMMENT '既往病史数组',
  `allergy_history` json DEFAULT NULL COMMENT '过敏史数组',
  `family_history` json DEFAULT NULL COMMENT '家族病史数组',
  `surgery_history` json DEFAULT NULL COMMENT '手术史数组',
  `medication_list` json DEFAULT NULL COMMENT '长期用药数组',
  `chronic_tags` json DEFAULT NULL COMMENT '慢病标签数组',
  `smoking` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '吸烟情况',
  `drinking` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '饮酒情况',
  `assessment_level` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最近一次评估等级',
  `remark` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '0未建档1正常2已归档',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_health_records_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='居民健康档案表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `merchant_delivery_settings`
--

DROP TABLE IF EXISTS `merchant_delivery_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_delivery_settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `base_fee` decimal(10,2) NOT NULL DEFAULT '0.00',
  `free_delivery_amount` decimal(10,2) NOT NULL DEFAULT '0.00',
  `max_distance` int unsigned NOT NULL DEFAULT '10',
  `distance_rules` json DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家配送设置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `merchant_fees`
--

DROP TABLE IF EXISTS `merchant_fees`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_fees` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `year` int unsigned NOT NULL,
  `amount` decimal(10,2) NOT NULL DEFAULT '0.00',
  `status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL,
  `pay_time` datetime DEFAULT NULL,
  `free_reason` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家年费表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `merchant_rates`
--

DROP TABLE IF EXISTS `merchant_rates`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_rates` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `rate_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `rate` decimal(5,4) NOT NULL,
  `effective_time` datetime NOT NULL,
  `expire_time` datetime DEFAULT NULL,
  `remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_merchant_rates_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家手续费率表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `merchant_staff_roles`
--

DROP TABLE IF EXISTS `merchant_staff_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_staff_roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `staff_id` bigint unsigned NOT NULL COMMENT '员工ID(merchant_staffs.id)',
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID(sys_roles.id)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_merchant_staff_roles` (`staff_id`,`role_id`)
) ENGINE=InnoDB AUTO_INCREMENT=33 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='员工角色关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `merchant_staffs`
--

DROP TABLE IF EXISTS `merchant_staffs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_staffs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `unionid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信UnionID',
  `wechat_bound_at` datetime DEFAULT NULL,
  `role` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'staff',
  `notify_enabled` tinyint(1) NOT NULL DEFAULT '1',
  `browse_notify_enabled` tinyint(1) NOT NULL DEFAULT '1',
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `department_id` bigint unsigned DEFAULT NULL COMMENT '部门ID(sys_departments.id)',
  `last_login_at` datetime DEFAULT NULL,
  `last_wechat_login_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_merchant_staffs_username` (`username`),
  KEY `idx_merchant_staffs_openid` (`openid`),
  KEY `idx_merchant_staffs_phone` (`phone`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家员工表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `merchants`
--

DROP TABLE IF EXISTS `merchants`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchants` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '商家名称',
  `logo` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '商家Logo图片地址(七牛私有路径)',
  `contact_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系人姓名',
  `contact_phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话(用于用户端拨打退款)',
  `contact_email` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系邮箱',
  `address` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '商家地址',
  `lat` decimal(10,6) DEFAULT NULL COMMENT '纬度',
  `lng` decimal(10,6) DEFAULT NULL COMMENT '经度',
  `business_category` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '经营类目',
  `business_hours` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '营业时间描述',
  `announcement` text COLLATE utf8mb4_unicode_ci COMMENT '商家公告',
  `sub_mch_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信支付子商户号(线下进件后回填)',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '营业状态: 1=营业中 0=休息中',
  `rating` decimal(2,1) NOT NULL DEFAULT '5.0' COMMENT '商家评分(1.0-5.0)',
  `sales_count` bigint unsigned NOT NULL DEFAULT '0' COMMENT '累计销量',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `cover_image` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '商家背景/封面图地址(七牛私有路径)',
  `payment_config_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '支付配置状态: 0=未完成配置 1=已完成配置(已回填sub_mch_id)',
  `profit_sharing_enabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否开启自动分账',
  PRIMARY KEY (`id`),
  KEY `idx_merchants_sub_mch_id` (`sub_mch_id`),
  KEY `idx_merchants_status` (`status`),
  KEY `idx_merchants_location` (`lat`,`lng`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `order_items`
--

DROP TABLE IF EXISTS `order_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `order_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL,
  `product_id` bigint unsigned NOT NULL,
  `product_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `image` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `price` decimal(10,2) NOT NULL,
  `quantity` int unsigned NOT NULL DEFAULT '1',
  `spec_info` json DEFAULT NULL,
  `subtotal` decimal(10,2) NOT NULL,
  `sale_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '销售类型快照: 1=一口价 2=租赁',
  `rental_unit` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '租赁计费周期快照',
  `rental_duration` int unsigned NOT NULL DEFAULT '0' COMMENT '租赁时长(下单时选择)',
  `unit_rental_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '单位租金快照',
  `rental_subtotal` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '租金小计=单价×时长',
  `deposit` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '单商品押金快照',
  `deposit_deduct` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '押金扣除金额(损坏赔偿)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order_items_order_id` (`order_id`),
  KEY `idx_order_items_product_id` (`product_id`),
  CONSTRAINT `fk_order_items_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_order_items_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB AUTO_INCREMENT=23 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `orders`
--

DROP TABLE IF EXISTS `orders`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_no` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单编号',
  `user_id` bigint unsigned NOT NULL COMMENT '下单用户ID',
  `order_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '订单类型: 1=零售 2=租赁 3=康养上门 4=陪诊 5=科普体验 6=长护险服务',
  `total_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '商品总金额(元)',
  `delivery_fee` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '配送费(元)',
  `discount_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '优惠减免金额(元)',
  `pay_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '实付金额(元)=总金额+配送费-优惠',
  `total_deposit` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '总押金(元)',
  `deposit_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '押金状态: 0=无押金 1=已收 2=已退 3=部分扣除',
  `deposit_refund_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '押金退还金额',
  `deposit_deduct_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '押金扣除总额(损坏赔偿)',
  `deposit_refunded_at` datetime DEFAULT NULL COMMENT '押金退还时间',
  `rental_returned_at` datetime DEFAULT NULL COMMENT '租赁归还时间',
  `rental_return_remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '归还备注(验机情况)',
  `delivery_address` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收货地址(配送时填写)',
  `contact_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系人姓名(配送时填写)',
  `contact_phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话(配送时填写)',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '订单状态: 1=待支付 2=已支付 3=已完成 4=已取消 5=退款中 6=已退款',
  `biz_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '业务子状态(按order_type语义不同)',
  `remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户备注',
  `transaction_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信支付交易单号',
  `paid_at` datetime(3) DEFAULT NULL COMMENT '支付完成时间',
  `pay_notify_payload` json DEFAULT NULL COMMENT '微信支付回调原始报文JSON',
  `profit_sharing_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '分账状态:0未分账1分账中2分账成功3分账失败4已跳过',
  `profit_sharing_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '分账总额(元)',
  `profit_sharing_order_no` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信分账单号',
  `profit_sharing_at` datetime DEFAULT NULL COMMENT '分账完成时间',
  `profit_sharing_error` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分账失败原因',
  `completed_at` datetime(3) DEFAULT NULL COMMENT '核销完成时间',
  `cancelled_at` datetime(3) DEFAULT NULL COMMENT '取消时间',
  `refunded_at` datetime(3) DEFAULT NULL COMMENT '退款完成时间(status=6时写入)',
  `scheduled_at` datetime(3) DEFAULT NULL COMMENT '预约服务开始时间(康养/陪诊/科普)',
  `assigned_staff_id` bigint unsigned DEFAULT NULL COMMENT '指派的服务人员ID(service_staffs.id)',
  `actual_started_at` datetime(3) DEFAULT NULL COMMENT '实际服务开始时间(签到时间)',
  `actual_ended_at` datetime(3) DEFAULT NULL COMMENT '实际服务结束时间(签退时间)',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_orders_order_no` (`order_no`),
  KEY `idx_orders_user_id` (`user_id`),
  KEY `idx_orders_status` (`status`),
  KEY `idx_orders_created_at` (`created_at`),
  KEY `idx_orders_paid_at` (`paid_at`),
  KEY `idx_orders_deposit_status` (`deposit_status`),
  KEY `idx_orders_assigned_staff` (`assigned_staff_id`,`status`),
  KEY `idx_orders_order_type` (`order_type`,`status`),
  KEY `idx_orders_scheduled_at` (`scheduled_at`),
  CONSTRAINT `fk_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `product_specs`
--

DROP TABLE IF EXISTS `product_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `product_specs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `product_id` bigint unsigned NOT NULL,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `options` json DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_product_specs_product_id` (`product_id`),
  CONSTRAINT `fk_product_specs_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=3088 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品规格表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `products`
--

DROP TABLE IF EXISTS `products`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `products` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `category_id` bigint unsigned DEFAULT NULL,
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `images` json DEFAULT NULL,
  `price` decimal(10,2) NOT NULL,
  `original_price` decimal(10,2) DEFAULT NULL,
  `stock` int unsigned NOT NULL DEFAULT '0',
  `unit` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '份',
  `product_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '商品类型: 1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务 5=科普活动 6=长护险服务',
  `service_content` json DEFAULT NULL COMMENT '服务型商品的内容描述JSON',
  `sale_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '销售类型: 1=一口价 2=租赁',
  `rental_unit` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '租赁计费周期: 0=非租赁 1=按天 2=按周 3=按月',
  `rental_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '单位租金(元)',
  `deposit` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '押金(元)',
  `max_rental_duration` int unsigned NOT NULL DEFAULT '0' COMMENT '最大租赁时长(0=不限)',
  `sales` int unsigned NOT NULL DEFAULT '0',
  `sort` int unsigned NOT NULL DEFAULT '0',
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `deleted_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_products_category_id` (`category_id`),
  KEY `idx_products_status` (`status`),
  KEY `idx_products_sale_type` (`sale_type`),
  KEY `idx_products_sales` (`sales`),
  KEY `idx_products_sort` (`sort`),
  KEY `idx_products_product_type` (`product_type`,`status`),
  CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=20007 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `profit_sharing_receivers`
--

DROP TABLE IF EXISTS `profit_sharing_receivers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `profit_sharing_receivers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL DEFAULT '1' COMMENT '商家ID(单商户恒为1)',
  `receiver_type` tinyint unsigned NOT NULL COMMENT '接收方类型: 1=商户号 2=个人微信openid',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '显示名称',
  `account` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '接收方账号(商户号 或 个人openid)',
  `personal_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '个人真实姓名(个人类型时, 微信实名校验, 可空)',
  `relation_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'SERVICE_PROVIDER' COMMENT '与特约商户关系, 默认SERVICE_PROVIDER',
  `default_ratio` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '自动分账默认比例(%)',
  `wechat_bound` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '微信接收方关系是否已建立: 0=未建立 1=已建立',
  `wechat_error` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信建立接收方关系失败原因',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=启用 0=停用',
  `sort` int unsigned NOT NULL DEFAULT '0' COMMENT '排序值(越小越靠前)',
  `remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_profit_sharing_receivers_merchant_id` (`merchant_id`),
  KEY `idx_profit_sharing_receivers_account` (`account`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账接收方表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `profit_sharing_record_receivers`
--

DROP TABLE IF EXISTS `profit_sharing_record_receivers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `profit_sharing_record_receivers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `record_id` bigint unsigned NOT NULL COMMENT '分账单ID',
  `receiver_id` bigint unsigned DEFAULT NULL COMMENT '接收方ID',
  `receiver_type` tinyint unsigned NOT NULL COMMENT '接收方类型快照: 1=商户号 2=个人微信openid',
  `receiver_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '接收方名称快照',
  `account` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '接收方账号快照',
  `amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '分账金额(元)',
  `result_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信分账结果: PROCESSING/SUCCESS/CLOSED/FAILED/FINISHED',
  `detail_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信分账明细单号',
  `fail_reason` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分账失败原因',
  `finish_time` datetime DEFAULT NULL COMMENT '分账完成时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_profit_sharing_record_receivers_record_id` (`record_id`),
  KEY `idx_profit_sharing_record_receivers_receiver_id` (`receiver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账单明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `profit_sharing_records`
--

DROP TABLE IF EXISTS `profit_sharing_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `profit_sharing_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL DEFAULT '1' COMMENT '商家ID(单商户恒为1)',
  `order_id` bigint unsigned NOT NULL COMMENT '订单ID',
  `order_no` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单编号',
  `sp_mchid` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '服务商商户号',
  `sub_mchid` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '特约商户号(分账出资方)',
  `appid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分账请求使用的appid(特约商户主体小程序)',
  `transaction_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信支付交易单号',
  `out_order_no` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '微信分账单号',
  `total_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '订单实付金额(元)',
  `total_share_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '本次分账总额(元)',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态: 0=待分账 1=分账中 2=分账成功 3=分账失败 4=已跳过',
  `share_time` datetime DEFAULT NULL COMMENT '分账完成时间',
  `error_message` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '失败原因',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_profit_sharing_records_merchant_id` (`merchant_id`),
  KEY `idx_profit_sharing_records_order_id` (`order_id`),
  KEY `idx_profit_sharing_records_order_no` (`order_no`),
  KEY `idx_profit_sharing_records_out_order_no` (`out_order_no`),
  KEY `idx_profit_sharing_records_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `refunds`
--

DROP TABLE IF EXISTS `refunds`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `refunds` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL,
  `refund_no` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `refund_amount` decimal(10,2) NOT NULL,
  `refund_reason` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned NOT NULL DEFAULT '0',
  `refund_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `refunded_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refunds_refund_no` (`refund_no`),
  KEY `idx_refunds_order_id` (`order_id`),
  KEY `idx_refunds_status` (`status`),
  CONSTRAINT `fk_refunds_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB AUTO_INCREMENT=29 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `service_staffs`
--

DROP TABLE IF EXISTS `service_staffs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `service_staffs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '服务人员ID',
  `username` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '登录用户名',
  `password` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '加密密码(不返回)',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '姓名',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号',
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信OpenID(用于快捷登录)',
  `avatar` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '头像URL',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态: 0=待审核 1=启用 2=禁用',
  `last_login_at` datetime(3) DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_service_staffs_open_id` (`openid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_departments`
--

DROP TABLE IF EXISTS `sys_departments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_departments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父部门ID(0=顶级)',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '部门名称',
  `leader` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '负责人',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
  `sort` int unsigned NOT NULL DEFAULT '0' COMMENT '排序值',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=启用 0=禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sys_departments_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统部门表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_menus`
--

DROP TABLE IF EXISTS `sys_menus`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父菜单ID(0=顶级)',
  `menu_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '类型: 1=菜单/目录 2=按钮',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单名称',
  `path` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '前端路由路径(菜单时)',
  `icon` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '菜单图标',
  `sort` int unsigned NOT NULL DEFAULT '0' COMMENT '排序值(越小越靠前)',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=启用 0=禁用',
  `visible` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '是否显示: 1=显示 0=隐藏',
  `permission` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '权限标识(如 order:view)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sys_menus_parent_id` (`parent_id`),
  KEY `idx_sys_menus_permission` (`permission`)
) ENGINE=InnoDB AUTO_INCREMENT=923 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统菜单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_role_menus`
--

DROP TABLE IF EXISTS `sys_role_menus`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_role_menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `menu_id` bigint unsigned NOT NULL COMMENT '菜单ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_role_menus` (`role_id`,`menu_id`)
) ENGINE=InnoDB AUTO_INCREMENT=112 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_roles`
--

DROP TABLE IF EXISTS `sys_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
  `code` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色编码(唯一)',
  `remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=启用 0=禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_roles_code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统角色表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_addresses`
--

DROP TABLE IF EXISTS `user_addresses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_addresses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `province` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `city` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `district` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `address` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
  `lat` decimal(10,6) DEFAULT NULL,
  `lng` decimal(10,6) DEFAULT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_addresses_user_id` (`user_id`),
  CONSTRAINT `fk_user_addresses_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=111 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户收货地址表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_behavior_events`
--

DROP TABLE IF EXISTS `user_behavior_events`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_behavior_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信OpenID',
  `event_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件类型: page_view=页面浏览 product_view=商品查看 submit_order=提交订单 pay_success=支付成功',
  `page` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '页面标识(如store_home/store_product)',
  `product_id` bigint unsigned DEFAULT NULL COMMENT '关联商品ID(商品查看事件)',
  `order_id` bigint unsigned DEFAULT NULL COMMENT '关联订单ID(下单/支付事件)',
  `source` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '事件来源: scan=扫码 direct=直接进入',
  `payload` json DEFAULT NULL COMMENT '事件附加数据JSON',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_behavior_events_user_id` (`user_id`),
  KEY `idx_user_behavior_events_openid` (`openid`),
  KEY `idx_user_behavior_events_event_type` (`event_type`),
  KEY `idx_user_behavior_events_product_id` (`product_id`),
  KEY `idx_user_behavior_events_order_id` (`order_id`),
  KEY `idx_user_behavior_events_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=671 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户行为事件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_visits`
--

DROP TABLE IF EXISTS `user_visits`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_visits` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信OpenID',
  `visit_time` datetime(3) DEFAULT NULL COMMENT '访问时间',
  `source` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '访问来源: scan=扫码 direct=直接进入',
  PRIMARY KEY (`id`),
  KEY `idx_user_visits_user_id` (`user_id`),
  KEY `idx_user_visits_openid` (`openid`),
  CONSTRAINT `fk_user_visits_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=252 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户访问记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信OpenID(用户唯一标识)',
  `union_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信UnionID(跨小程序唯一)',
  `nickname` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户昵称(默认微信用户)',
  `avatar` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户头像URL',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户手机号',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=正常 0=禁用',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `first_visit_at` datetime(3) DEFAULT NULL COMMENT '首次访问时间',
  `last_visit_at` datetime(3) DEFAULT NULL COMMENT '最后访问时间',
  `visit_count` bigint unsigned NOT NULL DEFAULT '1' COMMENT '累计访问次数',
  `has_ordered` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否下过单: true=是 false=否',
  `total_orders` bigint unsigned NOT NULL DEFAULT '0' COMMENT '累计订单数',
  `total_spent` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '累计消费金额(元)',
  `has_paid` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否完成过支付: true=是 false=否',
  `first_paid_at` datetime(3) DEFAULT NULL COMMENT '首次支付时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_openid` (`openid`),
  KEY `idx_users_union_id` (`union_id`),
  KEY `idx_users_phone` (`phone`)
) ENGINE=InnoDB AUTO_INCREMENT=65 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='C端用户表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-08-23 11:22:46
